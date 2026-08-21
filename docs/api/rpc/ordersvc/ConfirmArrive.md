# ConfirmArrive 司机到达上车点

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`ConfirmArrive(ConfirmArriveRequest) returns (ConfirmArriveResponse)`
- **功能**：司机确认已到达上车点。**订单状态保持不变（仍为「已接单」）**，仅记录一条状态日志用于轨迹展示与后续查询。
- **调用方**：司机端（B）

## 请求参数（ConfirmArriveRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |
| driver_id | int64 | 是 | 司机 ID，必须 > 0 |

## 响应参数（ConfirmArriveResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| status | OrderStatus | 仍为 `ORDER_STATUS_ACCEPTED`（已接单，状态不变） |

## 校验
- 订单当前状态必须为「已接单（ACCEPTED）」，否则返回 `order status not allowed`。
- 司机 ID 必须与订单 driver_id 一致，否则返回 `driver not matched`。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | order_id<=0 或 driver_id<=0 |
| order status not allowed | 订单当前状态不是已接单 |
| driver not matched | 司机 ID 与订单所属司机不一致 |

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
