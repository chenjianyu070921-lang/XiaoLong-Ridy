# 管理后台全模块 RPC 依赖调用清单

> 适用范围：`api/admin` 管理后台。  
> 调用类型：同步 RPC 用于强一致查询和状态修改；异步 MQ 用于通知、补偿、批量任务和最终一致场景。

## 一、调用规范

| 规范项 | 要求 |
| --- | --- |
| 鉴权 | 后台 HTTP 层先校验管理员 Bearer Token，再发起 RPC |
| 幂等 | 审核、退款、发券、黑名单变更必须传 `request_id` 或业务幂等号 |
| 审计 | 所有敏感操作成功后写入 `admin_operation_log` |
| 超时 | 查询类建议 1-3 秒，写操作建议 3-5 秒，批量任务走 MQ |
| 降级 | 订单详情等聚合查询允许部分模块超时后返回主信息和降级标识 |

## 二、五大模块依赖清单

### 1. 用户模块 `usersvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 用户列表 | `usersvc.ListUsers` | 同步 RPC | `keyword,status,start_time,end_time,page,page_size` | `list{id,phone,nickname,real_name,status,created_at},total` | 返回空列表或提示查询失败 |
| 用户详情 | `usersvc.GetUserDetail` | 同步 RPC | `user_id` | `id,phone,nickname,avatar_url,gender,real_name,id_card_no,status,created_at` | 404 显示资源不存在 |
| 用户冻结 | `usersvc.FreezeUser` | 同步 RPC | `user_id,reason,operator_id,request_id` | `success,updated_at` | 幂等重试，失败写操作日志失败原因 |
| 用户解冻 | `usersvc.UnfreezeUser` | 同步 RPC | `user_id,reason,operator_id,request_id` | `success,updated_at` | 幂等重试 |
| 用户优惠券查询 | `usersvc.ListUserCoupons` 或 `couponsvc.ListUserCoupons` | 同步 RPC | `user_id,status,page,page_size` | `list{coupon_id,name,status,expire_at,used_at},total` | 降级隐藏优惠券区块 |

### 2. 司机模块 `driversvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 司机列表 | `driversvc.ListDrivers` | 同步 RPC | `keyword,status,audit_status,page,page_size` | `list{id,phone,real_name,status,audit_status,score,created_at},total` | 提示查询失败 |
| 司机详情 | `driversvc.GetDriverDetail` | 同步 RPC | `driver_id` | `driver,vehicle,certification,service_stats` | 404 显示不存在 |
| 资质审核列表 | `driversvc.ListCertifications` | 同步 RPC | `keyword,audit_status,start_time,end_time,page,page_size` | `list{id,driver_id,vehicle_id,audit_status,submitted_at},total` | 失败提示重试 |
| 资质审核通过 | `driversvc.ApproveCertification` | 同步 RPC | `certification_id,operator_id,remark,request_id` | `driver_id,vehicle_id,audit_status,updated_at` | 幂等校验，重复审核返回状态冲突 |
| 资质审核驳回 | `driversvc.RejectCertification` | 同步 RPC | `certification_id,operator_id,remark,request_id` | `driver_id,audit_status,updated_at` | 驳回原因必填 |
| 冻结司机 | `driversvc.FreezeDriver` | 同步 RPC | `driver_id,reason,operator_id,request_id` | `success,status,updated_at` | 失败时禁止写成功日志 |

### 3. 订单与派单模块 `ordersvc` / `dispatchsvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 订单列表 | `ordersvc.ListOrders` | 同步 RPC | `keyword,status,user_id,driver_id,start_time,end_time,page,page_size` | `list{id,order_no,user_id,driver_id,status,estimated_price,created_at},total` | 大范围查询限制分页和时间 |
| 订单详情 | `ordersvc.GetOrderDetail` | 同步 RPC | `order_id` | `order,status_logs,price,payment,settlement` | 主信息优先返回，关联数据可降级 |
| 异常订单查询 | `ordersvc.ListAbnormalOrders` | 同步 RPC | `abnormal_type,keyword,user_id,driver_id,start_time,end_time,page,page_size` | `list{order,abnormal_type,abnormal_reason,payment_status,dispatch_status},total` | 超时提示缩小筛选 |
| 订单轨迹 | `locationsvc.GetOrderTrack` | 同步 RPC | `order_id,start_time,end_time` | `points{lng,lat,speed,recorded_at}` | 超时隐藏轨迹区块 |
| 人工改派 | `dispatchsvc.ManualRedispatch` | 同步 RPC | `order_id,driver_id,reason,operator_id,request_id` | `dispatch_id,status,created_at` | 订单完结时返回状态冲突 |
| 取消订单 | `ordersvc.AdminCancelOrder` | 同步 RPC | `order_id,reason,operator_id,request_id` | `order_id,from_status,to_status,updated_at` | 状态机拒绝则提示不可取消 |

### 4. 计价支付与营销模块 `pricesvc` / `paysvc` / `couponsvc`

| 后台场景 | RPC 方法 | 调用类型 | 请求参数 | 返回字段 | 失败处理 |
| --- | --- | --- | --- | --- | --- |
| 优惠券新增/编辑 | `couponsvc.SaveCouponTemplate` | 同步 RPC | `id,name,type,face_value,discount,threshold_amount,total_count,per_user_limit,valid_start_at,valid_end_at,status` | `coupon_id,version,updated_at` | 参数非法直接返回 400 |
| 优惠券发布 | `pricesvc.PublishPromotionRule` | 同步 RPC | `coupon_id,publish_scope,target_config,version,operator_id,request_id` | `publish_id,status,effective_at` | 发布失败记录 `admin_coupon_publish_record` |
| 活动冲突检测 | `pricesvc.CheckPromotionConflict` | 同步 RPC | `city_code,crowd_scope,start_at,end_at,priority` | `has_conflict,conflicts[]` | 阻断发布 |
| 批量发券 | `couponsvc.CreateIssueTask` | 异步 MQ | `coupon_id,target_type,target_config,operator_id,request_id` | `task_id,task_no,status` | 任务失败可重试 |
| 订单退款 | `paysvc.AdminRefund` | 同步 RPC | `order_id,amount,reason,operator_id,request_id` | `refund_id,refund_no,status` | 资金不足转补偿任务 |
| 支付记录查询 | `paysvc.GetPaymentByOrder` | 同步 RPC | `order_id` | `payment_no,amount,channel,status,refund_amount,paid_at` | 支付区块降级 |

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
| `GET /admin/v1/users` | `api/admin` 直查 `user` 表 | `usersvc.ListUsers` |
| `GET /admin/v1/driver-certifications` | `api/admin` 直查司机审核表 | `driversvc.ListCertifications` |
| `POST /admin/v1/driver-certifications/{id}/approve` | 本地事务更新司机/车辆/审核表 | `driversvc.ApproveCertification` + `pushsvc` MQ |
| `GET /admin/v1/orders` | `api/admin` 直查 `ride_order` | `ordersvc.ListOrders` |
| `GET /admin/v1/orders/abnormal` | 本地聚合订单/支付/派单状态 | `ordersvc.ListAbnormalOrders` |
| `GET /admin/v1/coupons` | `api/admin` 直查 `coupon` | `couponsvc.ListCouponTemplates` |
| `POST /admin/v1/coupons` | 本地写 `coupon` + 操作日志 | `couponsvc.SaveCouponTemplate` |
| `PUT /admin/v1/coupons/{id}` | 本地更新 `coupon` + 操作日志 | `couponsvc.SaveCouponTemplate` |

## 四、异步消息清单

| 事件名 | 生产方 | 消费方 | 触发场景 | 核心字段 |
| --- | --- | --- | --- | --- |
| `DriverAuditApprovedEvent` | `driversvc/adminsvc` | `pushsvc` | 司机审核通过 | `driver_id,certification_id,operator_id,occurred_at` |
| `DriverAuditRejectedEvent` | `driversvc/adminsvc` | `pushsvc` | 司机审核驳回 | `driver_id,certification_id,remark,occurred_at` |
| `CouponIssueTaskEvent` | `couponsvc/adminsvc` | `job/coupon-worker` | 批量发券 | `task_id,coupon_id,target_type,target_config` |
| `PromotionPublishedEvent` | `pricesvc` | `pushsvc/reportsvc` | 活动发布 | `promotion_id,version,scope,effective_at` |
| `RefundCompensationEvent` | `paysvc/adminsvc` | `job/refund-worker` | 退款资金不足或通道失败 | `refund_id,order_id,amount,reason` |
| `RiskBlacklistChangedEvent` | `api/admin` | `usersvc/driversvc/risk-cache` | 黑名单新增或解除 | `blacklist_id,target_type,target_id,status` |

## 五、2026-08-15 当前落地修订

| 后台接口 | 当前实现 | 下游依赖 | 说明 |
| --- | --- | --- | --- |
| `POST /admin/v1/orders/{id}/cancel` | `api/admin` 做鉴权与参数转换，调用 `adminsvc.CancelOrder` | `ordersvc.CancelOrder` 同步 RPC | 已落地 P0，传入 `operator_type=admin`、`operator_id=admin_id`、`reason` |
| `POST /admin/v1/driver-certifications/{id}/approve` | 保留 `adminsvc` 既有本地事务审核逻辑 | 后续切换 `driversvc.ApproveCertification` | 本次不强制切换，避免影响司机模块同事迭代 |
| `POST /admin/v1/driver-certifications/{id}/reject` | 保留 `adminsvc` 既有本地事务审核逻辑 | 后续切换 `driversvc.RejectCertification` | `driversvc` 已预留 proto 与服务端逻辑，后续需补真实数据联调记录 |

P1/P2 仅保留设计文档与接口清单：优惠券发放任务、活动配置、数据统计、导出、风控管理暂不新增运行时代码。
