# 管理后台全模块 RPC 依赖调用清单

> 适用范围：`api/admin` 管理后台。  
> 当前实现：`api/admin` 统一调用 `rpc/adminsvc`；`adminsvc` 负责数据库读写和必要的下游 RPC 调用。  
> 调用类型：同步 RPC 用于强一致查询和状态修改；异步 MQ/job 用于通知、补偿、批量任务和最终一致场景。

## 一、调用规范

| 规范项 | 要求 |
| --- | --- |
| 鉴权 | 后台 HTTP 层读取 Bearer Token，并调用 `adminsvc.ValidateSession` 校验登录态 |
| 幂等 | 审核、退款、发券、黑名单变更必须传 `request_id` 或业务幂等号 |
| 审计 | 所有敏感操作由 `adminsvc` 写入 `admin_operation_log`；司机审核审计失败时写 `admin_audit_outbox` 补偿记录 |
| 超时 | 查询类建议 1-3 秒，写操作建议 3-5 秒，批量任务走 MQ |
| 降级 | 订单详情等聚合查询允许部分模块超时后返回主信息和降级标识 |

## 二、五大模块依赖清单

### 1. 用户模块

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 用户列表 | 当前 `adminsvc.ListUsers` | 同步 RPC | `keyword,status,start_time,end_time,page,page_size` | `list{id,phone,nickname,real_name,status,created_at},total` | 返回空列表或提示查询失败 |
| 用户详情 | 当前 `adminsvc.GetUser` | 同步 RPC | `id` | `id,phone,nickname,avatar_url,gender,real_name,id_card_no,status,created_at` | 404 显示资源不存在 |
| 用户冻结 | 当前 `adminsvc.FreezeUser` | 同步 RPC | `id,reason,remark,admin_id,ip` | `message` | 失败时不返回假成功 |
| 用户解冻 | 当前 `adminsvc.UnfreezeUser` | 同步 RPC | `id,reason,remark,admin_id,ip` | `message` | 失败时不返回假成功 |
| 用户优惠券查询 | 未开放 HTTP 路由 | 后续规划 | `user_id,status,page,page_size` | `list{coupon_id,name,status,expire_at,used_at},total` | 当前后台未注册该路由；目标由 usersvc 提供后台查询 RPC，adminsvc 只做鉴权和适配 |

### 2. 司机模块 `driversvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 司机列表 | 未开放 HTTP 路由 | 后续规划 | `keyword,status,audit_status,page,page_size` | `list{id,phone,real_name,status,audit_status,score,created_at},total` | 目标接 driversvc 后台查询 RPC，adminsvc 不直接聚合司机表 |
| 司机详情 | 未开放 HTTP 路由 | 后续规划 | `driver_id` | `driver,vehicle,certification,service_stats` | 404 显示不存在 |
| 资质审核列表 | 当前 `adminsvc.ListDriverCertifications` | 同步 RPC | `keyword,audit_status,start_time,end_time,page,page_size` | `list{id,driver_id,vehicle_id,audit_status,submitted_at},total` | 失败提示重试 |
| 资质审核通过 | 当前 `adminsvc.ApproveDriverCertification -> driversvc.ApproveCertification` | 同步 RPC | `id,admin_id,remark,ip` | `message` | driversvc 事务更新认证、司机、车辆；adminsvc 写审计 |
| 资质审核驳回 | 当前 `adminsvc.RejectDriverCertification -> driversvc.RejectCertification` | 同步 RPC | `id,admin_id,remark,ip` | `message` | 驳回不激活司机和车辆；adminsvc 写审计 |
| 冻结司机 | 未开放 HTTP 路由 | 后续规划 | `driver_id,reason,operator_id,request_id` | `success,status,updated_at` | 失败时禁止写成功日志 |

### 3. 订单与派单模块 `ordersvc` / `dispatchsvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 订单列表 | 当前 `adminsvc.ListOrders` | 同步 RPC | `keyword,status,user_id,driver_id,start_time,end_time,page,page_size` | `list{id,order_no,user_id,driver_id,status,estimated_price,created_at},total` | 大范围查询限制分页和时间 |
| 订单详情 | 当前 `adminsvc.GetOrder` | 同步 RPC | `id` | `order,status_logs,dispatch_records,price,payment,settlement` | 主信息优先返回，关联数据可为空 |
| 异常订单查询 | 当前 `adminsvc.ListAbnormalOrders` | 同步 RPC | `abnormal_type,keyword,user_id,driver_id,start_time,end_time,page,page_size` | `list{order,abnormal_type,abnormal_reason,payment_status,dispatch_status},total` | 超时提示缩小筛选 |
| 订单轨迹 | 未开放 HTTP 路由 | 后续规划 | `order_id,start_time,end_time` | `points{lng,lat,speed,recorded_at}` | 后续接 locationsvc |
| 人工改派 | 未开放 HTTP 路由 | 后续规划 | `order_id,driver_id,reason,operator_id,request_id` | `dispatch_id,status,created_at` | 后续接 dispatchsvc |
| 取消订单 | 当前 `adminsvc.CancelOrder -> ordersvc.CancelOrder` | 同步 RPC | `order_id,reason,admin_id,ip` | `message` | 状态机拒绝则提示不可取消 |

### 4. 计价支付与营销模块

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 优惠券新增/编辑/下架 | 当前 `adminsvc.CreateCoupon/UpdateCoupon/DisableCoupon` | 同步 RPC | `id,name,type,face_value,discount,threshold_amount,total_count,per_user_limit,valid_start_at,valid_end_at,status,admin_id,ip` | `message`，创建后 HTTP 层会按名称回查新 ID | 参数非法直接返回 400 |
| 批量发券 | 当前 `adminsvc.IssueCoupon` | 同步 RPC | `coupon_id,target_type,target_config,admin_id,ip` | `task_no,total_count,success_count,fail_count,status` | 当前同步写任务和用户券，后续可拆 MQ/job |
| 计价规则管理 | 当前 `adminsvc -> pricesvc` | 同步 RPC | `id,name,city_code,car_type,base_price,base_distance_km,per_km_price,per_minute_price,status,effective_at,expire_at` | `list/detail/message` | pricesvc 负责 `price_rule` 读写，adminsvc 不直接改表 |
| 活动配置 | 当前 `adminsvc.List/Create/Update/Publish/RollbackPromotionActivity` | 同步 RPC | `id,name,type,config,start_at,end_at,status,publish_scope,target_config,admin_id,ip` | `list/message` | 当前更新活动状态并写操作日志 |
| 订单退款 | 未开放 HTTP 路由 | 后续规划 | `order_id,amount,reason,operator_id,request_id` | `refund_id,refund_no,status` | 当前后台未注册退款接口；后续可由 `adminsvc` 调用支付服务完成退款 |
| 支付记录查询 | 当前仅订单详情聚合读取支付区块 | 同步 RPC/本地聚合 | `order_id` | `payment_no,amount,channel,status,refund_amount,paid_at` | 支付区块异常时订单主信息优先返回 |

### 5. 基础设施与风控模块 `locationsvc` / `pushsvc` / `auditsvc`

| 后台场景 | RPC/MQ 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 轨迹查询 | `locationsvc.GetOrderTrack` | 同步 RPC | `order_id,start_time,end_time` | `points,total_distance_m` | 超时降级 |
| 审核通知司机 | `pushsvc.SendDriverNotice` | 异步 MQ | `driver_id,title,content,biz_type,biz_id` | `message_id,status` | MQ 重试，站内信兜底 |
| 投诉结果通知用户 | `pushsvc.SendUserNotice` | 异步 MQ | `user_id,title,content,biz_type,biz_id` | `message_id,status` | MQ 重试 |
| 写审计日志 | `auditsvc.WriteAuditLog` | 同步 RPC | `operator_id,module,action,target_type,target_id,before,after,request_id` | `audit_id,created_at` | 高危操作审计失败应阻断 |
| 风控命中上报 | `auditsvc.ReportRiskHit` | 异步 MQ | `target_type,target_id,scene,risk_level,hit_reason,request_id` | `accepted` | 异步补偿 |

## 三、当前已落地接口与后续切换点

| 后台接口 | 当前实现 | 后续 RPC 切换目标 |
| --- | --- | --- |
| `GET /admin/v1/users` | `api/admin -> adminsvc.ListUsers` | 当前仍由 adminsvc 直读用户表的过渡实现；目标为 `adminsvc -> usersvc.AdminListUsers` |
| `GET /admin/v1/driver-certifications` | `api/admin -> adminsvc.ListDriverCertifications` | 后续可切 `driversvc.ListCertifications` |
| `POST /admin/v1/driver-certifications/{id}/approve` | `api/admin -> adminsvc.ApproveDriverCertification -> driversvc.ApproveCertification` | 后续补 `pushsvc` MQ 通知司机端 |
| `POST /admin/v1/driver-certifications/{id}/reject` | `api/admin -> adminsvc.RejectDriverCertification -> driversvc.RejectCertification` | 后续补 `pushsvc` MQ 通知司机端 |
| `GET /admin/v1/orders` | `api/admin -> adminsvc.ListOrders` | 当前仍由 adminsvc 直读订单表的过渡实现；目标为 `adminsvc -> ordersvc.AdminListOrders` |
| `GET /admin/v1/orders/abnormal` | `api/admin -> adminsvc.ListAbnormalOrders` | 当前仍由 adminsvc 直读订单关联表的过渡实现；目标为 `ordersvc.AdminListAbnormalOrders` |
| `GET /admin/v1/coupons` | `api/admin -> adminsvc.ListCoupons` | 当前由 `adminsvc` 读 `coupon`；后续如拆营销服务，可再迁移 |
| `POST /admin/v1/coupons` | `api/admin -> adminsvc.CreateCoupon` | 当前由 `adminsvc` 写 `coupon` 并记录操作日志 |
| `PUT /admin/v1/coupons/{id}` | `api/admin -> adminsvc.UpdateCoupon` | 当前由 `adminsvc` 写 `coupon` 并记录操作日志 |
| `POST /admin/v1/coupons/{id}/issue` | `api/admin -> adminsvc.IssueCoupon`，同步写 `admin_coupon_issue_task/user_coupon` | 后续切 MQ/Job 异步批量发券 |
| `GET /admin/v1/price-rules` | `api/admin -> adminsvc.ListPriceRules -> pricesvc.ListPriceRules` | 已接线 pricesvc |
| `POST /admin/v1/price-rules` | `api/admin -> adminsvc.CreatePriceRule -> pricesvc.CreatePriceRule` | 已接线 pricesvc |
| `PUT /admin/v1/price-rules/{id}` | `api/admin -> adminsvc.UpdatePriceRule -> pricesvc.UpdatePriceRule` | 已接线 pricesvc |
| `POST /admin/v1/price-rules/{id}/enable` | `api/admin -> adminsvc.EnablePriceRule -> pricesvc.EnablePriceRule` | 已接线 pricesvc |
| `POST /admin/v1/price-rules/{id}/disable` | `api/admin -> adminsvc.DisablePriceRule -> pricesvc.DisablePriceRule` | 已接线 pricesvc |
| `GET /admin/v1/promotion-activities` | `api/admin -> adminsvc.ListPromotionActivities` | 后续联动 `pricesvc.CheckPromotionConflict` |
| `POST /admin/v1/promotion-activities/{id}/publish` | `api/admin -> adminsvc.PublishPromotionActivity` | 后续联动 `pricesvc.PublishPromotionRule` |
| `GET /admin/v1/statistics/overview` | `api/admin -> adminsvc.GetStatisticsOverview` | 后续独立 `reportsvc` |
| `POST /admin/v1/export-tasks` | `api/admin -> adminsvc.CreateExportTask`，写入 `admin_export_task` 并由 goroutine 生成 CSV | 后续对象存储下载 URL、重试/取消接口、过期清理 Job |
| `GET /admin/v1/blacklist` | `api/admin -> adminsvc.ListBlacklists` | 后续联动风控缓存刷新 |

## 四、异步消息清单

| 事件名 | 生产方 | 消费方 | 触发场景 | 核心字段 |
| --- | --- | --- | --- | --- |
| `DriverAuditApprovedEvent` | `driversvc/adminsvc` | `pushsvc` | 司机审核通过 | `driver_id,certification_id,operator_id,occurred_at` |
| `DriverAuditRejectedEvent` | `driversvc/adminsvc` | `pushsvc` | 司机审核驳回 | `driver_id,certification_id,remark,occurred_at` |
| `CouponIssueTaskEvent` | 后续规划：`adminsvc` | 后续规划：`job/coupon-worker` | 批量发券异步化 | `task_id,coupon_id,target_type,target_config` |
| `PromotionPublishedEvent` | `pricesvc` | `pushsvc/reportsvc` | 活动发布 | `promotion_id,version,scope,effective_at` |
| `RefundCompensationEvent` | 后续规划：支付服务/adminsvc | 后续规划：`job/refund-worker` | 退款资金不足或通道失败 | `refund_id,order_id,amount,reason` |
| `RiskBlacklistChangedEvent` | 后续规划：`adminsvc` | 后续规划：用户/司机/风控缓存消费者 | 黑名单新增或解除 | `blacklist_id,target_type,target_id,status` |

## 五、2026-08-20 当前落地修订

| 后台接口 | 当前实现 | 下游依赖 | 说明 |
| --- | --- | --- | --- |
| `POST /admin/v1/orders/{id}/cancel` | `api/admin` 做鉴权与参数转换，调用 `adminsvc.CancelOrder` | `ordersvc.CancelOrder` 同步 RPC | 已落地 P0，传入 `operator_type=admin`、`operator_id=admin_id`、`reason` |
| `POST /admin/v1/driver-certifications/{id}/approve` | `api/admin` 做鉴权与参数转换，调用 `adminsvc.ApproveDriverCertification` | `driversvc.ApproveCertification` 同步 RPC | 已切换 P0；driversvc 负责司机、车辆、认证状态事务更新 |
| `POST /admin/v1/driver-certifications/{id}/reject` | `api/admin` 做鉴权与参数转换，调用 `adminsvc.RejectDriverCertification` | `driversvc.RejectCertification` 同步 RPC | 已切换 P0；adminsvc 负责审核成功后的操作日志 |

P1/P2 当前已补基础可调用接口：优惠券发放任务、计价规则、活动配置、数据统计、导出任务、风控黑名单均已接入 `api/admin -> adminsvc`。其中计价规则由 `adminsvc` 转发到 `pricesvc`，批量发券仍为同步发放；导出任务已使用独立任务表和 goroutine 生成 CSV，后续可继续演进为 MQ/Job、对象存储和过期清理。
