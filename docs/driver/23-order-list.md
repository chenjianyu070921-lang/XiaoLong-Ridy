# 我的订单列表接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/list` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `ordersvc.ListOrders` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `page` | int32 | 否 | 页码，默认 1 |
| `pageSize` | int32 | 否 | 每页条数，默认 20，上限 100 |
| `status` | int32 | 否 | 订单状态，0 为全部，1 待接单，2 已接单，3 行程中，4 待支付，5 已完成，6 已取消 |

## 3. 请求示例

```json
{"page":1,"pageSize":8,"status":0}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `list` | array | 订单摘要列表 |
| `list[].orderId` | int64 | 订单 ID |
| `list[].orderNo` | string | 订单编号 |
| `list[].fromAddress` | string | 上车点 |
| `list[].toAddress` | string | 目的地 |
| `list[].status` | int32 | 订单状态 |
| `list[].estimatedPriceCents` | int64 | 预估价格，单位分 |
| `list[].createdAt` | int64 | 创建时间，Unix 秒 |
| `total` | int64 | 总条数 |
| `page` | int32 | 当前页 |
| `pageSize` | int32 | 每页条数 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-ORDER-LIST-E01 | 未登录 | HTTP 401 |
| DRIVER-ORDER-LIST-E02 | `page<0`、`pageSize<0` 或 `status` 不在 0-6 | HTTP 400 |
| DRIVER-ORDER-LIST-E03 | ordersvc 不可用 | HTTP 502 |

## 6. 处理链路

`api/driver -> OrderLogic.ListMyOrders -> ordersvc.ListOrders(driver_id)`。
