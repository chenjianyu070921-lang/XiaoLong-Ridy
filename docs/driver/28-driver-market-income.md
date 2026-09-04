# 抢单大厅与收入查询接口文档

## 1. 抢单大厅

| 项 | 说明 |
| --- | --- |
| 请求方法 | `POST` |
| 请求路径 | `/api/driver/v1/orders/available` |
| 认证 | `Authorization: Bearer <JWT>` |
| driverId 来源 | 只从 JWT 解析，前端禁止传入 |

### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| page | int32 | 否 | 页码，默认由服务端归一化 |
| pageSize | int32 | 否 | 每页数量，默认由服务端归一化 |
| status | int32 | 否 | 当前接口忽略客户端状态，只查询 WAIT_ACCEPT |

### 处理链路

1. API 层从 JWT 解析司机 ID。
2. 读取 Redis `driver:online` 校验司机在线；不在线直接返回空列表。
3. 读取 Redis `driver:pos:<driver_id>` 获取司机实时经纬度；位置不存在或非法直接返回空列表。
4. 调用 `ordersvc.ListOrders(status=WAIT_ACCEPT, page=1, pageSize=100)` 获取待抢订单。
5. 循环调用 `ordersvc.GetOrder` 获取订单起点经纬度。
6. API 层按司机位置到订单起点的距离做内存过滤、排序、分页。

### 业务边界

这是公开抢单大厅模型：司机可看到附近所有 `WAIT_ACCEPT` 待接单订单并主动抢单，不使用 `dispatch_record` 定向派单过滤。当前半径为 `3000m`，结果按距离由近到远排序，距离相同按创建时间排序。

## 2. 今日收入

| 项 | 说明 |
| --- | --- |
| 请求方法 | `GET` |
| 请求路径 | `/api/driver/v1/income/today` |
| 认证 | `Authorization: Bearer <JWT>` |

### 响应字段

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| driverId | int64 | 当前司机 ID |
| period | string | 固定为 `today` |
| completedOrders | int64 | 今日完成订单数 |
| totalIncomeCents | int64 | 今日收入，单位分 |
| startAt | int64 | 今日 00:00 Unix 秒 |
| endAt | int64 | 明日 00:00 Unix 秒 |
| source | string | 数据源，当前为 `ordersvc.completed_orders` + 订单详情结算金额 |

## 3. 本周收入

| 项 | 说明 |
| --- | --- |
| 请求方法 | `GET` |
| 请求路径 | `/api/driver/v1/income/week` |
| 认证 | `Authorization: Bearer <JWT>` |

本周收入按周一 00:00 到下周一 00:00 聚合，只统计当前司机已完成订单。当前 `ordersvc.ListOrders` 没有时间范围参数，因此司机 API 会分页读取当前司机 `COMPLETED` 订单后，按 `createdAt` 在内存聚合。

## 4. 计价能力边界

当前项目不做行程中实时刷新金额。

现有计价只有两类：

- 下单预估：`pricesvc.EstimatePrice` 按预估里程和时长计算，订单落 `estimated_price` 快照。
- 结束结算：`ordersvc.FinishTrip` 在行程结束时按实际里程和时长计算服务端权威金额，并通过 `pricesvc.SaveActualOrderPrice` 保存实际计价明细。

司机收入查询只统计已完成订单，金额口径使用订单结算金额，不使用下单预估价作为收入。
