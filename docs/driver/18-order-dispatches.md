# 我的派单列表接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/dispatches` |
| 是否登录 | 是 |
| 当前状态 | 已实现 |
| 下游 RPC | `dispatchsvc.ListDispatchRecords`、`ordersvc.GetOrder` |

## 2. 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `page` | int32 | 否 | 页码，默认 1 |
| `pageSize` | int32 | 否 | 每页条数，默认 20，上限 100 |
| `status` | int32 | 否 | 派单状态，0 时默认查询待处理状态 |

## 3. 请求示例

```json
{"page":1,"pageSize":20,"status":1}
```

## 4. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `list` | array | 派单记录与订单摘要组合列表 |
| `list[].dispatch` | object | 派单记录 |
| `list[].order` | object | 订单摘要 |
| `total` | int64 | 总条数 |
| `page` | int32 | 当前页 |
| `pageSize` | int32 | 每页条数 |

## 5. 异常用例

| 用例编号 | 场景 | 预期 |
| --- | --- | --- |
| DRIVER-DISPATCHES-E01 | 未登录 | HTTP 401 |
| DRIVER-DISPATCHES-E02 | `status<0` | HTTP 400 |
| DRIVER-DISPATCHES-E03 | dispatchsvc 不可用 | HTTP 502 |
| DRIVER-DISPATCHES-E04 | ordersvc.GetOrder 失败 | 错误透传 |

## 6. 处理链路

`api/driver -> OrderLogic.ListMyDispatches -> dispatchsvc.ListDispatchRecords(driver_id) -> ordersvc.GetOrder`。
