# ReportLocation RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `ReportLocation(ReportLocationRequest) returns (ReportLocationResponse)` |
| 对应 HTTP | `POST /api/driver/v1/drivers/location/report` |
| 当前状态 | 已实现 |
| 业务逻辑 | `ReportLocationLogic.ReportLocation` |

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
| `online_status` | int32 | 当前在线状态 |
| `kicked` | bool | 是否被其他设备顶替 |
| `report_time` | int64 | Unix 秒 |

## 4. 处理链路

`api/driver.ReportLocation -> driversvc.ReportLocation -> OnlineStore/driver_location`。HTTP 文档见 [../13-driver-location-report.md](../13-driver-location-report.md)。
