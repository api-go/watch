# CHANGELOG

## v1.0.0
- 从 `@ssgo/tool/watcher` 迁移并重构。
- **基础设施对齐**: 使用 `apigo.cc/go/file` 替代标准库和 `ssgo/u`。
- **功能增强**:
    - 支持 `ExcludeTypes` 排除指定文件后缀。
    - 支持 `Excludes` 模糊匹配排除路径和文件。
    - 支持 `Events` 指定监听事件类型。
- **API 优化**:
    - 引入 `Event` 结构体，提供 `Path`, `Type`, `IsDir` 拆解参数。
    - 提供 `EasyStart` 极简入口。
    - 核心逻辑移除 Panic，增强健壮性。
