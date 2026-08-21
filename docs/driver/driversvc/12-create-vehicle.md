# CreateVehicle RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `CreateVehicle(CreateVehicleRequest) returns (CreateVehicleResponse)` |
| 对应 HTTP | 无司机端 HTTP 路由 |
| 当前状态 | 已实现 |
| 业务逻辑 | `CreateVehicleLogic.CreateVehicle` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `driver_id` | int64 | 是 | 司机 ID |
| `plate_no` | string | 是 | 车牌号 |
| `brand` | string | 否 | 品牌 |
| `model` | string | 否 | 车型 |
| `color` | string | 否 | 颜色 |
| `vehicle_type` | int32 | 是 | 车辆类型 |
| `registration_date` | optional int64 | 否 | 注册时间，Unix 秒 |
| `insurance_no` | string | 否 | 保险号 |
| `insurance_expire_at` | optional int64 | 否 | 保险到期时间，Unix 秒 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 车辆 ID |
| `status` | VehicleStatus | 初始为 `VEHICLE_STATUS_PENDING` |
| `created_at` | int64 | Unix 秒 |

## 4. API 对齐

司机端当前没有车辆 CRUD HTTP 路由；资质上传接口只传 `vehicleId` 并调用 `UploadCertification`。
