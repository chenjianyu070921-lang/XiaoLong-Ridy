# 司机端位置上报单元压测文档

## 范围

本压测覆盖司机端 P0 高频链路：

`LocationLogic.ReportLocation -> DriverClient.ReportLocation`

测试目标是验证 API 层在不启动 HTTP 服务、Redis、数据库和真实 RPC 的情况下，完成参数校验、请求组装和下游调用的稳定性与基准开销。真实网络、数据库写入、Redis 在线状态刷新不在本单元压测范围内，需要另行做联调或端到端压测。

## 测试文件

测试代码位于：

`api/driver/internal/logic/location_logic_test.go`

包含三类用例：

| 用例 | 命令默认执行 | 目的 |
| --- | --- | --- |
| `TestReportLocationForwardsCurrentDriverAndLocation` | 是 | 验证当前司机 ID、deviceId、经纬度会正确转发到 driversvc |
| `TestReportLocationParallelStress` | 是 | 用 64 个 goroutine 并发调用 6400 次，验证并发调用无错误且调用次数完整 |
| `BenchmarkReportLocation*` | 否 | 用 `go test -bench` 输出串行/并发基准性能和内存分配 |

## 执行命令

只跑位置上报单元压测：

```powershell
go test ./api/driver/internal/logic -run TestReportLocationParallelStress -count=1
```

跑 benchmark 压测：

```powershell
go test ./api/driver/internal/logic -run '^$' -bench 'BenchmarkReportLocation' -benchmem -benchtime=10s '-cpu=1,4,8'
```

带数据竞争检查：

```powershell
go test ./api/driver/internal/logic -run TestReportLocationParallelStress -race -count=1
```

全量回归：

```powershell
go test ./...
```

## 结果解读

`TestReportLocationParallelStress` 通过标准：

- 测试退出码为 `0`
- 没有 `ReportLocation() parallel stress error`
- 调用次数等于 `6400`

benchmark 重点看：

| 指标 | 含义 |
| --- | --- |
| `ns/op` | 单次逻辑调用耗时，越低越好 |
| `B/op` | 单次调用分配字节数，越低越好 |
| `allocs/op` | 单次调用分配次数，越低越好 |

该 benchmark 使用本地 fake client，不代表真实 RPC QPS。它用于发现 API 层逻辑退化，例如新增了不必要的反射、JSON 编解码、锁竞争或大对象分配。

## 建议基线

每次改动以下文件后至少执行一次本压测：

- `api/driver/internal/logic/location_logic.go`
- `api/driver/internal/svc/service_context.go`
- `rpc/driversvc/proto/driversvc.proto`
- `rpc/driversvc/proto/driversvc.pb.go`
- `rpc/driversvc/proto/driversvc_grpc.pb.go`

如果 benchmark 的 `allocs/op` 或 `B/op` 明显上升，需要检查请求构造、字符串处理和下游 client 封装是否引入了额外分配。
