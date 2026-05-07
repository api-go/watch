package watch

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"apigo.cc/go/file"
)

// EventType 定义事件类型
type EventType string

const (
	Create EventType = "create" // 创建文件或目录
	Change EventType = "change" // 修改文件
	Remove EventType = "remove" // 删除文件或目录
	Rename EventType = "rename" // 重命名文件或目录
)

// Event 包含事件的详细信息
type Event struct {
	Path  string    // 触发事件的路径
	Type  EventType // 事件类型
	IsDir bool      // 是否为目录
}

// Config 监听配置
type Config struct {
	Paths        []string    // 监听的路径列表
	Types        []string    // 包含的文件类型，例如 [".go", ".js"]
	ExcludeTypes []string    // 排除的文件类型
	Excludes     []string    // 排除的路径或文件名，支持模糊匹配
	Events       []EventType // 监听的事件类型，默认全部
}

// Watcher 监听器实例
type Watcher struct {
	config    *Config
	callback  func(*Event)
	fsWatcher *fsnotify.Watcher
	isRunning bool
	mu        sync.RWMutex
	stopChan  chan struct{}
}

// Start 开始监听
func Start(config Config, callback func(*Event)) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// 规范化配置
	if len(config.Events) == 0 {
		config.Events = []EventType{Create, Change, Remove, Rename}
	}
	for i, t := range config.Types {
		if !strings.HasPrefix(t, ".") {
			config.Types[i] = "." + t
		}
	}
	for i, t := range config.ExcludeTypes {
		if !strings.HasPrefix(t, ".") {
			config.ExcludeTypes[i] = "." + t
		}
	}

	w := &Watcher{
		config:    &config,
		callback:  callback,
		fsWatcher: fsWatcher,
		isRunning: true,
		stopChan:  make(chan struct{}),
	}

	// 初始添加路径
	for _, p := range config.Paths {
		if err := w.Add(p); err != nil {
			_ = fsWatcher.Close()
			return nil, err
		}
	}

	go w.run()

	return w, nil
}

// Stop 停止监听
func (w *Watcher) Stop() {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return
	}
	w.isRunning = false
	w.mu.Unlock()

	_ = w.fsWatcher.Close()
	<-w.stopChan
}

// Add 动态添加监听路径（递归）
func (w *Watcher) Add(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// 如果路径不存在，尝试创建（保持与原逻辑一致）
	if !file.Exists(absPath) {
		_ = os.MkdirAll(absPath, 0755)
	}

	return filepath.Walk(absPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误继续
		}

		if info.IsDir() {
			if w.isMatchExclude(p) {
				return filepath.SkipDir
			}
			return w.fsWatcher.Add(p)
		}
		return nil
	})
}

// Remove 动态移除监听路径
func (w *Watcher) Remove(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return w.fsWatcher.Remove(absPath)
}

func (w *Watcher) run() {
	defer close(w.stopChan)

	for {
		select {
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}

			w.handleFsEvent(event)

		case _, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (w *Watcher) handleFsEvent(fsEvent fsnotify.Event) {
	name := fsEvent.Name
	var eventType EventType
	isDir := false

	// 获取文件信息判断是否为目录
	if info, err := os.Stat(name); err == nil {
		isDir = info.IsDir()
	}

	// 映射事件类型
	switch {
	case fsEvent.Has(fsnotify.Create):
		eventType = Create
		if isDir {
			_ = w.Add(name) // 自动递归监听新目录
		}
	case fsEvent.Has(fsnotify.Write):
		eventType = Change
	case fsEvent.Has(fsnotify.Remove):
		eventType = Remove
	case fsEvent.Has(fsnotify.Rename):
		eventType = Rename
	default:
		return
	}

	// 过滤逻辑
	if !w.isMatch(name, eventType, isDir) {
		return
	}

	w.callback(&Event{
		Path:  name,
		Type:  eventType,
		IsDir: isDir,
	})
}

func (w *Watcher) isMatch(path string, et EventType, isDir bool) bool {
	// 1. 事件类型过滤
	eventMatch := false
	for _, configEt := range w.config.Events {
		if et == configEt {
			eventMatch = true
			break
		}
	}
	if !eventMatch {
		return false
	}

	// 2. 排除过滤 (通用)
	if w.isMatchExclude(path) {
		return false
	}

	// 3. 类型过滤 (仅针对文件)
	if !isDir {
		ext := filepath.Ext(path)
		
		// 排除类型
		if len(w.config.ExcludeTypes) > 0 {
			for _, t := range w.config.ExcludeTypes {
				if ext == t {
					return false
				}
			}
		}

		// 包含类型
		if len(w.config.Types) > 0 {
			typeMatch := false
			for _, t := range w.config.Types {
				if ext == t {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				return false
			}
		}
	}

	return true
}

func (w *Watcher) isMatchExclude(path string) bool {
	if len(w.config.Excludes) == 0 {
		return false
	}

	for _, pattern := range w.config.Excludes {
		// 简单模糊匹配：如果包含 * 则尝试 Match，否则使用 Contains
		if strings.Contains(pattern, "*") {
			if matched, _ := filepath.Match(pattern, path); matched {
				return true
			}
			// 额外支持在路径中包含 pattern 的情况
			if strings.Contains(path, strings.ReplaceAll(pattern, "*", "")) {
				return true
			}
		} else {
			if strings.Contains(path, pattern) {
				return true
			}
		}
	}
	return false
}

// WatchList 返回当前监听的所有路径
func (w *Watcher) WatchList() []string {
	return w.fsWatcher.WatchList()
}

// EasyStart 极简启动方式
func EasyStart(path string, callback func(string, EventType)) (*Watcher, error) {
	return Start(Config{
		Paths: []string{path},
	}, func(e *Event) {
		callback(e.Path, e.Type)
	})
}
