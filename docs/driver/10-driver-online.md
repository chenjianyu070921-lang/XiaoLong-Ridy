# 司机上线接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/drivers/online` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.SetDriverOnline` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `deviceId` | string | 是 | 设备标识 |
| `longitude` | float64 | 是 | 经度，范围 `[-180,180]` |
| `latitude` | float64 | 是 | 纬度，范围 `[-90,90]` |

## 3. 请求示例

```json
{"deviceId":"dev-001","longitude":116.40,"latitude":39.90}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driverId` | int64 | 当前司机 ID |
| `onlineStatus` | int | 在线状态，`1` 表示在线 |
| `kicked` | bool | 当前请求是否被其他设备顶替 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-ONLINE-E01 | 未登录 | HTTP 401 |
| DRIVER-ONLINE-E02 | 经纬度越界 | HTTP 400 |
| DRIVER-ONLINE-E03 | deviceId 为空 | HTTP 400 |

## 6. 处理链路

`api/driver -> OnlineLogic.SetOnline -> driversvc.SetDriverOnline -> OnlineStore/driver_location`。
