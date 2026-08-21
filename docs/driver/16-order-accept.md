# 司机接单接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/accept` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `ordersvc.AcceptOrder` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `orderId` | int64 | 是 | 订单 ID |

## 3. 请求示例

```json
{"orderId":5001}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `orderId` | int64 | 订单 ID |
| `status` | int32 | 订单状态，成功后通常为已接单 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-ORDER-ACCEPT-E01 | 未登录 | HTTP 401 |
| DRIVER-ORDER-ACCEPT-E02 | `orderId<=0` | HTTP 400 |
| DRIVER-ORDER-ACCEPT-E03 | 订单状态不可接单 | ordersvc 状态机错误透传 |

## 6. 处理链路

`api/driver -> OrderLogic.AcceptOrder -> ordersvc.AcceptOrder`。司机 ID 从 JWT 获取，不接受客户端覆盖。
