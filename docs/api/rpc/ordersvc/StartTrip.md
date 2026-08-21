# StartTrip 开始行程

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`StartTrip(StartTripRequest) returns (StartTripResponse)`
- **功能**：司机确认乘客上车后开始行程，订单从「已接单（ACCEPTED）」变为「行程中（ON_TRIP）」。
- **调用方**：司机端（B）

## 请求参数（StartTripRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |
| driver_id | int64 | 是 | 司机 ID，必须 > 0 |

## 响应参数（StartTripResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| status | OrderStatus | 固定为 `ORDER_STATUS_ON_TRIP`（行程中） |

## 状态流转
`已接单(ACCEPTED)` → `行程中(ON_TRIP)`，写入 `开始行程` 状态日志（operator=driver）。

## 校验
- 仅当订单当前状态为「已接单」且状态流转合法时成功。
- 司机 ID 必须与订单 driver_id 一致。

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
  "status": "ORDER_STATUS_ON_TRIP"
}
```
