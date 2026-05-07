# Test Report

## 单元测试结果
执行时间: 2026-05-07

```
=== RUN   TestWatch
--- PASS: TestWatch (0.21s)
=== RUN   TestDebounce
--- PASS: TestDebounce (0.26s)
=== RUN   TestEasyStart
--- PASS: TestEasyStart (0.00s)
PASS
ok  	apigo.cc/go/watch	0.849s
```

## 性能测试结果 (Benchmark)
```
BenchmarkIsMatch-16                 	15307893	        74.12 ns/op	       0 B/op	       0 allocs/op
BenchmarkIsExcluded-16              	45664744	        26.09 ns/op	       0 B/op	       0 allocs/op
BenchmarkDebounce-16                	 3345813	       369.3 ns/op	     224 B/op	       5 allocs/op
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
