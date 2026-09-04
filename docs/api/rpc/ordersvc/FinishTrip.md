# FinishTrip 结束行程

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`FinishTrip(FinishTripRequest) returns (FinishTripResponse)`
- **功能**：司机结束行程，订单从「行程中（ON_TRIP）」变为「待支付（WAIT_PAY）」。服务端以计价结果为准：优先按实际里程/时长调 pricesvc 重算应收（失败降级为下单预估快照），并对司机上报金额做偏差校验；随后调用 paysvc 生成支付单（失败不阻断订单状态）。
- **调用方**：司机端（B）

## 请求参数（FinishTripRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |
| driver_id | int64 | 是 | 司机 ID，必须 > 0 |
| actual_distance_m | int64 | 是 | 实际行驶距离（米），>= 0 |
| actual_duration_s | int64 | 是 | 实际行驶时长（秒），>= 0 |
| actual_price_cents | int64 | 是 | 司机上报实付金额（分），>= 0 |

## 响应参数（FinishTripResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| status | OrderStatus | 固定为 `ORDER_STATUS_WAIT_PAY`（待支付） |
| payable_amount_cents | int64 | 实际生效应付金额（分），即服务端权威金额（或司机上报兜底值） |

## 状态流转
`行程中(ON_TRIP)` → `待支付(WAIT_PAY)`，写入 `行程结束` 状态日志（operator=driver）。

## 金额校验规则
- **服务端权威金额**：优先 pricesvc 按实际里程/时长重算；不可用时用下单预估快照。两者均无合法金额时，采用司机上报值兜底。
- **偏差容忍**：司机上报金额仅做校验，不覆盖权威金额。
  - 权威金额来自服务端实时计价：容忍偏差 **10%**。
  - 权威金额来自降级快照：容忍偏差 **50%**。
  - 超出容忍范围返回 `price mismatch`。

## 其他副作用
- 若实际费用来自服务端实时计价，调用 `PriceClient.SaveActualOrderPrice` 落库实际计价明细（失败仅记日志）。
- 调用 `PayClient.CreatePayment` 生成支付单（默认微信渠道 channel=1）；失败仅记日志，不阻断。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | order_id<=0、driver_id<=0、距离/时长/价格 < 0 |
| order status not allowed | 订单当前状态不是行程中，或条件更新失败 |
| driver not matched | 司机 ID 与订单所属司机不一致 |
| price mismatch | 司机上报金额与权威金额偏差超容忍范围 |

## 调用示例
**请求**
```json
{
  "order_id": 50001,
  "driver_id": 2001,
  "actual_distance_m": 12500,
  "actual_duration_s": 1850,
  "actual_price_cents": 3800
}
```
**响应**
```json
{
  "order_id": 50001,
  "status": "ORDER_STATUS_WAIT_PAY",
  "payable_amount_cents": 3760
}
```
