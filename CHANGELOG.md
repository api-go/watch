# CHANGELOG

## v1.5.6 (2026-08-24)
- **重启可靠性**:
  - 文件事件、终端 `rs` 和 `SIGUSR1` 重启请求统一串行处理，防止并行启动多个子进程。
  - 等待旧进程完整退出后再启动新进程，并输出停止或启动失败信息。
- **依赖更新**: 对齐 `shell v1.5.6`、`file v1.5.6`、`cast v1.5.5`、`encoding v1.5.6`、`rand v1.5.4`、`safe v1.5.3`。

## v1.5.5 (2026-08-24)
- **CLI**: 增加 `watch.pid` 生命周期管理和 `SIGUSR1` 手动重启支持。

## v1.5.4 (2026-06-27)
- **修复 Bug**: 修复了配置 `Types`（文件类型白名单）时，目录本身的创建/删除等事件未能被正确过滤的问题。
- **CLI 体验大幅升级**:
  - **紧凑型 UI**: 重新设计了日志输出（Banner 与 Event），视觉更加紧凑专业，彻底分离了应用输出与 Watch 日志。
  - **智能追加参数**: 当使用 `--go` 等预设并提供自定义参数（如 `watch --go --dev`）时，CLI 会智能判断并将其追加到默认命令（如 `go run . -- --dev`）。
  - **.gitignore 支持**: 默认自动读取当前目录的 `.gitignore` 与 `.ignore` 文件并将其中的规则加入排除列表。
  - **.watch 任务文件**: 支持基于 `.watch` 文件的极简任务配置（键值对格式），可无参数直接执行默认任务，或通过 `watch --taskname` 执行特定任务。
  - **清屏支持**: 新增 `-c, --clear` 参数，在重启前自动清理终端。
  - **手动重启**: 支持在运行时键入 `rs` + 回车强行重启服务。
  - **恢复参数**: 重新暴露 `-E, --exclude-ext` 选项（原 `-exclude-type`）方便灵活排除特定后缀。
  - **参数标准化**: `-path` 改为 `-w/--watch`，`-type` 改为 `-e/--ext`，`-exclude-path` 改为 `-i/--ignore` 等。

## v1.5.3 (2026-06-22)
- **新增 CLI**: 新增 `watch/watch` 子包，提供 `watch` 命令行工具，支持 `go install apigo.cc/go/watch/watch@latest` 安装。
  - 支持 `-type`/`-exclude-type`/`-exclude-path` 等 watch.Start 全部配置能力。
  - 预设快捷方式：`-go`/`-web`/`-js`/`-py` 对应各开发场景。
  - 预设值与显式参数合并，不替换。
  - 无命令时打印变更事件到 stdout。
- **依赖更新**: 新增依赖 `go/shell` v1.5.4，用于子进程生命周期管理。

## v1.5.2 (2026-06-21)
- **依赖更新**: 升级依赖 `jsmod` 至 `v1.5.3`，`cast` 至 `v1.5.3`，`rand` 至 `v1.5.3`，`encoding` 至 `v1.5.4`，`safe` 至 `v1.5.2`，`file` 至 `v1.5.5`。

## v1.5.1 (2026-06-11)
- **版本对齐**: 基础设施全局对齐 v1.5.1。

## v1.5.0 (2026-05-10)
- **版本对齐**: 基础设施全局对齐 v1.5.0。

## v1.1.1
- **性能优化**: 将 `Types`、`ExcludeTypes` 和 `Events` 的过滤逻辑从 $O(n)$ 优化为 $O(1)$ 的 Map 查找。
- **基础设施对齐**: 更新依赖 `go/file` 至 v1.0.7, `go/encoding` 至 v1.0.6。
- **代码质量**: 统一变量命名规范（`watcher`, `event`），重命名内部方法 `isMatchExclude` 为 `isExcluded` 提升可读性。
- **测试增强**: 引入专门的 `bench_test.go`，补全匹配逻辑与防抖逻辑的性能基准测试。

## v1.1.0
- **核心升级**: 引入 `github.com/gobwas/glob` 引擎。
- **Gitignore 语义**: `Excludes` 现在支持 `**/node_modules/**` 等复杂路径模式，并能正确处理目录边界。
- **路径归一化**: 内部统一使用绝对路径与 `/` 分隔符进行匹配，确保跨平台表现一致。
- **事件防抖 (Debounce)**: 引入 `apigo.cc/go/timer` 的防抖能力，支持 `Debounce` 配置以合并碎片化事件。

## v1.0.0
- 从 `@ssgo/tool/watcher` 迁移并重构。
- **基础设施对齐**: 使用 `apigo.cc/go/file` 替代标准库。
- **API 优化**: 引入 `Event` 结构体，提供 `EasyStart` 极简入口。
  - Add `-r/--reload` PID-file control and ensure `-h/--help` takes precedence over `.watch` tasks.
