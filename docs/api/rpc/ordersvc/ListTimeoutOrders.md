# ListTimeoutOrders 查询超时未接单订单

## 接口概述
- **服务**：ordersvc（Order 服务）
- **方法**：`ListTimeoutOrders(ListTimeoutOrdersRequest) returns (ListTimeoutOrdersResponse)`
- **功能**：查询超过指定秒数仍未接单（状态为「待接单」且 driver_id=0）的订单，供超时扫描任务分页拉取后触发 TimeoutCancel。
- **调用方**：系统定时任务 / 超时扫描消费者

## 请求参数（ListTimeoutOrdersRequest）
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| timeout_seconds | int32 | 否 | 超时秒数；0 时默认 300（5 分钟）；不可为负数 |
| page | int32 | 否 | 页码，默认 1 |
| page_size | int32 | 否 | 每页条数，默认 20，上限 100 |

## 响应参数（ListTimeoutOrdersResponse）
| 字段 | 类型 | 说明 |
|------|------|------|
| list | repeated OrderSummary | 超时订单摘要列表 |
| total | int64 | 满足条件的总记录数 |
| page | int32 | 实际页码 |
| page_size | int32 | 实际每页条数 |

> OrderSummary 结构见 ListOrders 文档。

## 查询逻辑
- 计算时间点 `before = now - timeout`，返回创建时间早于 `before` 且仍为「待接单」的订单。

## 错误码
| 错误 | 触发条件 |
|------|----------|
| invalid order params | timeout_seconds < 0 |

## 调用示例
**请求**
```json
{
  "timeout_seconds": 300,
  "page": 1,
  "page_size": 20
}
```
**响应**
```json
{
  "list": [
    {
      "order_id": 50002,
      "order_no": "XL202608210002",
      "from_address": "北京市西城区金融街",
      "to_address": "北京市东城区王府井",
      "status": "ORDER_STATUS_WAIT_ACCEPT",
      "estimated_price_cents": 4200,
      "created_at": 1789599000
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```
