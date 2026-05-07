# watch

高性能、AI 友好的顶级文件监听库。支持递归监听、`.gitignore` 语义的路径过滤、事件防抖以及极简的 API。

## 特性

- **递归监听**: 自动监听子目录及其新创建的子目录。
- **顶级过滤能力**:
  - `Types`: 仅监听指定后缀的文件。
  - `ExcludeTypes`: 排除指定后缀的文件。
  - **`Excludes` (Gitignore 语义)**: 使用高性能 Glob 引擎（`github.com/gobwas/glob`），支持 `**/node_modules/**`、`*.log` 等复杂模式，自动处理目录边界。
- **事件防抖 (Debounce)**: 有效合并编辑器保存等产生的并发碎片事件，提升业务性能。
- **AI 友好**: 提供结构化的 `Event` 对象及拆解好的参数。
- **极简接口**: `EasyStart` 快速上手。
- **高性能**: 纯 Go 实现，无 CGO 依赖，基于 `fsnotify` 与 `go/timer` 高性能引擎。

## 安装

```bash
go get apigo.cc/go/watch
```

## API 指南

### Start
启动高级监听器。

```go
config := watch.Config{
    Paths:        []string{"./src"},
    Types:        []string{".go", ".js"},
    Excludes:     []string{"**/node_modules/**", "vendor/**", "*.log"},
    Debounce:     100 * time.Millisecond, // 开启防抖
}

w, err := watch.Start(config, func(e *watch.Event) {
    fmt.Printf("Path: %s, Type: %s, IsDir: %v\n", e.Path, e.Type, e.IsDir)
})
```

### EasyStart
启动极简监听器。

```go
w, err := watch.EasyStart("./data", func(path string, et watch.EventType) {
    fmt.Printf("File %s has event %s\n", path, et)
})
```

### Watcher 方法
- `Stop()`: 停止监听并释放资源（包括清理防抖定时器）。
- `Add(path string)`: 动态添加监听路径。
- `Remove(path string)`: 动态移除监听路径。
- `WatchList()`: 获取当前正在监听的所有路径。
