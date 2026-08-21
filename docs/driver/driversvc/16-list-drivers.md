# ListDrivers RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `ListDrivers(ListDriversRequest) returns (ListDriversResponse)` |
| 对应 HTTP | 无司机端 HTTP 路由 |
| 当前状态 | 已实现 |
| 业务逻辑 | `ListDriversLogic.ListDrivers` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `page` | int32 | 否 | 默认 1 |
| `page_size` | int32 | 否 | 默认 20，上限 100 |
| `status` | optional DriverStatus | 否 | 状态过滤 |
| `keyword` | string | 否 | 关键字 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `drivers` | repeated Driver | 司机列表 |
| `total` | int64 | 总数 |

## 4. API 对齐

当前无司机端 HTTP 路由；后台或内部查询司机列表时使用。
