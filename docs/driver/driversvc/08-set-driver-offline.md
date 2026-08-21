# SetDriverOffline RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `SetDriverOffline(SetDriverOfflineRequest) returns (SetDriverOfflineResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers/offline` |
| 当前状态 | 已实现 |
| 业务逻辑 | `SetDriverOfflineLogic.SetDriverOffline` |

## 2. 请求/响应字段

字段同 [SetDriverOnline](07-set-driver-online.md)，成功后 `online_status=0`。

## 3. 处理链路

`api/driver.SetOffline -> driversvc.SetDriverOffline -> OnlineStore/driver_location`。HTTP 文档见 [../11-driver-offline.md](../11-driver-offline.md)。
