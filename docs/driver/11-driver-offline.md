# 司机下线接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers/offline` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.SetDriverOffline` |

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
| `onlineStatus` | int | 在线状态，`0` 表示离线 |
| `kicked` | bool | 当前设备是否已被新设备顶替 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-OFFLINE-E01 | 未登录 | HTTP 401 |
| DRIVER-OFFLINE-E02 | 经纬度越界 | HTTP 400 |
| DRIVER-OFFLINE-E03 | 非当前绑定设备请求下线 | 返回 `kicked=true` |

## 6. 处理链路

`api/driver -> OfflineLogic.SetOffline -> driversvc.SetDriverOffline -> OnlineStore/driver_location`。
