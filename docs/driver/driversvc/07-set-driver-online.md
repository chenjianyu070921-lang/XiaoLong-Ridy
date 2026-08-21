# SetDriverOnline RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `SetDriverOnline(SetDriverOnlineRequest) returns (SetDriverOnlineResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers/online` |
| 当前状态 | 已实现 |
| 业务逻辑 | `SetDriverOnlineLogic.SetDriverOnline` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `driver_id` | int64 | 是 | 司机 ID，由 HTTP JWT 注入 |
| `device_id` | string | 是 | 设备 ID |
| `longitude` | double | 是 | 经度 |
| `latitude` | double | 是 | 纬度 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driver_id` | int64 | 司机 ID |
| `online_status` | int32 | `1` 在线 |
| `kicked` | bool | 是否被其他设备顶替 |

## 4. 处理链路

`api/driver.SetOnline -> driversvc.SetDriverOnline -> OnlineStore/driver_location`。HTTP 文档见 [../10-driver-online.md](../10-driver-online.md)。
