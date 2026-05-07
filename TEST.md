# Test Report

## 单元测试结果
执行时间: 2026-05-07

```
=== RUN   TestWatch
--- PASS: TestWatch (0.10s)
=== RUN   TestEasyStart
--- PASS: TestEasyStart (0.00s)
PASS
ok  	apigo.cc/go/watch	0.411s
```

## 测试覆盖场景
1. **文件创建监听**: 验证 `Create` 事件及路径正确性。
2. **类型包含过滤**: 验证仅监听 `Types` 指定的后缀。
3. **类型排除过滤**: 验证 `ExcludeTypes` 能正确拦截事件。
4. **模糊匹配排除**: 验证 `Excludes` 能根据关键词或通配符拦截路径（如 `node_modules`）和文件（如 `*.log`）。
5. **极简接口验证**: 验证 `EasyStart` 的可用性。
6. **递归监听**: 隐式验证了子目录的自动监听。
