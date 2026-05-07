# Test Report

## 单元测试结果
执行时间: 2026-05-07

```
=== RUN   TestWatch
--- PASS: TestWatch (0.20s)
=== RUN   TestDebounce
--- PASS: TestDebounce (0.25s)
=== RUN   TestEasyStart
--- PASS: TestEasyStart (0.00s)
PASS
ok  	apigo.cc/go/watch	0.748s
```

## 测试覆盖场景
1. **基础监听**: 验证文件创建事件。
2. **Gitignore 语义过滤**:
   - 验证 `*.log` 后缀匹配。
   - 验证 `**/node_modules/**` 目录及其内容的深度匹配与拦截。
3. **事件防抖 (Debounce)**:
   - 验证 100ms 内对同一文件的 5 次连续写入被合并为 1 次回调。
4. **路径归一化**: 验证绝对路径下的准确匹配。
5. **极简接口**: 验证 `EasyStart` 的可用性。
