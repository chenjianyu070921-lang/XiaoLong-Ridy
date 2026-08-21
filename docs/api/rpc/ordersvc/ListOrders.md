# ListOrders 分页查询订单列表

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`ListOrders(ListOrdersRequest) returns (ListOrdersResponse)`
- **功能**：按乘客 / 司机 / 状态组合条件分页查询订单摘要列表。
- **调用方**：乘客端（C）、司机端（B）、管理后台（D）

## 请求参数（ListOrdersRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| user_id | int64 | 否 | 按乘客查询（传 0 表示不按乘客过滤） |
| driver_id | int64 | 否 | 按司机查询（传 0 表示不按司机过滤） |
| status | OrderStatus | 否 | 按状态过滤（0 表示不按状态过滤） |
| page | int32 | 否 | 页码，默认 1 |
| page_size | int32 | 否 | 每页条数，默认 20，上限 100 |

## 响应参数（ListOrdersResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| list | repeated OrderSummary | 订单摘要列表 |
| total | int64 | 满足条件的总记录数 |
| page | int32 | 实际页码 |
| page_size | int32 | 实际每页条数 |

### OrderSummary 结构
| 字段 | 类型 | 说明 |
|------|------|------|
| order_id | int64 | 订单 ID |
| order_no | string | 订单编号 |
| from_address | string | 上车点地址 |
| to_address | string | 下车点地址 |
| status | OrderStatus | 订单状态 |
| estimated_price_cents | int64 | 预估价格（分） |
| created_at | int64 | 创建时间，Unix 秒 |

## 分页规则
- page <= 0 时归一化为 1；page_size <= 0 时归一化为 20；page_size > 100 时截断为 100。
- status 取值必须在 0 ~ CANCELLED(6) 之间，否则返回 `invalid order params`。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | status 取值越界 |

## 调用示例
**请求**
```json
{
  "user_id": 1001,
  "status": "ORDER_STATUS_ON_TRIP",
  "page": 1,
  "page_size": 20
}
```
**响应**
```json
{
  "list": [
    {
      "order_id": 50001,
      "order_no": "XL202608210001",
      "from_address": "北京市朝阳区望京 SOHO",
      "to_address": "北京市海淀区中关村",
      "status": "ORDER_STATUS_ON_TRIP",
      "estimated_price_cents": 3750,
      "created_at": 1789600000
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```
