# CHANGELOG

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
