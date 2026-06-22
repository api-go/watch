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

### CLI 命令行工具

```bash
go install apigo.cc/go/watch/watch@latest
```

```bash
# Go 开发，变更时重新运行测试
watch -go go test ./...

# Web 前端，变更时自动构建
watch -web npm run build

# Python 开发，变更时重启脚本
watch -py python main.py --dev

# 自定义：监听 src 目录的 Go 文件，变更时运行
watch -path ./src -type go go run .

# 仅打印变更事件（不执行命令）
watch -go
```

**参数说明**：

| 参数 | 说明 |
|---|---|
| `-path` | 监听路径（默认 `.`，递归） |
| `-type` | 包含文件类型，逗号分隔 |
| `-exclude-type` | 排除文件类型，逗号分隔 |
| `-exclude-path` | 排除路径模式，可重复（gitignore 语义） |
| `-go` | Go 开发预设（go \| vendor/\*\*, \*\*/node_modules/\*\*） |
| `-web` | Web 前端预设（html,css,js,ts,vue,jsx,tsx \| node_modules/\*\*, dist/\*\*） |
| `-js` | Node.js 预设（js,ts,json \| node_modules/\*\*） |
| `-py` | Python 预设（py,toml,yaml,yml \| \_\_pycache\_\_/\*\*, .venv/\*\*） |

预设值会与用户显式指定的参数合并追加，不会互相替换。

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
