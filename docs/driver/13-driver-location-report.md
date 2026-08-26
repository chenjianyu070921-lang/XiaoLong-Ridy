# 司机位置上报接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers/location/report` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.ReportLocation` |

## 2. 请求参数

同 [司机上线接口](10-driver-online.md)。

## 3. 请求示例

```json
{"deviceId":"dev-001","longitude":116.40,"latitude":39.90}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driverId` | int64 | 当前司机 ID |
| `onlineStatus` | int | 当前在线状态 |
| `kicked` | bool | 是否被其他设备顶替 |
| `reportTime` | int64 | 上报处理时间，Unix 秒 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-LOC-E01 | 未登录 | HTTP 401 |
| DRIVER-LOC-E02 | deviceId 为空 | HTTP 400 |
| DRIVER-LOC-E03 | 经纬度越界 | HTTP 400 |

## 6. 处理链路

`api/driver -> LocationLogic.ReportLocation -> driversvc.ReportLocation -> OnlineStore/driver_location`。
