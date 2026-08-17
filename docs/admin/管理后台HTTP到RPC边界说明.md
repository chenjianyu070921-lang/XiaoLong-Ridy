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
| `GET /admin/v1/operation-logs` | 解析日志筛选条件 | `AdminService.ListOperationLogs` |

## 3. 数据写入边界

| 操作 | 写入位置 | 说明 |
| --- | --- | --- |
| 司机审核通过 | `rpc/adminsvc` | 更新 `driver_certification`，同步更新司机和车辆状态为可用 |
| 司机审核驳回 | `rpc/adminsvc` | 更新 `driver_certification` 审核状态和驳回原因 |
| 优惠券新增/编辑/下架 | `rpc/adminsvc` | 写入或更新 `coupon` 表 |
| 用户冻结/解封 | `rpc/adminsvc` | 更新 `user.status` |
| 操作审计 | `rpc/adminsvc` | 写入 `admin_operation_log` |

## 4. 联调要求

1. 先启动 `rpc/adminsvc`，默认监听 `127.0.0.1:8080`。
2. 再启动 `api/admin`，默认读取 `api/admin/etc/admin.json` 中的 `admin_rpc.target`。
3. 前端和 Postman 仍然调用原 HTTP 路径，不直接调用 RPC。
4. MySQL 使用本地 Docker `3306`，Redis 使用本地 Docker `6379`。
