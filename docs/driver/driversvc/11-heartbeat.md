# Heartbeat RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `Heartbeat(HeartbeatRequest) returns (HeartbeatResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers/heartbeat` |
| 当前状态 | 已实现 |
| 业务逻辑 | `HeartbeatLogic.Heartbeat` |

## 2. 请求字段

字段同 [ReportLocation](09-report-location.md)。

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `online_status` | int32 | 当前在线状态 |
| `kicked` | bool | 是否被其他设备顶替 |
| `server_time` | int64 | Unix 秒 |

## 4. 处理链路

`api/driver.Heartbeat -> driversvc.Heartbeat -> OnlineStore.Heartbeat`。HTTP 文档见 [../12-driver-heartbeat.md](../12-driver-heartbeat.md)。
