# 司机心跳接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers/heartbeat` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.Heartbeat` |

## 2. 请求参数

同 [司机上线接口](10-driver-online.md)。

## 3. 请求示例

```json
{"deviceId":"dev-001","longitude":116.40,"latitude":39.90}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `onlineStatus` | int | 当前在线状态 |
| `kicked` | bool | 是否被其他设备顶替 |
| `serverTime` | int64 | 服务端时间，Unix 秒 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-HEARTBEAT-E01 | 未登录 | HTTP 401 |
| DRIVER-HEARTBEAT-E02 | deviceId 为空 | HTTP 400 |
| DRIVER-HEARTBEAT-E03 | 经纬度越界 | HTTP 400 |

## 6. 处理链路

`api/driver -> HeartbeatLogic.Heartbeat -> driversvc.Heartbeat -> OnlineStore.Heartbeat`。`kicked=true` 时客户端应重新登录。
