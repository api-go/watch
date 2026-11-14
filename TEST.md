# Test Report

## v1.5.10 验证
- 验证 Watch 使用可配置进程组停机期限的编译与既有生命周期行为。
- `go test -v ./...` 与 `go test -bench=. ./...` 通过。

## v1.5.8 验证
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
ok  	github.com/api-go/watch	1.018s
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

## GitHub migration validation (pending release)

Module paths and internal dependency versions are migrated. Development verification uses GOWORK=off and explicit sibling replacements; remove these replacements before release. See migration validation logs for actual results. Existing unrelated working changes are preserved and require separate review.

API inventory verification: `GOWORK=off go doc` succeeded; output retained in the migration validation directory. Source exports were preserved by the path migration.

## GitHub release verification

GOWORK=off go test -v ./... and go test -bench=. ./... passed using published GitHub dependencies, without local replacements. go mod verify passed.

## 最终依赖图验证（2026-11-14，未发布）

显式本地 replace 覆盖本轮待发布依赖链；GOWORK=off go mod tidy、go test -v ./...、go test -bench=. -benchtime=1x ./...。基准按固定 1 次采样，避免排除计时的昂贵准备阶段反复校准。本轮结果：[('tidy', '0'), ('test', '0'), ('bench', '0')]。0 为通过，其他结果待定位。

固定一次采样的基准摘要（非稳定性能测量）：

```text
BenchmarkIsMatch-16                 	       1	    130875 ns/op
BenchmarkIsExcluded-16              	       1	      6833 ns/op
BenchmarkDebounce-16                	       1	    101084 ns/op
```

## v1.5.12 正式发布验证（2026-11-14）

GOWORK=off；移除全部 replace，使用已发布依赖。go mod tidy、go test -v ./...、go test -bench=. ./...、go mod verify 均通过。API 工具不可用，使用 go doc 与源码核对。

```text
BenchmarkIsMatch-16                 	25257688	        47.56 ns/op
BenchmarkIsExcluded-16              	40430925	        29.61 ns/op
BenchmarkDebounce-16                	 7644148	       161.3 ns/op
```
