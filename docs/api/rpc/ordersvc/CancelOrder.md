# CancelOrder 取消订单

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`CancelOrder(CancelOrderRequest) returns (CancelOrderResponse)`
- **功能**：乘客 / 司机 / 系统 / 管理员取消订单，订单进入「已取消（CANCELLED）」。取消成功后同步失效该订单的待派单记录（syncCancelDispatch），避免被重派任务重复处理。
- **调用方**：乘客端（C）、司机端（B）、系统任务、管理后台（D 模块后台）

## 请求参数（CancelOrderRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |
| operator_type | string | 是 | 取消方：user / driver / system / admin |
| operator_id | int64 | 条件必填 | 操作人 ID；operator_type != system 时必须 > 0 |
| reason | string | 是 | 取消原因，不可为空 |

## 响应参数（CancelOrderResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| status | OrderStatus | 固定为 `ORDER_STATUS_CANCELLED`（已取消） |

## 状态流转
仅允许从 `待接单(WAIT_ACCEPT)` 或 `已接单(ACCEPTED)` → `已取消(CANCELLED)`。

## 权限与校验
- operator_type 必须为 user / driver / system / admin 之一。
- 取消方权限（`canCancelByOperator`）：
  - **user**：仅可取消自己（user_id 匹配）且状态为「待接单」的订单。
  - **driver**：仅可取消状态为「已接单」且 driver_id 匹配自己的订单。
  - **system / admin**：可取消任意允许取消的订单。
- 使用订单级 Redis 分布式锁，避免与接单 / 超时取消并发竞态。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | order_id<=0、operator_type 非法、非 system 时 operator_id<=0 |
| cancel reason required | reason 为空 |
| order status not cancelable | 当前状态不允许取消（非待接单/已接单） |
| operator not allowed to cancel this order | 操作人无权限取消该订单 |

## 调用示例
**请求**
```json
{
  "order_id": 50001,
  "operator_type": "user",
  "operator_id": 1001,
  "reason": "临时有事不去了"
}
```
**响应**
```json
{
  "order_id": 50001,
  "status": "ORDER_STATUS_CANCELLED"
}
```
