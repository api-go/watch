package watch

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"apigo.cc/go/file"
	"apigo.cc/go/timer"
	"github.com/fsnotify/fsnotify"
	"github.com/gobwas/glob"
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
	Paths        []string      // 监听的路径列表
	Types        []string      // 包含的文件类型，例如 [".go", ".js"]
	ExcludeTypes []string      // 排除的文件类型
	Excludes     []string      // 排除模式，支持 gitignore 语义 (e.g. "**/node_modules/**", "*.log")
	Events       []EventType   // 监听的事件类型，默认全部
	Debounce     time.Duration // 防抖时间，相同文件的频繁事件会被合并
}

// Watcher 监听器实例
type Watcher struct {
	config    *Config
	callback  func(*Event)
	fsWatcher *fsnotify.Watcher
	isRunning bool
	mu        sync.RWMutex
	stopChan  chan struct{}

	excludeGlobs []glob.Glob
	eventMap     map[EventType]struct{}
	typeMap      map[string]struct{}
	exTypeMap    map[string]struct{}

	debounceGroups sync.Map // map[string]*timer.Debouncer
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

	watcher := &Watcher{
		config:    &config,
		callback:  callback,
		fsWatcher: fsWatcher,
		isRunning: true,
		stopChan:  make(chan struct{}),
		eventMap:  make(map[EventType]struct{}),
		typeMap:   make(map[string]struct{}),
		exTypeMap: make(map[string]struct{}),
	}

	for _, et := range config.Events {
		watcher.eventMap[et] = struct{}{}
	}

	for _, t := range config.Types {
		if !strings.HasPrefix(t, ".") {
			t = "." + t
		}
		watcher.typeMap[t] = struct{}{}
	}

	for _, t := range config.ExcludeTypes {
		if !strings.HasPrefix(t, ".") {
			t = "." + t
		}
		watcher.exTypeMap[t] = struct{}{}
	}

	// 编译 Glob 模式
	for _, p := range config.Excludes {
		if g, err := glob.Compile(p, '/'); err == nil {
			watcher.excludeGlobs = append(watcher.excludeGlobs, g)
		}
	}

	// 初始添加路径
	for _, p := range config.Paths {
		if err := watcher.Add(p); err != nil {
			_ = fsWatcher.Close()
			return nil, err
		}
	}

	go watcher.run()

	return watcher, nil
}

// Stop 停止监听
func (watcher *Watcher) Stop() {
	watcher.mu.Lock()
	if !watcher.isRunning {
		watcher.mu.Unlock()
		return
	}
	watcher.isRunning = false
	watcher.mu.Unlock()

	_ = watcher.fsWatcher.Close()
	<-watcher.stopChan
}

// Add 动态添加监听路径（递归）
func (watcher *Watcher) Add(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if !file.Exists(absPath) {
		_ = os.MkdirAll(absPath, 0755)
	}

	return filepath.Walk(absPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if watcher.isExcluded(p, true) {
				return filepath.SkipDir
			}
			return watcher.fsWatcher.Add(p)
		}
		return nil
	})
}

func (watcher *Watcher) Remove(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	return watcher.fsWatcher.Remove(absPath)
}

func (watcher *Watcher) run() {
	defer close(watcher.stopChan)

	for {
		select {
		case fsEvent, ok := <-watcher.fsWatcher.Events:
			if !ok {
				return
			}
			watcher.handleFsEvent(fsEvent)

		case _, ok := <-watcher.fsWatcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (watcher *Watcher) handleFsEvent(fsEvent fsnotify.Event) {
	name, err := filepath.Abs(fsEvent.Name)
	if err != nil {
		name = fsEvent.Name
	}
	var eventType EventType
	isDir := false

	if info, err := os.Stat(name); err == nil {
		isDir = info.IsDir()
	}

	switch {
	case fsEvent.Has(fsnotify.Create):
		eventType = Create
		if isDir {
			_ = watcher.Add(name)
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

	if !watcher.isMatch(name, eventType, isDir) {
		return
	}

	event := &Event{
		Path:  name,
		Type:  eventType,
		IsDir: isDir,
	}

	if watcher.config.Debounce > 0 {
		watcher.debounce(event)
	} else {
		watcher.callback(event)
	}
}

func (watcher *Watcher) debounce(event *Event) {
	// 优化 key 生成，避免过多的字符串拼接，但在此时直接拼接是比较简单且易读的
	key := event.Path + ":" + string(event.Type)
	d, _ := watcher.debounceGroups.LoadOrStore(key, timer.NewDebouncer(watcher.config.Debounce, func() {
		watcher.callback(event)
		watcher.debounceGroups.Delete(key)
	}))
	d.(*timer.Debouncer).Trigger()
}

func (watcher *Watcher) isMatch(path string, et EventType, isDir bool) bool {
	// 1. 事件类型过滤
	if _, ok := watcher.eventMap[et]; !ok {
		return false
	}

	// 2. 排除过滤 (通用)
	if watcher.isExcluded(path, isDir) {
		return false
	}

	// 3. 类型过滤 (仅针对文件)
	if !isDir {
		ext := filepath.Ext(path)
		if len(watcher.exTypeMap) > 0 {
			if _, ok := watcher.exTypeMap[ext]; ok {
				return false
			}
		}

		if len(watcher.typeMap) > 0 {
			if _, ok := watcher.typeMap[ext]; !ok {
				return false
			}
		}
	}

	return true
}

func (watcher *Watcher) isExcluded(path string, isDir bool) bool {
	if len(watcher.excludeGlobs) == 0 {
		return false
	}

	// 统一使用 / 作为路径分隔符进行匹配，对齐 gitignore 习惯
	p := filepath.ToSlash(path)
	base := filepath.Base(path)

	for _, g := range watcher.excludeGlobs {
		if g.Match(p) || g.Match(base) {
			return true
		}
		// 如果是目录，尝试加上 / 匹配 (支持 **/node_modules/** 这种语法匹配目录本身)
		if isDir && g.Match(p+"/") {
			return true
		}
	}
	return false
}

func (watcher *Watcher) WatchList() []string {
	return watcher.fsWatcher.WatchList()
}

func EasyStart(path string, callback func(string, EventType)) (*Watcher, error) {
	return Start(Config{
		Paths: []string{path},
	}, func(e *Event) {
		callback(e.Path, e.Type)
	})
}
