# 管理后台全模块 RPC 依赖调用清单

> 适用范围：`api/admin` 管理后台。  
> 当前实现：`api/admin` 统一调用 `rpc/adminsvc`；`adminsvc` 负责数据库读写和必要的下游 RPC 调用。  
> 最近同步：2026-09-05。  
> 说明：司机处罚域 RPC（`punishment-*`）的 HTTP 路由已注册，但 `rpc/adminsvc/admin.proto` 源文件尚未包含对应定义，本清单不列示该域。  
> 调用类型：同步 RPC 用于强一致查询和状态修改；异步 MQ/job 用于通知、补偿、批量任务和最终一致场景。

## 一、调用规范

| 规范项 | 要求 |
| --- | --- |
| 鉴权 | 后台 HTTP 层读取 Bearer Token，并调用 `adminsvc.ValidateSession` 校验登录态；已校验 token 以 gRPC metadata `x-admin-token` 透传 adminsvc，服务端再做 RBAC |
| 幂等 | 改派、退款、发券、审核、处罚等资金/状态变更必须传 `request_id` 或业务幂等号 |
| 审计 | 所有敏感操作由 `adminsvc` 写入 `admin_operation_log`；失败时写 `admin_audit_outbox` 或领域事件 outbox 补偿 |
| 超时 | 下游 RPC 服务端统一放宽至 30s（`Timeout: 30000`），避免远端 MySQL 慢查询先被截断 |
| 降级 | 订单详情等聚合查询允许部分模块超时后返回主信息和降级标识 |

## 二、模块依赖清单

### 1. 管理员与认证域（adminsvc 本地）

| 后台场景 | RPC 方法 | 数据/依赖 | 说明 |
| --- | --- | --- | --- |
| 注册/登录/登出 | `Register`/`Login`/`Logout` | `admin_user` + Redis 会话 | 首个管理员免 token 注册；后续注册/管理员管理仅超管 |
| 会话与信息 | `ValidateSession`/`Me`/`Menus` | Redis 会话 + `admin_operation_log` | 网关鉴权中间件每次都调用 `ValidateSession` |
| 管理员管理 | `ListAdmins`/`CreateAdmin`/`UpdateAdmin`/`SetAdminStatus`/`ResetAdminPassword` | `admin_user` | 仅超管；写操作落 `admin_operation_log` |
| 操作日志 | `ListOperationLogs` | `admin_operation_log` | 按管理员/模块/动作/目标筛选 |

### 2. 用户模块

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 用户列表 | `adminsvc.ListUsers -> usersvc.AdminListUsers` | 同步 RPC | `keyword,status,start_time,end_time,page,page_size` | `list{id,phone,nickname,real_name,status,created_at,...},total` | 已切 usersvc；失败返回错误/空列表 |
| 用户详情 | `adminsvc.GetUser -> usersvc.AdminGetUser` | 同步 RPC | `id,sensitive` | `id,phone,nickname,avatar_url,gender,real_name,id_card_no,status,created_at` | 默认脱敏；`sensitive=true` 仅超管/运营可查看并写审计 |
| 用户冻结/解冻 | `adminsvc.FreezeUser/UnfreezeUser -> usersvc.AdminFreezeUser/AdminUnfreezeUser` | 同步 RPC | `id,reason,remark,admin_id,ip` | `message` | 已切 usersvc，不直写用户表 |
| 用户订单历史 | `adminsvc.ListUserOrders -> ordersvc.ListOrders` | 同步 RPC | `user_id,status,page,page_size` | 订单分页 | 按用户维度查询订单 |
| 用户优惠券历史 | `adminsvc.ListUserCoupons -> usersvc.ListMyCoupons` | 同步 RPC | `user_id,status,page,page_size` | `list{user_coupon_id,coupon_id,name,type,status,...},total` | 后台子资源（`/users/{id}/coupons`）与全局入口（`/user-coupons`，必须带 `user_id`）共用 |

### 3. 司机模块 `driversvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 司机列表 | `adminsvc.ListDrivers -> driversvc.ListDrivers` | 同步 RPC | `keyword,status,page,page_size` | `list{id,phone,real_name,status,online_status,vehicle_id,plate_no,vehicle_status,certification_id,audit_status,created_at},total` | driversvc 聚合司机、车辆、认证与 Redis 在线状态 |
| 司机详情 | `adminsvc.GetDriver -> driversvc.GetDriver` | 同步 RPC | `id,sensitive` | `driver{...}` | 默认脱敏；`sensitive=true` 仅超管/运营可看明文并写审计 |
| 司机冻结/解冻 | `adminsvc.FreezeDriver/UnfreezeDriver -> driversvc.FreezeDriver/UnfreezeDriver` | 同步 RPC | `driver_id,reason,remark,admin_id,ip` | `message` | 冻结后写审计并 `pushsvc.SendNotice/SendPush` 通知，失败写 `admin_audit_outbox` |
| 资质审核列表/详情 | `adminsvc.ListDriverCertifications/GetDriverCertification -> driversvc.AdminListCertifications/AdminGetCertification` | 同步 RPC | `keyword,audit_status,start_time,end_time,page,page_size` | 认证分页/详情 | 已切 driversvc |
| 资质审核通过 | `adminsvc.ApproveDriverCertification -> driversvc.ApproveCertification` | 同步 RPC | `id,remark,admin_id,ip` | `message` | driversvc 事务更新认证、司机、车辆；adminsvc 写审计 |
| 资质审核驳回 | `adminsvc.RejectDriverCertification -> driversvc.RejectCertification` | 同步 RPC | `id,remark,admin_id,ip` | `message` | 驳回不激活司机和车辆；adminsvc 写审计 |
| 提现列表 | `adminsvc.ListDriverWithdrawals -> driversvc.AdminListWithdraws` | 同步 RPC | `status,driver_id,keyword,page,page_size` | 提现分页 | 已切 driversvc |
| 提现审核 | `adminsvc.HandleDriverWithdraw -> driversvc.AuditWithdraw` | 同步 RPC | `id,approve,remark,admin_id,ip` | `message` | approve=打款成功，reject=打款失败；adminsvc 写审计 |

### 4. 订单与派单模块 `ordersvc` / `dispatchsvc` / `locationsvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 订单列表 | `adminsvc.ListOrders -> ordersvc.ListOrders` | 同步 RPC | `keyword,status,user_id,driver_id,start_time,end_time,page,page_size` | 订单分页 | 已切 ordersvc，adminsvc 不直读订单表 |
| 订单详情 | `adminsvc.GetOrder -> ordersvc.GetOrder/ListOrderStatusLogs` | 同步 RPC | `id` | `order,status_logs,dispatch_records,price,payment,settlement` | 主信息优先，关联数据可为空 |
| 异常订单查询 | `adminsvc.ListAbnormalOrders` | 同步 RPC | `abnormal_type,keyword,user_id,driver_id,start_time,end_time,page,page_size` | 异常订单分页 | 按取消/支付/派单异常口径查询 |
| 订单轨迹 | `adminsvc.GetOrderTrack -> locationsvc.GetOrderTrack` | 同步 RPC | `order_id,start_time,end_time,limit` | `points{longitude,latitude,speed_kmh,direction,recorded_at}` | locationsvc 读既有 `ride_track_point` 表 |
| 取消订单 | `adminsvc.CancelOrder -> ordersvc.CancelOrder` | 同步 RPC | `order_id,reason,admin_id,ip,request_id` | `message` | 固定 `operator_type=admin` |
| 人工改派 | `adminsvc.RedispatchOrder -> ordersvc.RedispatchOrder` | 同步 RPC | `order_id,new_driver_id,reason,admin_id,ip,request_id` | `order_id,status,driver_id,message` | request_id 幂等；ordersvc 释放原司机并重新派单 |
| 订单退款 | `adminsvc.RefundOrder -> ordersvc.ForceRefundOrder` | 同步 RPC | `order_id,refund_amount_cents,reason,admin_id,ip,request_id` | `order_id,status,refund_cents,refund_no,message` | request_id 作 refund_no；失败/待重试写 `admin_refund_compensation_task`，由 `job` 补偿 |
| 退款补偿队列 | `adminsvc.ListRefundRetryTasks/RetryRefundTask` | 同步 RPC | `page,page_size` / `refund_no,admin_id,ip` | 补偿任务分页 | 实际重试由 `job.RetryRefundEvents/RunRefundCompensation` 执行 |

### 5. 计价支付与营销模块 `pricesvc` / `usersvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 优惠券模板管理 | `adminsvc.ListCoupons/CreateCoupon/UpdateCoupon/DisableCoupon` | 同步 RPC | 模板字段 | `id/message` | adminsvc 写 `coupon`，创建与 `admin_operation_log` 同事务 |
| 批量发券 | `adminsvc.IssueCoupon -> usersvc.AdminIssueCoupon` | 同步 RPC | `coupon_id,target_type,target_config,admin_id,ip` | `task_no,total_count,success_count,fail_count,status` | 用户券写入、库存、领取上限校验由 usersvc 负责，adminsvc 不再直写 `user_coupon`/`coupon`；本地事务写 `admin_coupon_issue_task`、`admin_coupon_publish_record`、操作日志 |
| 发券任务查询 | `adminsvc.ListCouponIssueTasks` | 同步 RPC | `coupon_id,status,start_time,end_time,page,page_size` | 任务分页 | 读 `admin_coupon_issue_task` |
| 计价规则管理 | `adminsvc -> pricesvc`（List/Get/Create/Update/Enable/DisablePriceRule） | 同步 RPC | 规则字段 | `list/detail/id/message` | pricesvc 负责 `price_rule` 读写与启停 |
| 活动配置 | `adminsvc.List/Create/Update/Publish/RollbackPromotionActivity` | 同步 RPC | 活动字段 | `list/message` | 更新 `promotion_activity.status` 并写结构化操作日志，回填发布范围/时间 |
| 用户统计/发券/冻结补偿 | 统计 `GetUserStatistics` 等 | 同步 RPC/本地聚合 | `start_time,end_time,city_code` | 指标结构体 | 基于业务表实时聚合 |

### 6. 基础设施与风控模块 `locationsvc` / `pushsvc` / `usersvc`

| 后台场景 | RPC/MQ 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 运力地图 | `adminsvc.GetCapacityMap -> driversvc.ListDrivers` | 同步 RPC | `status,online_status,limit` | `drivers{driver_id,lng,lat,...},total,available_count,...` | 实时只读快照 |
| 司机操作通知 | `adminsvc -> pushsvc.SendNotice/SendPush` | 同步 RPC + outbox 补偿 | `driver_id,title,content,biz_type` | `notice_id/success` | 冻结/风控冻结等成功后通知司机；失败写 `admin_audit_outbox`，`job.RetryAdminAuditOutbox` 每 30s 重试 |
| 风控命中处置 | `adminsvc.HandleRiskHitRecords` | 同步 RPC | `ids,action,reason,...` | `success_count,fail_count,work_order_ids,failure_reasons` | action 支持 `review_pass/add_blacklist/create_work_order/freeze`；freeze 联动 usersvc/driversvc |
| 黑名单/命中查询 | `adminsvc.ListBlacklists/AddBlacklist/ReleaseBlacklist/ListRiskHitRecords` | 同步 RPC | 筛选与写入字段 | 分页/`message` | 写 `blacklist`，读 `risk_blacklist_hit_record` |
| 工单管理 | `adminsvc.Create/List/Get/Act/BatchActWorkOrder(+Evidence)` | 同步 RPC | 工单/证据字段 | 工单对象/分页 | 本地表 `admin_complaint_work_order`、`admin_work_order_flow`、`admin_work_order_evidence` |
| 导出任务 | `adminsvc.Create/List/GetExportTask(+DownloadExport)` | 同步 RPC + goroutine | `export_type,filters,admin_id,ip` | `task_no,status,...` | 写 `admin_export_task`，goroutine 生成 CSV，HTTP 流式下载；支持五类导出 |
| 审计补偿查询 | `adminsvc.ListAdminAuditOutbox` | 同步 RPC | `status,module,action,target_id,page,page_size` | outbox 分页 | 读 `admin_audit_outbox` |

## 三、异步与补偿任务清单（job）

| 任务 | 频率 | 覆盖范围 |
| --- | --- | --- |
| `RetryAdminAuditOutbox` | 每 30 秒 | 重写审计日志、重放 `driversvc.FreezeDriver`、重试 `pushsvc.SendNotice/SendPush` |
| `RetryAdminDomainOutbox` | 每 10 秒 | 把 `admin_domain_outbox` 中处罚、退款、发券、活动、通知事件可靠投递 Kafka |
| `RetryRefundEvents` / `RunRefundCompensation` | 每 10 秒 | 消费退款补偿队列/`admin_refund_compensation_task`（pending/processing/retrying/success/manual_review/failed） |
| `TimeoutCancelOrders` | 每 1 分钟 | 超时未接单订单自动取消 |
| `RescheduleExpiredDispatches` | 每 30 秒 | 派单超时重派 |
| `RetryPendingDispatches` | 每 10 秒 | 派单失败补偿重试 |
| `CleanExpiredLocation` | 每 1 小时 | 清理过期位置数据 |
| `DailyReport` | 每日凌晨 1 点 | 生成统计报表 |

## 四、当前已落地接口与切换点

| 后台接口 | 当前实现 |
| --- | --- |
| `GET /admin/v1/users` | `api/admin -> adminsvc.ListUsers -> usersvc.AdminListUsers`（已切换，不再直读用户表） |
| `GET /admin/v1/users/{id}` | `api/admin -> adminsvc.GetUser -> usersvc.AdminGetUser` |
| `POST /admin/v1/users/{id}/freeze|unfreeze` | `api/admin -> adminsvc -> usersvc.AdminFreezeUser/AdminUnfreezeUser` |
| `GET /admin/v1/users/{id}/orders` | `api/admin -> adminsvc.ListUserOrders -> ordersvc.ListOrders` |
| `GET /admin/v1/users/{id}/coupons`、`GET /admin/v1/user-coupons` | `api/admin -> adminsvc.ListUserCoupons -> usersvc.ListMyCoupons`（全局入口必须带 `user_id`） |
| `GET /admin/v1/drivers`、`GET /admin/v1/drivers/{id}` | `api/admin -> adminsvc -> driversvc.ListDrivers/GetDriver` |
| `POST /admin/v1/drivers/{id}/freeze|unfreeze` | `api/admin -> adminsvc -> driversvc.FreezeDriver/UnfreezeDriver` + pushsvc 通知补偿 |
| `GET /admin/v1/driver-certifications`、`/driver-certifications/{id}` | `api/admin -> adminsvc -> driversvc.AdminListCertifications/AdminGetCertification` |
| `POST /admin/v1/driver-certifications/{id}/approve|reject` | `api/admin -> adminsvc -> driversvc.ApproveCertification/RejectCertification` |
| `GET /admin/v1/driver-withdrawals`、`POST /driver-withdrawals/{id}/approve|reject` | `api/admin -> adminsvc -> driversvc.AdminListWithdraws/AuditWithdraw` |
| `GET /admin/v1/orders`、`GET /admin/v1/orders/{id}` | `api/admin -> adminsvc -> ordersvc.ListOrders/GetOrder`（已切换） |
| `GET /admin/v1/orders/abnormal` | `api/admin -> adminsvc.ListAbnormalOrders` |
| `GET /admin/v1/orders/{id}/track` | `api/admin -> adminsvc.GetOrderTrack -> locationsvc.GetOrderTrack` |
| `POST /admin/v1/orders/{id}/cancel|redispatch|refund` | `api/admin -> adminsvc -> ordersvc.CancelOrder/RedispatchOrder/ForceRefundOrder` |
| `GET /admin/v1/refund-retry-tasks`、`POST /refund-retry-tasks/{refund_no}` | `api/admin -> adminsvc.ListRefundRetryTasks/RetryRefundTask` |
| `GET /admin/v1/coupons`、`POST /admin/v1/coupons`、`PUT /coupons/{id}`、`POST /coupons/{id}/disable` | `api/admin -> adminsvc`（写 `coupon`） |
| `POST /admin/v1/coupons/{id}/issue` | `api/admin -> adminsvc.IssueCoupon -> usersvc.AdminIssueCoupon`（同步写任务+发布记录） |
| `GET /admin/v1/coupon-issue-tasks` | `api/admin -> adminsvc.ListCouponIssueTasks` |
| 计价规则 `GET|POST /price-rules`、`GET|PUT|POST(enable/disable) /price-rules/{id}` | `api/admin -> adminsvc -> pricesvc` |
| 活动配置 `GET|POST /promotion-activities`、`PUT|POST(publish/rollback) /promotion-activities/{id}` | `api/admin -> adminsvc` |
| 统计 `GET /statistics/{overview,orders,drivers,revenue,coupons,users}` | `api/admin -> adminsvc` 实时聚合 |
| `GET /capacity/map` | `api/admin -> adminsvc.GetCapacityMap -> driversvc.ListDrivers` |
| 导出 `POST|GET /export-tasks`、`GET /export-tasks/{task_no}`、`GET /export-tasks/{task_no}/download` | `api/admin -> adminsvc`，下载走 stream |
| 工单 `GET|POST /work-orders`、`GET /work-orders/{id}`、`POST /work-orders/{id}/actions`、`POST /work-orders/batch-actions`、`GET|POST /work-orders/{id}/evidence` | `api/admin -> adminsvc` |
| 风控 `GET|POST /blacklist`、`POST|PATCH /blacklist/{id}/release`、`GET /risk/hit-records`、`POST /risk/hit-records/actions` | `api/admin -> adminsvc` |
| `GET /notification-outbox` | `api/admin -> adminsvc.ListAdminAuditOutbox` |
| AI `POST /ai-agent/ask`、`GET /ai-agent/{suggestions,history}`、`POST /ai-agent/feedback`、`DELETE /ai-agent/conversations/{id}` | `api/admin -> adminsvc` |

## 五、事件与后续规划

| 事件名 | 生产方 | 消费方 | 触发场景 | 核心字段 |
| --- | --- | --- | --- | --- |
| 司机审核通过/驳回通知 | adminsvc（写 outbox） | `job.RetryAdminDomainOutbox` → Kafka → 通知消费者 | 审核结果通知 | `driver_id,certification_id,remark,...` |
| 退款补偿 | adminsvc/job | `job.RetryRefundEvents`/`RunRefundCompensation` | 退款失败、超时 | `refund_no,order_id,amount,reason` |
| 发券事件 | 后续规划：adminsvc → Kafka | usersvc/coupon 消费者 | 批量发券异步化 | `task_id,coupon_id,target_type,target_config` |
| 活动发布事件 | adminsvc → Kafka | 计价/通知消费者 | 活动发布/回滚 | `promotion_id,scope,version` |

> 历史版本中"司机审核失败写 `admin_audit_outbox`、job 每 30 秒 `RetryAdminAuditOutbox`"的描述仍然成立；在此基础上 2026-09 新增了 `admin_domain_outbox` 统一领域事件表与 `RetryAdminDomainOutbox`，用于向 Kafka 可靠投递处罚、退款、发券、活动、通知等事件。
