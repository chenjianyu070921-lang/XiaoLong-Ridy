# UpdateVehicle RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `UpdateVehicle(UpdateVehicleRequest) returns (UpdateVehicleResponse)` |
| 对应 HTTP | 无司机端 HTTP 路由 |
| 当前状态 | 已实现 |
| 业务逻辑 | `UpdateVehicleLogic.UpdateVehicle` |

## 2. 请求字段

`id` 必填，其余车辆字段为 optional；字段含义同 [CreateVehicle](12-create-vehicle.md)。

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 车辆 ID |
| `status` | VehicleStatus | 更新后状态 |
| `updated_at` | int64 | Unix 秒 |

## 4. API 对齐

当前无司机端 HTTP 路由，属于内部/后台车辆资料维护能力。
