# SetDriverServiceStatus RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `SetDriverServiceStatus(SetDriverServiceStatusRequest) returns (SetDriverServiceStatusResponse)` |
| 对应 HTTP | 无直接路由，订单行程逻辑内部使用 |
| 当前状态 | 已实现 |
| 业务逻辑 | `SetDriverServiceStatusLogic.SetDriverServiceStatus` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `driver_id` | int64 | 是 | 司机 ID |
| `online_status` | int32 | 是 | 0离线、1在线、2行程中 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `driver_id` | int64 | 司机 ID |
| `online_status` | int32 | 更新后的服务状态 |
| `updated_at` | int64 | Unix 秒 |

## 4. API 对齐

`POST /orders/start-trip` 成功后调用本 RPC 设置 `online_status=2`；`POST /orders/finish-trip` 成功后调用本 RPC 设置 `online_status=1`。
