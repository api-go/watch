# Test Report

## 单元测试结果
执行时间: 2026-06-21

```
=== RUN   TestWatch
--- PASS: TestWatch (0.20s)
=== RUN   TestDebounce
--- PASS: TestDebounce (0.28s)
=== RUN   TestEasyStart
--- PASS: TestEasyStart (0.00s)
PASS
ok  	apigo.cc/go/watch	2.028s
```

## 性能测试结果 (Benchmark)
```
BenchmarkIsMatch-16                 	14138798	        84.94 ns/op
BenchmarkIsExcluded-16              	49184965	        23.74 ns/op
BenchmarkDebounce-16                	 3307248	       378.8 ns/op
BenchmarkKeyGeneration-16           	27976675	        44.53 ns/op
BenchmarkKeyGenerationWithFmt-16    	 6969997	       169.7 ns/op
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
