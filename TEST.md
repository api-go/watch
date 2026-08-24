# Test Report

## v1.5.6 验证
- 验证文件事件、`rs` 与 `SIGUSR1` 重启请求串行执行。
- 验证旧进程退出并释放资源后才启动新进程。
- 全量测试、竞态测试和基准测试均通过。

## 单元测试结果
执行时间: 2026-06-27

```
=== RUN   TestWatch
--- PASS: TestWatch (0.21s)
=== RUN   TestDirEventWithTypeFilter
--- PASS: TestDirEventWithTypeFilter (0.15s)
=== RUN   TestDebounce
--- PASS: TestDebounce (0.26s)
=== RUN   TestEasyStart
--- PASS: TestEasyStart (0.00s)
PASS
ok  	apigo.cc/go/watch	1.018s
```

## 性能测试结果 (Benchmark)
```
BenchmarkIsMatch-16                 	26264806	        45.35 ns/op
BenchmarkIsExcluded-16              	42229480	        28.65 ns/op
BenchmarkDebounce-16                	 8131125	       148.0 ns/op
BenchmarkKeyGeneration-16           	47584114	        25.56 ns/op
BenchmarkKeyGenerationWithFmt-16    	13956841	        88.51 ns/op
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
6. **性能优化验证**: 验证在高频匹配场景下的 $O(1)$ 查找能力。
7. **类型白名单与目录过滤**: 验证开启 Types 白名单时，正常过滤目录变更事件，不误触发回调 (`TestDirEventWithTypeFilter`)。
8. **重启并发控制**: 连续发送多个 `SIGUSR1` 时重启串行执行，旧监听端口释放后才启动新进程。
