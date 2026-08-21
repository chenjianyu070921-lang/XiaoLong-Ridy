# 司机拒绝派单接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/reject` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `dispatchsvc.RejectDispatch` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `orderId` | int64 | 是 | 订单 ID |
| `reason` | string | 否 | 拒单原因 |

## 3. 请求示例

```json
{"orderId":5001,"reason":"距离较远"}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `orderId` | int64 | 订单 ID |
| `driverId` | int64 | 当前司机 ID |
| `status` | int32 | 拒单后的派单状态 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-ORDER-REJECT-E01 | 未登录 | HTTP 401 |
| DRIVER-ORDER-REJECT-E02 | `orderId<=0` | HTTP 400 |
| DRIVER-ORDER-REJECT-E03 | 派单记录不存在或不属于当前司机 | dispatchsvc 错误透传 |

## 6. 处理链路

`api/driver -> OrderLogic.RejectOrder -> dispatchsvc.RejectDispatch`。派单归属和状态由 `dispatchsvc` 校验。
