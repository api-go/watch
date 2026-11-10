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
go get github.com/api-go/watch
```

### CLI 命令行工具

```bash
go install github.com/api-go/watch/watch@latest
```

**快速体验**：
```bash
# Go 开发：无需提供命令，默认自动执行 go run . (并支持智能追加参数)
watch --go
watch --go --dev      # 实际执行: go run . -- --dev

# Go 测试：默认执行 go test -v -count=1 ./...
watch --gotest

# Web 前端：默认执行 npm run dev
watch --web

# 自定义执行命令（清屏并编译）
watch -c -e go -i "vendor/**" go build -o app
```

**任务文件 `.watch`**：
支持在项目根目录创建 `.watch` 文件（键值对格式），实现极简任务流。自动忽略文件中的空行和 `#` 注释。
```ini
dev = -c --go
test = --gotest
build = -c -e go -i vendor/** go build -o app
```
- 运行 `watch --dev`：直接加载 `dev` 的配置运行。
- 运行 `watch`（无参数）：自动执行第一项任务！

**自动排除 `.gitignore`**：
`watch` 默认会自动读取当前目录下的 `.gitignore` 和 `.ignore` 文件，并将其中的规则加入排除列表，**完全开箱即用**。

**交互式重启**：
当程序运行在后台时，在终端内敲击 `rs` 并回车，即可强行手动重启。
watch 启动时会在当前目录写入 `watch.pid`，也可通过信号手动重启：
```bash
kill -USR1 "$(cat watch.pid)"
watch -r                         # 读取当前目录 watch.pid
watch --reload path/to/watch.pid # 指定 PID 文件
```
所有文件事件、`rs` 和信号重启请求均串行执行；旧命令退出完成后才会启动新命令。

**参数说明**：

| 参数 | 说明 |
|---|---|
| `-w, --watch` | 监听路径（默认 `.`，递归） |
| `-e, --ext` | 包含文件类型，逗号分隔 (如 `go,js`) |
| `-E, --exclude-ext` | 排除文件类型，逗号分隔 |
| `-i, --ignore` | 排除路径模式，可重复（gitignore 语义） |
| `-c, --clear` | 重启前自动清屏 |
| `-r, --reload [pidfile]` | 向 PID 文件中的 watch 进程发送重启信号 |
| `--go` | Go 预设 (自动补齐 `go run .` 与忽略路径) |
| `--gotest` | Go 测试预设 (自动补齐 `go test -v -count=1 ./...`) |
| `--web` | Web 预设 (自动补齐 `npm run dev`) |
| `--js` | Node.js 预设 (自动补齐 `npm start`) |
| `--py` | Python 预设 (自动补齐 `python main.py`) |

预设值会与用户显式指定的参数合并追加，遇到显式声明的新命令时，会自动替换预设默认命令。

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
