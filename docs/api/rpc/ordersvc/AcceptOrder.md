# AcceptOrder 司机接单

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`AcceptOrder(AcceptOrderRequest) returns (AcceptOrderResponse)`
- **功能**：司机接单，订单从「待接单（WAIT_ACCEPT）」变为「已接单（ACCEPTED）」。接单成功后闭环派单记录（MarkDispatchAccepted），并发布 `order.status.changed` 事件。
- **调用方**：司机端（B）

## 请求参数（AcceptOrderRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |
| driver_id | int64 | 是 | 司机 ID，必须 > 0 |

## 响应参数（AcceptOrderResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| status | OrderStatus | 固定为 `ORDER_STATUS_ACCEPTED`（已接单） |

## 状态流转
`待接单(WAIT_ACCEPT)` → `已接单(ACCEPTED)`，写入 `司机接单` 状态日志（operator=driver）。

## 并发与校验
- 加 Redis 分布式锁，保证同一订单只有一个司机接单成功。
- 仅当订单当前状态为「待接单」且状态流转合法时成功；否则返回 `order status not allowed`。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | order_id<=0 或 driver_id<=0 |
| order status not allowed | 订单当前状态不是待接单，或条件更新失败 |

## 调用示例
**请求**
```json
{
  "order_id": 50001,
  "driver_id": 2001
}
```
**响应**
```json
{
  "order_id": 50001,
  "status": "ORDER_STATUS_ACCEPTED"
}
```
