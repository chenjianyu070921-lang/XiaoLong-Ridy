# CreateOrder 创建订单

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`CreateOrder(CreateOrderRequest) returns (CreateOrderResponse)`
- **功能**：乘客叫车后创建一笔订单，初始状态为「待接单（WAIT_ACCEPT）」。服务端会优先调用 pricesvc 复核计价（失败时降级为入参预估价格），随后发布 `order.created` 事件由派单消费者触发派单；事件总线不可用时回退为同步直派（DispatchClient.DispatchOrder）。
- **调用方**：乘客端（C）

## 请求参数（CreateOrderRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | int64 | 是 | 乘客用户 ID，必须 > 0 |
| car_type | int32 | 是 | 车型：1 特惠快车 / 2 快车 / 3 拼车（范围 1~3） |
| from_address | string | 是 | 上车点地址，不可为空 |
| from_longitude | double | 是 | 上车点经度，范围 [-180, 180] |
| from_latitude | double | 是 | 上车点纬度，范围 [-90, 90] |
| to_address | string | 是 | 下车点地址，不可为空 |
| to_longitude | double | 是 | 下车点经度，范围 [-180, 180] |
| to_latitude | double | 是 | 下车点纬度，范围 [-90, 90] |
| estimated_distance_m | int64 | 是 | 预估距离（米），>= 0 |
| estimated_duration_s | int64 | 是 | 预估时长（秒），>= 0 |
| estimated_price_cents | int64 | 是 | 预估价格（分），>= 0 |
| city_code | string | 否 | 城市编码（行政区划码），空时按默认城市（"110000"）计价 |

> 入参校验不通过返回 `invalid order params` 错误。

## 响应参数（CreateOrderResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 新订单 ID |
| order_no | string | 订单编号（业务唯一号） |
| estimated_price_cents | int64 | 实际生效的预估价格（分），为 pricesvc 复核结果或入参降级值 |
| status | OrderStatus | 固定为 `ORDER_STATUS_WAIT_ACCEPT`（待接单） |
| created_at | int64 | 创建时间，Unix 秒 |

## 状态流转
`初始(0)` → `待接单(WAIT_ACCEPT)`，并写入一条 `创建订单` 状态日志（operator=user）。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | user_id<=0、car_type 不在 1~3、地址为空、经纬度越界、距离/时长/价格 < 0 |

## 调用示例
**请求**
```json
{
  "user_id": 1001,
  "car_type": 2,
  "from_address": "北京市朝阳区望京 SOHO",
  "from_longitude": 116.481,
  "from_latitude": 39.996,
  "to_address": "北京市海淀区中关村",
  "to_longitude": 116.316,
  "to_latitude": 39.983,
  "estimated_distance_m": 12000,
  "estimated_duration_s": 1800,
  "estimated_price_cents": 3800,
  "city_code": "110000"
}
```
**响应**
```json
{
  "order_id": 50001,
  "order_no": "XL202608210001",
  "estimated_price_cents": 3750,
  "status": "ORDER_STATUS_WAIT_ACCEPT",
  "created_at": 1789600000
}
```
