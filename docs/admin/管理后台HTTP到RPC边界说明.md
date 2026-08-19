# 管理后台 HTTP 到 RPC 边界说明

## 1. 当前架构边界

`api/admin` 只负责 HTTP 接入层能力：

| 职责 | 说明 |
| --- | --- |
| 路由注册 | 保留现有 `/admin/v1/...` 路径，兼容 Postman 和前端联调 |
| 鉴权 | 读取 `Authorization: Bearer <token>`，校验 Redis 管理员登录态 |
| 参数转换 | 将 Path、Query、Body 转换为 `rpc/adminsvc` 的 protobuf 请求 |
| 响应包装 | 将 RPC 返回包装为统一 JSON：`code/message/data` |

`rpc/adminsvc` 负责后台核心业务动作：

| 业务模块 | RPC 方法 |
| --- | --- |
| 用户管理 | `ListUsers`、`GetUser`、`FreezeUser`、`UnfreezeUser` |
| 司机审核 | `ListDriverCertifications`、`GetDriverCertification`、`ApproveDriverCertification`、`RejectDriverCertification` |
| 订单管理 | `ListOrders`、`GetOrder`、`ListAbnormalOrders` |
| 优惠券配置 | `ListCoupons`、`CreateCoupon`、`UpdateCoupon`、`DisableCoupon` |
| 操作日志 | `ListOperationLogs` |

## 2. HTTP 接口到 RPC 方法映射

| HTTP 接口 | HTTP 层处理 | RPC 方法 |
| --- | --- | --- |
| `GET /admin/v1/users` | 解析分页、关键字、状态、时间范围 | `AdminService.ListUsers` |
| `GET /admin/v1/users/{id}` | 解析用户 ID | `AdminService.GetUser` |
| `GET /admin/v1/driver-certifications` | 解析审核状态和筛选条件 | `AdminService.ListDriverCertifications` |
| `GET /admin/v1/driver-certifications/{id}` | 解析审核记录 ID | `AdminService.GetDriverCertification` |
| `POST /admin/v1/driver-certifications/{id}/approve` | 解析审核 ID、备注、管理员 ID、IP | `AdminService.ApproveDriverCertification` |
| `POST /admin/v1/driver-certifications/{id}/reject` | 解析审核 ID、驳回原因、管理员 ID、IP | `AdminService.RejectDriverCertification` |
| `GET /admin/v1/orders` | 解析订单筛选条件 | `AdminService.ListOrders` |
| `GET /admin/v1/orders/{id}` | 解析订单 ID | `AdminService.GetOrder` |
| `GET /admin/v1/orders/abnormal` | 解析 `abnormal_type=cancel/payment/dispatch` | `AdminService.ListAbnormalOrders` |
| `GET /admin/v1/coupons` | 解析优惠券筛选条件 | `AdminService.ListCoupons` |
| `POST /admin/v1/coupons` | 解析优惠券模板请求体、管理员 ID、IP | `AdminService.CreateCoupon` |
| `PUT /admin/v1/coupons/{id}` | 解析优惠券 ID、模板请求体、管理员 ID、IP | `AdminService.UpdateCoupon` |
| `POST /admin/v1/coupons/{id}/issue` | 解析优惠券 ID、发券目标、管理员 ID、IP | `AdminService.IssueCoupon` |
| `GET /admin/v1/coupon-issue-tasks` | 解析发券任务筛选条件 | `AdminService.ListCouponIssueTasks` |
| `GET /admin/v1/promotion-activities` | 解析活动筛选条件 | `AdminService.ListPromotionActivities` |
| `POST /admin/v1/promotion-activities` | 解析活动配置、管理员 ID、IP | `AdminService.CreatePromotionActivity` |
| `PUT /admin/v1/promotion-activities/{id}` | 解析活动 ID、活动配置、管理员 ID、IP | `AdminService.UpdatePromotionActivity` |
| `POST /admin/v1/promotion-activities/{id}/publish` | 解析发布范围和灰度配置 | `AdminService.PublishPromotionActivity` |
| `POST /admin/v1/promotion-activities/{id}/rollback` | 解析回滚配置 | `AdminService.RollbackPromotionActivity` |
| `GET /admin/v1/statistics/overview` | 解析统计时间范围 | `AdminService.GetStatisticsOverview` |
| `GET /admin/v1/statistics/orders` | 解析统计时间范围 | `AdminService.GetOrderStatistics` |
| `GET /admin/v1/statistics/coupons` | 解析统计时间范围 | `AdminService.GetCouponStatistics` |
| `POST /admin/v1/export-tasks` | 解析导出类型、筛选条件、管理员 ID、IP | `AdminService.CreateExportTask` |
| `GET /admin/v1/export-tasks` | 解析导出任务筛选条件 | `AdminService.ListExportTasks` |
| `GET /admin/v1/blacklist` | 解析黑名单筛选条件 | `AdminService.ListBlacklists` |
| `POST /admin/v1/blacklist` | 解析拉黑对象、原因、管理员 ID、IP | `AdminService.AddBlacklist` |
| `POST /admin/v1/blacklist/{id}/release` | 解析黑名单 ID、解除原因 | `AdminService.ReleaseBlacklist` |
| `GET /admin/v1/risk/hit-records` | 解析风控命中筛选条件 | `AdminService.ListRiskHitRecords` |
| `GET /admin/v1/operation-logs` | 解析日志筛选条件 | `AdminService.ListOperationLogs` |

## 3. 数据写入边界

| 操作 | 写入位置 | 说明 |
| --- | --- | --- |
| 司机审核通过 | `rpc/driversvc` + `rpc/adminsvc` | `adminsvc` 同步调用 `driversvc.ApproveCertification`；driversvc 更新 `driver_certification`、`driver`、`driver_vehicle`；adminsvc 写审计日志 |
| 司机审核驳回 | `rpc/driversvc` + `rpc/adminsvc` | `adminsvc` 同步调用 `driversvc.RejectCertification`；driversvc 更新审核状态；adminsvc 写审计日志 |
| 优惠券新增/编辑/下架 | `rpc/adminsvc` | 写入或更新 `coupon` 表 |
| 优惠券发放任务 | `rpc/adminsvc` | 写入 `admin_coupon_issue_task`，同步写入 `user_coupon`，更新 `coupon.received_count` |
| 活动配置发布/回滚 | `rpc/adminsvc` | 更新 `promotion_activity.status`，写入 `admin_operation_log` |
| 风控黑名单 | `rpc/adminsvc` | 写入或更新 `blacklist`，查询 `risk_blacklist_hit_record` |
| 导出任务 | `rpc/adminsvc` | 当前不新增导出任务表，使用 `admin_operation_log` 记录可追踪任务 |
| 用户冻结/解封 | `rpc/adminsvc` | 更新 `user.status` |
| 操作审计 | `rpc/adminsvc` | 写入 `admin_operation_log` |

## 4. 联调要求

1. 先启动 `rpc/adminsvc`，默认监听 `127.0.0.1:8084`。
2. 再启动 `api/admin`，默认监听 `127.0.0.1:8083`，并读取 `api/admin/etc/admin.json` 中的 `admin_rpc.target=127.0.0.1:8084`。
3. 前端和 Postman 仍然调用原 HTTP 路径，不直接调用 RPC。
4. MySQL 使用统一配置的远程 `xiaolong_ridy` 数据库，Redis 使用本地 Docker `6379`。

## 5. 2026-08-15 P0 当前实现补充

### 5.1 后台订单取消

当前已按 `api/admin -> rpc/adminsvc -> rpc/ordersvc` 边界落地后台订单取消。

| 层级 | 已落地内容 |
| --- | --- |
| HTTP 路由 | `POST /admin/v1/orders/{id}/cancel` |
| HTTP 职责 | 校验登录态，解析订单 ID、取消原因、管理员 ID、客户端 IP |
| adminsvc RPC | `AdminService.CancelOrder(AdminCancelOrderRequest) returns (CommonResponse)` |
| 下游 RPC | 同步调用 `ordersvc.CancelOrder` |
| 下游入参 | `order_id`、`operator_type=admin`、`operator_id=admin_id`、`reason` |
| 当前验证 | `go test ./api/admin/...`、`go test ./rpc/adminsvc/...` 已通过 |

### 5.2 司机审核边界

当前已切换为 `api/admin -> adminsvc -> driversvc`：

| 动作 | adminsvc 职责 | driversvc 职责 |
| --- | --- | --- |
| 审核通过 | 参数校验、调用 `driversvc.ApproveCertification`、写 `admin_operation_log` | 事务更新 `driver_certification.audit_status=2`、`driver.status=2`、`driver_vehicle.status=2` |
| 审核驳回 | 参数校验、调用 `driversvc.RejectCertification`、写 `admin_operation_log` | 事务更新 `driver_certification.audit_status=3` 和驳回原因 |

### 5.3 P1/P2 当前落地边界

| 模块 | 当前落地 |
| --- | --- |
| 优惠券发放任务 | `POST /admin/v1/coupons/{id}/issue` 同步创建任务并写 `user_coupon`；`GET /admin/v1/coupon-issue-tasks` 查询任务 |
| 活动配置 | 支持活动列表、新增、编辑、发布、回滚；发布和回滚当前更新 `promotion_activity.status` 并写操作日志 |
| 数据统计 | 支持运营总览、订单统计、优惠券统计，基于现有业务表实时聚合 |
| 导出任务 | 支持创建和查询导出任务；当前 SQL 未定义独立导出表，先使用 `admin_operation_log` 承载任务记录 |
| 风控管理 | 支持黑名单列表、新增、解除，以及命中记录查询 |
