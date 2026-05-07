# watch

高性能、AI 友好的文件监听库。支持递归监听、多种过滤规则（类型包含/排除、模糊匹配排除）以及简洁的 API。

## 特性

- **递归监听**: 自动监听子目录及其新创建的子目录。
- **灵活过滤**:
  - `Types`: 仅监听指定后缀的文件。
  - `ExcludeTypes`: 排除指定后缀的文件。
  - `Excludes`: 支持模糊匹配（`strings.Contains` 或 `filepath.Match`）排除特定路径或文件。
- **AI 友好**: 提供结构化的 `Event` 对象及拆解好的参数。
- **极简接口**: `EasyStart` 快速上手。
- **基础设施对齐**: 使用 `apigo.cc/go/file` 等现代基础设施。

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
    ExcludeTypes: []string{".tmp"},
    Excludes:     []string{"node_modules", "vendor"},
    Events:       []watch.EventType{watch.Create, watch.Change},
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
- `Stop()`: 停止监听并释放资源。
- `Add(path string)`: 动态添加监听路径。
- `Remove(path string)`: 动态移除监听路径。
- `WatchList()`: 获取当前正在监听的所有路径。
