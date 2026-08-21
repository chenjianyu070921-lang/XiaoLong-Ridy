# TimeoutCancel 超时自动取消

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`TimeoutCancel(TimeoutCancelRequest) returns (TimeoutCancelResponse)`
- **功能**：由系统任务 / 消费者调用，取消超时未被接单的订单。订单进入「已取消（CANCELLED）」，并同步失效其待派单记录。
- **调用方**：系统定时任务 / 超时扫描消费者

## 请求参数（TimeoutCancelRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |
| reason | string | 否 | 取消原因；为空时默认填「超时未接单」 |

## 响应参数（TimeoutCancelResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| status | OrderStatus | 固定为 `ORDER_STATUS_CANCELLED`（已取消） |

## 状态流转
`待接单(WAIT_ACCEPT)` → `已取消(CANCELLED)`，写入状态日志（operator=system，remark=reason）。

## 校验
- 仅当订单当前状态为「待接单」且 **driver_id == 0（尚未被接单）** 时才允许取消，否则返回 `order status not cancelable`。
- 使用订单级 Redis 分布式锁，避免与接单 / 手动取消并发竞态。
- 取消成功后调用 `syncCancelDispatch` 失效待派单记录（失败不阻断主流程）。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | order_id<=0 |
| order status not cancelable | 订单已非待接单，或已被接单（driver_id != 0） |

## 调用示例
**请求**
```json
{
  "order_id": 50001,
  "reason": "超时未接单"
}
```
**响应**
```json
{
  "order_id": 50001,
  "status": "ORDER_STATUS_CANCELLED"
}
```
