# 管理后台 HTTP 到 RPC 边界说明

> 最近同步：2026-09-05。本文档以当前 `api/admin` 路由注册、`rpc/adminsvc/admin.proto` 与各 logic 实际调用为准。
> 说明：司机处罚域（`/admin/v1/punishment-rules`、`/punishments`、`/punishment-appeals`）的 HTTP 路由已注册，但 `rpc/adminsvc/admin.proto` 源文件尚未包含对应 RPC 定义，本文档暂不记录该域接口。

## 1. 当前架构边界

`api/admin` 只负责 HTTP 接入层能力：

| 职责 | 说明 |
| --- | --- |
| 路由注册 | 保留现有 `/admin/v1/...` 路径，兼容 Postman 和前端联调 |
| 鉴权 | 读取 `Authorization: Bearer <token>`，调用 `AdminService.ValidateSession` 校验管理员登录态；Redis 会话读写收敛在 `rpc/adminsvc` |
| 参数转换 | 将 Path、Query、Body 转换为 `rpc/adminsvc` 的 protobuf 请求 |
| 响应包装 | 将 RPC 返回包装为统一 JSON：`code/message/data` |

`rpc/adminsvc` 负责后台核心业务动作，并按领域调用下游 RPC：

| 业务模块 | adminsvc RPC 方法（admin.proto） | 下游依赖 |
| --- | --- | --- |
| 管理员管理 | `Register`、`Login`、`Logout`、`ValidateSession`、`Me`、`Menus`、`ListAdmins`、`CreateAdmin`、`UpdateAdmin`、`SetAdminStatus`、`ResetAdminPassword` | 本地表 `admin_user`；Redis 会话 |
| 操作日志 | `ListOperationLogs` | 本地表 `admin_operation_log` |
| 用户管理 | `ListUsers`、`GetUser`、`FreezeUser`、`UnfreezeUser`、`ListUserOrders`、`ListUserCoupons` | `usersvc.AdminListUsers`、`usersvc.AdminGetUser`、`usersvc.AdminFreezeUser`/`AdminUnfreezeUser`、`ordersvc.ListOrders`、`usersvc.ListMyCoupons` |
| 司机管理 | `ListDrivers`、`GetDriver`、`FreezeDriver`、`UnfreezeDriver` | `driversvc.ListDrivers`、`driversvc.GetDriver`、`driversvc.FreezeDriver`、`driversvc.UnfreezeDriver` |
| 司机资质审核 | `ListDriverCertifications`、`GetDriverCertification`、`ApproveDriverCertification`、`RejectDriverCertification` | `driversvc.AdminListCertifications`、`driversvc.AdminGetCertification`、`driversvc.ApproveCertification`、`driversvc.RejectCertification` |
| 司机提现审核 | `ListDriverWithdrawals`、`HandleDriverWithdraw` | `driversvc.AdminListWithdraws`、`driversvc.AuditWithdraw` |
| 订单管理 | `ListOrders`、`GetOrder`、`ListAbnormalOrders`、`CancelOrder`、`RedispatchOrder`、`RefundOrder`、`GetOrderTrack` | `ordersvc.ListOrders`、`ordersvc.GetOrder`/`ListOrderStatusLogs`、`ordersvc.CancelOrder`、`ordersvc.RedispatchOrder`、`ordersvc.ForceRefundOrder`、`locationsvc.GetOrderTrack` |
| 退款补偿 | `ListRefundRetryTasks`、`RetryRefundTask` | 本地表 `admin_refund_compensation_task`；实际补偿由 `job` 的 `RetryRefundEvents`/`RunRefundCompensation` 执行 |
| 优惠券配置 | `ListCoupons`、`CreateCoupon`、`UpdateCoupon`、`DisableCoupon`、`IssueCoupon`、`ListCouponIssueTasks` | 本地表 `coupon`、`admin_coupon_issue_task`；发券调用 `usersvc.AdminIssueCoupon` 写入 `user_coupon` |
| 计价规则 | `ListPriceRules`、`GetPriceRule`、`CreatePriceRule`、`UpdatePriceRule`、`EnablePriceRule`、`DisablePriceRule` | `pricesvc`（adminsvc 不直接读写 `price_rule`） |
| 营销活动 | `ListPromotionActivities`、`CreatePromotionActivity`、`UpdatePromotionActivity`、`PublishPromotionActivity`、`RollbackPromotionActivity` | 本地表 `promotion_activity`；发布/回滚写结构化 `admin_operation_log` |
| 工单管理 | `CreateWorkOrder`、`ListWorkOrders`、`GetWorkOrder`、`ActWorkOrder`、`BatchActWorkOrders`、`AddWorkOrderEvidence`、`ListWorkOrderEvidence` | 本地表 `admin_complaint_work_order`、`admin_work_order_flow`、`admin_work_order_evidence` |
| 统计与运力地图 | `GetStatisticsOverview`、`GetOrderStatistics`、`GetDriverStatistics`、`GetFinanceStatistics`、`GetCouponStatistics`、`GetUserStatistics`、`GetCapacityMap` | 基于业务表实时聚合；运力地图读取 `driversvc.ListDrivers` |
| 导出任务 | `CreateExportTask`、`ListExportTasks`、`GetExportTask`、`GetExportDownload`、`DownloadExport` | 本地表 `admin_export_task`；goroutine 异步生成 CSV，HTTP 层流式返回 |
| 风控管理 | `ListBlacklists`、`AddBlacklist`、`ReleaseBlacklist`、`ListRiskHitRecords`、`HandleRiskHitRecords` | 本地表 `blacklist`、`risk_blacklist_hit_record`；拉黑司机/用户联动 `driversvc.FreezeDriver`/`usersvc.AdminFreezeUser` |
| 审计补偿 | `ListAdminAuditOutbox` | 本地表 `admin_audit_outbox`，实际重试由 `job.RetryAdminAuditOutbox` 执行 |
| AI 运营助手 | `AskAiAgent`、`GetAiSuggestions`、`GetAiHistory`、`AiFeedback`、`DeleteAiConversation` | adminsvc 本地降级/演示引擎；问答会话存 Redis，`Ask`/`Feedback`/`DeleteConversation` 写脱敏审计 |

## 2. HTTP 接口到 RPC 方法映射

除注册接口与登录外，其余全部经 `authRequired`（Bearer token → `ValidateSession` → gRPC metadata `x-admin-token` 透传 → 服务端 RBAC）。

| HTTP 接口（方法见路径行为） | HTTP 层处理 | RPC 方法 |
| --- | --- | --- |
| `POST /admin/v1/auth/register` | 解析注册请求；可选 Bearer Token 用于后续管理员注册授权 | `AdminService.Register` |
| `POST /admin/v1/auth/login` | 解析账号密码 | `AdminService.Login` |
| `POST /admin/v1/auth/logout` | 透传 token | `AdminService.Logout` |
| `GET /admin/v1/auth/me` | 透传 token | `AdminService.Me` |
| `GET /admin/v1/menus` | 透传 token | `AdminService.Menus` |
| `GET|POST /admin/v1/admins` | 管理员列表 / 新增 | `AdminService.ListAdmins` / `CreateAdmin` |
| `PUT /admin/v1/admins/{id}` | 编辑管理员 | `AdminService.UpdateAdmin` |
| `POST /admin/v1/admins/{id}/status` | 启停管理员 | `AdminService.SetAdminStatus` |
| `POST /admin/v1/admins/{id}/reset-password` | 重置密码 | `AdminService.ResetAdminPassword` |
| `GET /admin/v1/operation-logs` | 解析日志筛选条件 | `AdminService.ListOperationLogs` |
| `GET /admin/v1/users` | 解析分页、关键字、状态、时间范围 | `AdminService.ListUsers` |
| `GET /admin/v1/users/{id}` | 解析用户 ID，支持 `sensitive=1` | `AdminService.GetUser` |
| `GET /admin/v1/users/{id}/orders` | 用户订单历史 | `AdminService.ListUserOrders -> ordersvc.ListOrders` |
| `GET /admin/v1/users/{id}/coupons` | 用户优惠券历史 | `AdminService.ListUserCoupons -> usersvc.ListMyCoupons` |
| `POST /admin/v1/users/{id}/freeze` | 解析用户 ID、原因、备注、管理员 ID、IP | `AdminService.FreezeUser` |
| `POST /admin/v1/users/{id}/unfreeze` | 同上 | `AdminService.UnfreezeUser` |
| `GET /admin/v1/user-coupons` | 全局入口，必须带 `user_id` | `AdminService.ListUserCoupons -> usersvc.ListMyCoupons` |
| `GET /admin/v1/drivers` | 解析分页、关键字、状态 | `AdminService.ListDrivers -> driversvc.ListDrivers` |
| `GET /admin/v1/drivers/{id}` | 解析司机 ID，支持 `sensitive=1` | `AdminService.GetDriver -> driversvc.GetDriver` |
| `POST /admin/v1/drivers/{id}/freeze` | 解析司机 ID、原因、备注、管理员 ID、IP | `AdminService.FreezeDriver -> driversvc.FreezeDriver` |
| `POST /admin/v1/drivers/{id}/unfreeze` | 同上 | `AdminService.UnfreezeDriver -> driversvc.UnfreezeDriver` |
| `GET /admin/v1/driver-withdrawals` | 解析提现筛选条件 | `AdminService.ListDriverWithdrawals -> driversvc.AdminListWithdraws` |
| `POST /admin/v1/driver-withdrawals/{id}/approve` | 审核打款成功 | `AdminService.HandleDriverWithdraw -> driversvc.AuditWithdraw` |
| `POST /admin/v1/driver-withdrawals/{id}/reject` | 审核打款失败 | `AdminService.HandleDriverWithdraw -> driversvc.AuditWithdraw` |
| `GET /admin/v1/driver-certifications` | 解析审核状态和筛选条件 | `AdminService.ListDriverCertifications -> driversvc.AdminListCertifications` |
| `GET /admin/v1/driver-certifications/{id}` | 解析审核记录 ID | `AdminService.GetDriverCertification -> driversvc.AdminGetCertification` |
| `POST /admin/v1/driver-certifications/{id}/approve` | 解析审核 ID、备注、管理员 ID、IP | `AdminService.ApproveDriverCertification -> driversvc.ApproveCertification` |
| `POST /admin/v1/driver-certifications/{id}/reject` | 解析审核 ID、驳回原因、管理员 ID、IP | `AdminService.RejectDriverCertification -> driversvc.RejectCertification` |
| `GET /admin/v1/orders` | 解析订单筛选条件 | `AdminService.ListOrders -> ordersvc.ListOrders` |
| `GET /admin/v1/orders/abnormal` | 解析 `abnormal_type=cancel/payment/dispatch` | `AdminService.ListAbnormalOrders` |
| `GET /admin/v1/orders/{id}` | 解析订单 ID | `AdminService.GetOrder -> ordersvc.GetOrder`（含状态日志等聚合） |
| `GET /admin/v1/orders/{id}/track` | 解析订单轨迹 | `AdminService.GetOrderTrack -> locationsvc.GetOrderTrack` |
| `POST /admin/v1/orders/{id}/cancel` | 解析订单 ID、取消原因、管理员 ID、IP | `AdminService.CancelOrder -> ordersvc.CancelOrder` |
| `POST /admin/v1/orders/{id}/redispatch` | 解析目标司机、原因、request_id | `AdminService.RedispatchOrder -> ordersvc.RedispatchOrder` |
| `POST /admin/v1/orders/{id}/refund` | 解析退款金额、原因、request_id | `AdminService.RefundOrder -> ordersvc.ForceRefundOrder` |
| `GET /admin/v1/refund-retry-tasks` | 查询退款补偿队列 | `AdminService.ListRefundRetryTasks` |
| `POST /admin/v1/refund-retry-tasks/{refund_no}` | 立即触发指定退款补偿 | `AdminService.RetryRefundTask` |
| `GET /admin/v1/coupons` | 解析优惠券筛选条件 | `AdminService.ListCoupons` |
| `POST /admin/v1/coupons` | 解析优惠券模板请求体、管理员 ID、IP | `AdminService.CreateCoupon` |
| `PUT /admin/v1/coupons/{id}` | 解析优惠券 ID、模板请求体、管理员 ID、IP | `AdminService.UpdateCoupon` |
| `POST /admin/v1/coupons/{id}/disable` | 解析优惠券 ID、管理员 ID、IP | `AdminService.DisableCoupon` |
| `POST /admin/v1/coupons/{id}/issue` | 解析优惠券 ID、发券目标、管理员 ID、IP | `AdminService.IssueCoupon` |
| `GET /admin/v1/coupon-issue-tasks` | 解析发券任务筛选条件 | `AdminService.ListCouponIssueTasks` |
| `GET /admin/v1/price-rules` | 解析计价规则筛选条件 | `AdminService.ListPriceRules -> pricesvc.ListPriceRules` |
| `POST /admin/v1/price-rules` | 解析计价规则配置、管理员 ID、IP | `AdminService.CreatePriceRule -> pricesvc.CreatePriceRule` |
| `GET /admin/v1/price-rules/{id}` | 解析计价规则 ID | `AdminService.GetPriceRule -> pricesvc.GetPriceRule` |
| `PUT /admin/v1/price-rules/{id}` | 解析计价规则 ID、配置、管理员 ID、IP | `AdminService.UpdatePriceRule -> pricesvc.UpdatePriceRule` |
| `POST /admin/v1/price-rules/{id}/enable` | 启用计价规则 | `AdminService.EnablePriceRule -> pricesvc.EnablePriceRule` |
| `POST /admin/v1/price-rules/{id}/disable` | 停用计价规则 | `AdminService.DisablePriceRule -> pricesvc.DisablePriceRule` |
| `GET /admin/v1/promotion-activities` | 解析活动筛选条件 | `AdminService.ListPromotionActivities` |
| `POST /admin/v1/promotion-activities` | 解析活动配置、管理员 ID、IP | `AdminService.CreatePromotionActivity` |
| `PUT /admin/v1/promotion-activities/{id}` | 解析活动 ID、活动配置、管理员 ID、IP | `AdminService.UpdatePromotionActivity` |
| `POST /admin/v1/promotion-activities/{id}/publish` | 解析发布范围和灰度配置 | `AdminService.PublishPromotionActivity` |
| `POST /admin/v1/promotion-activities/{id}/rollback` | 解析回滚配置 | `AdminService.RollbackPromotionActivity` |
| `GET /admin/v1/statistics/overview` | 解析统计时间范围 | `AdminService.GetStatisticsOverview` |
| `GET /admin/v1/statistics/orders` | 同上 | `AdminService.GetOrderStatistics` |
| `GET /admin/v1/statistics/drivers` | 同上 | `AdminService.GetDriverStatistics` |
| `GET /admin/v1/statistics/revenue` | 同上 | `AdminService.GetFinanceStatistics` |
| `GET /admin/v1/statistics/coupons` | 同上 | `AdminService.GetCouponStatistics` |
| `GET /admin/v1/statistics/users` | 同上 | `AdminService.GetUserStatistics` |
| `GET /admin/v1/capacity/map` | 解析状态、听单状态、limit | `AdminService.GetCapacityMap -> driversvc.ListDrivers` |
| `POST /admin/v1/export-tasks` | 解析导出类型、筛选条件、管理员 ID、IP | `AdminService.CreateExportTask` |
| `GET /admin/v1/export-tasks` | 解析导出任务筛选条件 | `AdminService.ListExportTasks` |
| `GET /admin/v1/export-tasks/{task_no}` | 解析导出任务编号 | `AdminService.GetExportTask` |
| `GET /admin/v1/export-tasks/{task_no}/download` | 鉴权后流式返回 CSV | `AdminService.DownloadExport`（HTTP 直接建 stream 客户端） |
| `GET|POST /admin/v1/work-orders` | 工单列表 / 创建 | `AdminService.ListWorkOrders` / `CreateWorkOrder` |
| `GET /admin/v1/work-orders/{id}` | 工单详情 | `AdminService.GetWorkOrder` |
| `POST /admin/v1/work-orders/{id}/actions` | 工单流转 | `AdminService.ActWorkOrder` |
| `POST /admin/v1/work-orders/batch-actions` | 批量流转 | `AdminService.BatchActWorkOrders` |
| `GET /admin/v1/work-orders/{id}/evidence` | 证据列表 | `AdminService.ListWorkOrderEvidence` |
| `POST /admin/v1/work-orders/{id}/evidence` | 新增证据索引 | `AdminService.AddWorkOrderEvidence` |
| `GET /admin/v1/blacklist` | 解析黑名单筛选条件 | `AdminService.ListBlacklists` |
| `POST /admin/v1/blacklist` | 解析拉黑对象、原因、管理员 ID、IP | `AdminService.AddBlacklist` |
| `POST|PATCH /admin/v1/blacklist/{id}/release` | 解析黑名单 ID、解除原因 | `AdminService.ReleaseBlacklist` |
| `GET /admin/v1/risk/hit-records` | 解析风控命中筛选条件 | `AdminService.ListRiskHitRecords` |
| `POST /admin/v1/risk/hit-records/actions` | 复核/拉黑/转工单 | `AdminService.HandleRiskHitRecords` |
| `GET /admin/v1/notification-outbox` | 解析通知/审计补偿筛选条件 | `AdminService.ListAdminAuditOutbox` |
| `POST /admin/v1/ai-agent/ask` | 提交受限运营问答 | `AdminService.AskAiAgent` |
| `GET /admin/v1/ai-agent/suggestions` | 快捷问题 | `AdminService.GetAiSuggestions` |
| `GET /admin/v1/ai-agent/history` | 本人会话摘要 | `AdminService.GetAiHistory` |
| `POST /admin/v1/ai-agent/feedback` | 记录是否有帮助 | `AdminService.AiFeedback` |
| `DELETE /admin/v1/ai-agent/conversations/{id}` | 结束并清空会话 | `AdminService.DeleteAiConversation` |
| `GET /healthz`、`GET /` | 健康检查与服务说明 | 无 |

## 3. 数据写入边界

| 操作 | 写入位置 | 说明 |
| --- | --- | --- |
| 司机审核通过 | `rpc/driversvc` + `rpc/adminsvc` | `adminsvc` 同步调用 `driversvc.ApproveCertification`；driversvc 本地事务更新 `driver_certification`、`driver`、`driver_vehicle`；adminsvc 写审计日志，失败时写 `admin_audit_outbox` 补偿 |
| 司机审核驳回 | `rpc/driversvc` + `rpc/adminsvc` | `adminsvc` 同步调用 `driversvc.RejectCertification`；driversvc 只更新审核状态；adminsvc 写审计，失败时写 outbox |
| 司机冻结/解冻 | `rpc/driversvc` + `rpc/adminsvc` | 调用 `driversvc.FreezeDriver`/`UnfreezeDriver`；冻结后写审计并通过 `pushsvc.SendNotice/SendPush` 通知司机，通知失败写 `admin_audit_outbox` |
| 司机风控拉黑 | `rpc/adminsvc` + `rpc/driversvc`/`rpc/usersvc` + `rpc/pushesvc` | 写 `blacklist`；目标为 `driver` 联动 `driversvc.FreezeDriver`、为 `user` 联动 `usersvc.AdminFreezeUser` |
| 风控命中处置 | `rpc/adminsvc` | `review_pass` 写审计，`add_blacklist` 写黑名单，`create_work_order` 写工单；处置状态由审计、黑名单和工单表推导，不新增数据库字段 |
| 敏感信息查看 | `rpc/adminsvc` + `rpc/usersvc` / `rpc/driversvc` | 用户和司机详情默认脱敏；仅 `sensitive=1` 且超管/运营时返回明文并写 `admin_operation_log` |
| 用户冻结/解封 | `rpc/adminsvc` | 调用 `usersvc.AdminFreezeUser`/`AdminUnfreezeUser` 更新 `user.status`，并写审计 |
| 司机提现审核 | `rpc/driversvc` + `rpc/adminsvc` | 调用 `driversvc.AuditWithdraw` 打款成功/失败流转，adminsvc 写审计 |
| 优惠券新增/编辑/下架 | `rpc/adminsvc` | 写入或更新 `coupon`；新增与 `admin_operation_log` 同事务，返回新 ID |
| 优惠券发放任务 | `rpc/adminsvc` + `rpc/usersvc` | 写入 `admin_coupon_issue_task`，调用 `usersvc.AdminIssueCoupon` 同步写 `user_coupon` 并更新 `coupon.received_count` |
| 计价规则管理 | `rpc/adminsvc` + `rpc/pricesvc` | adminsvc 转发到 `pricesvc` 完成读写与启停，不直接修改 `price_rule` |
| 活动配置发布/回滚 | `rpc/adminsvc` | 更新 `promotion_activity.status`，写结构化 `admin_operation_log` |
| 工单与证据 | `rpc/adminsvc` | 写入 `admin_complaint_work_order`、`admin_work_order_flow`、`admin_work_order_evidence` |
| 导出任务 | `rpc/adminsvc` | 写入 `admin_export_task`，goroutine 异步生成 CSV；支持 `users`、`drivers`、`orders`、`operation_logs`、`statistics` 五类导出 |
| 退款补偿 | `rpc/adminsvc` + `rpc/ordersvc` | `RefundOrder` 调用 `ordersvc.ForceRefundOrder`；失败/待重试项写入 `admin_refund_compensation_task`，由 `job` 重试补偿 |
| 操作审计 | `rpc/adminsvc` | 写 `admin_operation_log`；失败时按动作写 `admin_audit_outbox` |

管理后台 outbox 补偿策略（`job`）：

| 任务 | 频率 | 覆盖范围 |
| --- | --- | --- |
| `RetryAdminAuditOutbox` | 每 30 秒 | 重写审计日志、重放 `driversvc.FreezeDriver`、重试 `pushsvc.SendNotice/SendPush` |
| `RetryAdminDomainOutbox` | 每 10 秒 | 将 `admin_domain_outbox` 中处罚、退款、发券、活动和通知事件可靠投递 Kafka |
| `RetryRefundEvents` / `RunRefundCompensation` | 每 10 秒 | 消费退款补偿队列/`admin_refund_compensation_task`，推进 `pending/processing/retrying/success/manual_review/failed` 状态 |

## 4. 联调要求

1. 先启动下游 RPC（usersvc/ordersvc/driversvc/pricesvc/locationsvc/pushsvc 等），再启动 `rpc/adminsvc`，默认监听 `0.0.0.0:8084`。
2. 再启动 `api/admin`，默认监听 `0.0.0.0:8717`。
3. 前端和 Postman 只调用原 HTTP 路径，不直接调用 RPC。
4. MySQL 使用统一配置的远程 `xiaolong_ridy` 数据库（`115.191.16.159`）；本地一键启动见 `scripts/run-passenger-to-admin.ps1`。

## 5. 2026-09 修订说明（相对历史版本的现状更新）

1. `ListUsers`/`GetUser` 已切 `usersvc.AdminListUsers`/`AdminGetUser`，不再由 adminsvc 直读用户表；`ListOrders`/`GetOrder` 已切 `ordersvc`，不再直读订单表。
2. 新增模块均已开放：管理员管理（`/admins`）、司机提现（`/driver-withdrawals`）、司机解冻、订单人工改派/退款/轨迹（`/orders/{id}/track|redispatch|refund`）、退款补偿队列（`/refund-retry-tasks`）、运力地图（`/capacity/map`）、工单（`/work-orders`，含证据与批量动作）、导出下载（`/export-tasks/{task_no}/download`）、AI 运营助手（`/ai-agent/*`）、审计补偿查询（`/notification-outbox`）。
3. 统计路由现为 `/statistics/{overview,orders,drivers,revenue,coupons,users}` 六个子资源。
4. 司机处罚域路由已注册但 `admin.proto` 尚未同步对应 RPC，本文档暂不列示（待 proto 补全后补充）。
