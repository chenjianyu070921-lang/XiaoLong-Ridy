# 司机车辆查询接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `GET` |
| 请求路径 | `/api/driver/v1/vehicles/get` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.GetVehicle` |

该接口只返回当前 JWT 司机名下的车辆。即使传入其他司机的车辆 ID，也会被服务端拦截。

## 2. 查询参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 车辆 ID |

## 3. 请求示例

```bash
curl "http://127.0.0.1:8082/api/driver/v1/vehicles/get?id=77" \
  -H "Authorization: Bearer $TOKEN"
```

## 4. 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "vehicle": {
      "id": 77,
      "driverId": 25,
      "plateNo": "粤B12345",
      "brand": "BYD",
      "model": "Han",
      "color": "黑色",
      "vehicleType": 1,
      "registrationDate": 1700000000,
      "insuranceNo": "INS-001",
      "insuranceExpireAt": 1800000000,
      "status": "VEHICLE_STATUS_PENDING",
      "createdAt": 1700000000,
      "updatedAt": 1700000000
    }
  },
  "timestamp": 1700000000,
  "traceId": "trace_xxx"
}
```

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-VEHICLE-GET-E01 | 未登录 | HTTP 401 |
| DRIVER-VEHICLE-GET-E02 | `id` 为空或小于等于 0 | HTTP 400 |
| DRIVER-VEHICLE-GET-E03 | 查询其他司机车辆 | HTTP 403 |

## 6. 处理链路

`web/driver -> /driver/vehicles/get -> api/driver -> VehicleLogic.GetVehicle -> driversvc.GetVehicle`。
