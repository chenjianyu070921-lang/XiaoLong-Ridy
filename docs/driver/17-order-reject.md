# 司机拒绝派单接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/reject` |
| 是否登录 | 是，司机 JWT |
| 当前状态 | 已实现 |
| 下游 RPC | `dispatchsvc.RejectDispatch` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `orderId` | int64 | 是 | 订单 ID |
| `reason` | string | 是 | 拒单原因，服务端会 trim，空字符串或全空白会返回参数错误 |

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

## 5. 派单列表字段

`POST /api/driver/v1/orders/dispatches` 返回的每条派单记录会带 `dispatch.rejectReason`，来源于 `dispatch_record.reject_reason`。

## 6. 异常用例

| 场景 | 预期 |
| --- | --- |
| 未登录或 JWT 无效 | HTTP 401 |
| `orderId <= 0` | HTTP 400 |
| `reason` 为空或全空白 | HTTP 400 |
| 派单记录不存在或不属于当前司机 | dispatchsvc 错误透传 |

## 7. 处理链路

`api/driver -> OrderLogic.RejectOrder -> dispatchsvc.RejectDispatch -> dispatch_record.reject_reason`。

API 层先 trim `reason` 并强制非空；dispatchsvc 会写入 `reject_reason`，同时继续写 `remark` 兼容旧展示。
