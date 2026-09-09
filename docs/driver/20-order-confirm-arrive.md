# 确认到达接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/confirm-arrive` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `ordersvc.ConfirmArrive` |

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
| `status` | int32 | 当前订单状态 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-ARRIVE-E01 | 未登录 | HTTP 401 |
| DRIVER-ARRIVE-E02 | `orderId<=0` | HTTP 400 |
| DRIVER-ARRIVE-E03 | 订单不属于当前司机 | ordersvc 错误透传 |

## 6. 处理链路

`api/driver -> OrderLogic.ConfirmArrive -> ordersvc.ConfirmArrive`。该动作记录司机已到达，订单状态通常保持已接单。
