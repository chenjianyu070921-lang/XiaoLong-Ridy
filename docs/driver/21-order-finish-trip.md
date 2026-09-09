# 结束行程接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/finish-trip` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `ordersvc.FinishTrip`、`driversvc.SetDriverServiceStatus` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `orderId` | int64 | 是 | 订单 ID |
| `actualDistanceM` | int64 | 是 | 实际里程，单位米 |
| `actualDurationS` | int64 | 是 | 实际时长，单位秒 |
| `actualPriceCents` | int64 | 是 | 实际金额，单位分 |

## 3. 请求示例

```json
{"orderId":5001,"actualDistanceM":1200,"actualDurationS":600,"actualPriceCents":1800}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `orderId` | int64 | 订单 ID |
| `status` | int32 | 订单状态，成功后通常为待支付 |
| `payableAmountCents` | int64 | 应付金额，单位分 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-FINISH-E01 | 未登录 | HTTP 401 |
| DRIVER-FINISH-E02 | `orderId<=0` | HTTP 400 |
| DRIVER-FINISH-E03 | 实际里程/时长/金额为负数 | HTTP 400 |
| DRIVER-FINISH-E04 | 订单非行程中状态 | ordersvc 状态机错误透传 |

## 6. 处理链路

`api/driver -> OrderLogic.FinishTrip -> ordersvc.FinishTrip -> driversvc.SetDriverServiceStatus(1)`。
