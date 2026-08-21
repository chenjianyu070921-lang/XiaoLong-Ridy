# GetOrder 查询订单详情

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`GetOrder(GetOrderRequest) returns (GetOrderResponse)`
- **功能**：根据订单 ID 查询订单完整详情。
- **调用方**：乘客端（C）、司机端（B）、管理后台（D）

## 请求参数（GetOrderRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_id | int64 | 是 | 订单 ID，必须 > 0 |

## 响应参数（GetOrderResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| order_no | string | 订单编号 |
| user_id | int64 | 乘客用户 ID |
| driver_id | int64 | 司机 ID（未接单时为 0） |
| car_type | int32 | 车型：1 特惠快车 / 2 快车 / 3 拼车 |
| from_address | string | 上车点地址 |
| from_longitude | double | 上车点经度 |
| from_latitude | double | 上车点纬度 |
| to_address | string | 下车点地址 |
| to_longitude | double | 下车点经度 |
| to_latitude | double | 下车点纬度 |
| estimated_distance_m | int64 | 预估距离（米） |
| estimated_duration_s | int64 | 预估时长（秒） |
| estimated_price_cents | int64 | 预估价格（分） |
| status | OrderStatus | 订单状态 |
| cancel_reason | string | 取消原因（未取消为空） |
| cancel_by | string | 取消方（user/driver/system/admin，未取消为空） |
| created_at | int64 | 创建时间，Unix 秒 |
| updated_at | int64 | 更新时间，Unix 秒 |

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | order_id<=0 |
| 其他 | 订单不存在时由仓储层返回对应错误 |

## 调用示例
**请求**
```json
{
  "order_id": 50001
}
```
**响应**
```json
{
  "order_id": 50001,
  "order_no": "XL202608210001",
  "user_id": 1001,
  "driver_id": 2001,
  "car_type": 2,
  "from_address": "北京市朝阳区望京 SOHO",
  "from_longitude": 116.481,
  "from_latitude": 39.996,
  "to_address": "北京市海淀区中关村",
  "to_longitude": 116.316,
  "to_latitude": 39.983,
  "estimated_distance_m": 12000,
  "estimated_duration_s": 1800,
  "estimated_price_cents": 3750,
  "status": "ORDER_STATUS_ON_TRIP",
  "cancel_reason": "",
  "cancel_by": "",
  "created_at": 1789600000,
  "updated_at": 1789601800
}
```
