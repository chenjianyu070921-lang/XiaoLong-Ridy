# DeleteVehicle RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `DeleteVehicle(DeleteVehicleRequest) returns (DeleteVehicleResponse)` |
| 对应 HTTP | 无司机端 HTTP 路由 |
| 当前状态 | 已实现 |
| 业务逻辑 | `DeleteVehicleLogic.DeleteVehicle` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | int64 | 是 | 车辆 ID |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | int64 | 车辆 ID |
| `success` | bool | 是否删除成功 |

## 4. API 对齐

当前无司机端 HTTP 路由。
