# 开始行程接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/start-trip` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `ordersvc.StartTrip`、`driversvc.SetDriverServiceStatus` |

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
| `status` | int32 | 订单状态，成功后通常为行程中 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-START-E01 | 未登录 | HTTP 401 |
| DRIVER-START-E02 | `orderId<=0` | HTTP 400 |
| DRIVER-START-E03 | 订单非已接单状态 | ordersvc 状态机错误透传 |

## 6. 处理链路

`api/driver -> OrderLogic.StartTrip -> ordersvc.StartTrip -> driversvc.SetDriverServiceStatus(2)`。
