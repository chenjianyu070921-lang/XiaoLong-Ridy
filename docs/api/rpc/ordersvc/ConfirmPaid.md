# ConfirmPaid 支付成功确认

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`ConfirmPaid(ConfirmPaidRequest) returns (ConfirmPaidResponse)`
- **功能**：支付成功后确认订单完成，订单从「待支付（WAIT_PAY）」变为「已完成（COMPLETED）」。会向 paysvc 核验支付单真实状态，防止支付绕过。
- **调用方**：支付模块（E，order.paid 事件 / 回调）

## 请求参数（ConfirmPaidRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |
| payment_no | string | 是 | 支付单号，不可为空 |
| amount_cents | int64 | 是 | 支付金额（分），必须 > 0 |
| paid_at | int64 | 是 | 支付完成时间，Unix 秒，必须 > 0 |

## 响应参数（ConfirmPaidResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| status | OrderStatus | 固定为 `ORDER_STATUS_COMPLETED`（已完成） |

## 状态流转
`待支付(WAIT_PAY)` → `已完成(COMPLETED)`，写入 `支付成功` 状态日志（operator=system）。

## 校验与安全
- 订单当前状态必须为「待支付」，否则返回 `order status not allowed`。
- **未配置 PayClient 时直接拒绝**（防止绕过支付），返回 `pay client not configured`。
- **支付单核验**：向 paysvc 查询支付单，要求 payment_no、order_id、amount_cents 完全一致且 status == 已支付（2）；任一不符返回 `payment verification failed`。

## 其他副作用
- 核验通过后若配置了优惠券消费者（CouponConsumer），按订单核销优惠券。
- 发布 `order.status.changed` 事件（WAIT_PAY → COMPLETED）。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | order_id<=0、payment_no 为空、amount_cents<=0、paid_at<=0 |
| order status not allowed | 订单当前状态不是待支付，或完成更新失败 |
| pay client not configured | 支付客户端未配置 |
| payment verification failed | 支付单查询失败或支付单信息/状态不一致 |

## 调用示例
**请求**
```json
{
  "order_id": 50001,
  "payment_no": "PAY202608210001",
  "amount_cents": 3760,
  "paid_at": 1789603600
}
```
**响应**
```json
{
  "order_id": 50001,
  "status": "ORDER_STATUS_COMPLETED"
}
```
