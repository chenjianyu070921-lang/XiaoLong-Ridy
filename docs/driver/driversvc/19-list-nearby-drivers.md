# ListNearbyDrivers RPC 接口文档

## 1. 接口说明

| 项 | 值 |
| --- | --- |
| RPC | `ListNearbyDrivers(ListNearbyDriversRequest) returns (ListNearbyDriversResponse)` |
| 对应 HTTP | 无司机端 HTTP 路由，派单引擎内部使用 |
| 当前状态 | 已实现 |
| 业务逻辑 | `ListNearbyDriversLogic.ListNearbyDrivers` |

## 2. 请求字段

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `longitude` | double | 是 | 中心经度 |
| `latitude` | double | 是 | 中心纬度 |
| `radius_meters` | double | 否 | 搜索半径 |
| `limit` | int32 | 否 | 返回上限 |

## 3. 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `drivers` | repeated NearbyDriver | 附近在线司机列表 |

## 4. API 对齐

司机端不直接调用；派单服务可按订单起点查询附近司机。
