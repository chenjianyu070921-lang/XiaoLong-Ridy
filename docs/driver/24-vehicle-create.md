# 司机车辆提交接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/vehicles` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `driversvc.CreateVehicle` |

司机 ID 统一从 JWT 中获取，前端不得传 `driverId`。

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `plateNo` | string | 是 | 车牌号 |
| `brand` | string | 是 | 品牌 |
| `model` | string | 是 | 型号 |
| `color` | string | 否 | 车辆颜色 |
| `vehicleType` | int32 | 是 | 车辆类型，示例：`1` 快车 |
| `registrationDate` | int64 | 否 | 注册日期，Unix 秒 |
| `insuranceNo` | string | 否 | 保险单号 |
| `insuranceExpireAt` | int64 | 否 | 保险到期时间，Unix 秒 |

## 3. 请求示例

```bash
curl -X POST http://127.0.0.1:8082/api/driver/v1/vehicles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"plateNo":"粤B12345","brand":"BYD","model":"Han","color":"黑色","vehicleType":1,"registrationDate":1700000000,"insuranceNo":"INS-001","insuranceExpireAt":1800000000}'
```

## 4. 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 77,
    "status": "VEHICLE_STATUS_PENDING",
    "createdAt": 1700000000
  },
  "timestamp": 1700000000,
  "traceId": "trace_xxx"
}
```

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-VEHICLE-CREATE-E01 | 未登录 | HTTP 401 |
| DRIVER-VEHICLE-CREATE-E02 | 车牌号为空或格式非法 | HTTP 400 |
| DRIVER-VEHICLE-CREATE-E03 | 品牌/型号为空 | HTTP 400 |
| DRIVER-VEHICLE-CREATE-E04 | `vehicleType<=0` | HTTP 400 |

## 6. 处理链路

`web/driver -> /driver/vehicles -> api/driver -> VehicleLogic.CreateVehicle -> driversvc.CreateVehicle`。
