# 管理后台接口联调文档（ApiPost 版）

> 适用范围：`api/admin`（模块三：管理后台）
>
> 基础路径：`/admin/v1`，服务地址：`http://127.0.0.1:8717`
>
> 生成日期：2026-08-20（对齐当前 `api/admin/internal/handler/router.go`、`rpc/adminsvc/admin.proto` 与全量路由冒烟测试）
>
> 配套文件：[admin_openapi.json](admin_openapi.json)（可直接导入 ApiPost）

---

## 一、运行流程（先看这里）

### 1. 前置环境

1. 启动 Docker Desktop，确认以下容器在运行：
   - `redis`：`127.0.0.1:6379`（管理后台会话缓存，必须）
   - `mysql8`：`127.0.0.1:3306`（本机 MySQL 容器）
   - 远程 MySQL：`115.191.16.159:3306/xiaolong_ridy`（**后台业务库**，配置文件中 DSN 指向这里，需本机网络可达）
2. Go 工具链可用（本项目 `go.mod` 声明 `go 1.25.0`）。

### 2. 一键启动服务栈

在项目根目录打开 PowerShell：

```powershell
# 首次启动（自动编译，耗时约 1 分钟）
.\scripts\admin-test\start_admin_stack.ps1

# 之后可跳过编译
.\scripts\admin-test\start_admin_stack.ps1 -SkipBuild
```

启动内容：

| 服务 | 地址 | 说明 |
| --- | --- | --- |
| `api/admin` | `127.0.0.1:8717` | HTTP 网关（ApiPost 只调这个） |
| `rpc/adminsvc` | `127.0.0.1:8084` | 后台业务 RPC |
| `rpc/driversvc` | `127.0.0.1:8080` | 司机认证审核 RPC |
| `rpc/ordersvc` | `127.0.0.1:50051` | 后台取消订单下游 RPC |

验证服务可用：

```powershell
Invoke-RestMethod http://127.0.0.1:8717/healthz
# 期望 {"code":0,"message":"ok"}
```

停止服务栈：

```powershell
.\scripts\admin-test\start_admin_stack.ps1 -Stop
```

### 3. 手动启动（不依赖脚本）

```powershell
cd C:\Users\21848\Desktop\hxl\XiaoLong-Ridy
$env:GOMODCACHE = "C:\Users\21848\Desktop\hxl\XiaoLong-Ridy\.gotmp\pkg\mod"
$env:GOCACHE = "C:\Users\21848\Desktop\hxl\XiaoLong-Ridy\.gotmp\gocache"

# 1) driversvc
go build -o .gotmp\bin\driversvc.exe .\rpc\driversvc\driversvc.go
Start-Process .gotmp\bin\driversvc.exe -ArgumentList @("-f",".gotmp\driversvc-admin-test.yaml") -WorkingDirectory "C:\Users\21848\Desktop\hxl\XiaoLong-Ridy" -WindowStyle Hidden

# 2) adminsvc（使用不含 etcd 的测试配置，由 start_admin_stack.ps1 生成 .gotmp\adminsvc-admin-test.yaml）
go build -o .gotmp\bin\adminsvc.exe .\rpc\adminsvc\admin.go
Start-Process .gotmp\bin\adminsvc.exe -ArgumentList @("-f",".gotmp\adminsvc-admin-test.yaml") -WorkingDirectory "C:\Users\21848\Desktop\hxl\XiaoLong-Ridy" -WindowStyle Hidden

# 3) api/admin（工作目录必须是 api\admin，它按相对路径读 etc\admin.json）
cd api\admin
go build -o ..\..\.gotmp\bin\admin-api.exe .
Start-Process ..\..\.gotmp\bin\admin-api.exe -WorkingDirectory "C:\Users\21848\Desktop\hxl\XiaoLong-Ridy\api\admin" -WindowStyle Hidden
```

### 4. 数据库现状与账号

- 远程库已存在管理员：`admin / 123456`（role=1 超管）。若库被清空，`POST /admin/v1/auth/register` 可创建首个管理员（免 token）。
- 测试期间产生的数据：`AUTOTEST_*` 优惠券（已下架）、活动（已回滚）、黑名单（已解除）。
- 数据库迁移需按发布流程执行，代码不会自动修改线上数据库。导出任务依赖 `scripts/sql/migrate/09_admin_export_audit_task.sql` 中的 `admin_export_task` 表。

### 5. ApiPost 使用步骤

1. ApiPost 左侧“接口管理” → 导入 → 选择 **OpenAPI**，导入 `docs/api/admin_openapi.json`，自动生成全部接口目录。
2. 新建环境变量：
   - `baseUrl = http://127.0.0.1:8717`
   - `token = <登录接口返回的 token>`
3. 全局 Header：`Authorization: Bearer {{token}}`、`Content-Type: application/json`（登录/注册接口不需要 Authorization）。
4. 先调 `POST /admin/v1/auth/login`（body `{"username":"admin","password":"123456"}`），把返回的 `data.token` 填入环境变量 `token`，然后按模块逐个测试。
5. 每个接口都验证：HTTP 200 且 `code=0`。

---

## 二、通用规范

### 1. 鉴权

除 `POST /auth/register`、`POST /auth/login` 外，所有接口必须携带：

```http
Authorization: Bearer <token>
Content-Type: application/json
```

### 2. 统一响应

```json
{ "code": 0, "message": "ok", "data": { } }
```

### 3. 分页参数

| 参数 | 类型 | 必填 | 默认 | 说明 |
| --- | --- | --- | --- | --- |
| `page` | int | 否 | 1 | 页码，从 1 开始 |
| `page_size` | int | 否 | 20 | 每页条数，服务端最大钳制为 100 |

分页响应 `data`：

```json
{ "list": [], "total": 0, "page": 1, "page_size": 20 }
```

### 4. 错误码

| code | HTTP | 说明 |
| --- | --- | --- |
| `0` | 200 | 成功 |
| `40001` | 400 | 参数错误 / 方法不允许 |
| `40003` | 403 | 无权限 |
| `40004` | 401 | token 缺失或失效 |
| `40401` | 404 | 资源不存在 |
| `40902` | 409 | 冲突 |
| `50000` | 500 | 系统异常 |

### 5. 常用枚举

| 枚举 | 取值 |
| --- | --- |
| 管理员角色 `role` | 1 超管，2 运营，3 客服 |
| 用户状态 `status` | 1 正常，2 冻结 |
| 认证审核 `audit_status` | 1 待审核，2 通过，3 驳回 |
| 司机/车辆状态 | 1 待审核，2 正常 |
| 订单状态 `status` | 1 待接单，2 已接单，3 行程中，4 待支付，5 已完成，6 已取消 |
| 优惠券类型 `type` | 1 满减券，2 折扣券，3 立减券 |
| 优惠券状态 `status` | 1 草稿，2 启用，3 停用 |
| 活动类型 `type` | 1 满减，2 折扣，3 立减（与优惠券口径一致） |
| 活动状态 `status` | 1 草稿/未开始，2 运行中，3 已结束 |
| 黑名单 `target_type` | `user` / `driver` / `device` / `phone` |
| 黑名单状态 `status` | 1 生效，2 已解除 |
| 发券任务 `status` | `pending` / `running` / `success` / `partial_failed` / `failed` |

时间格式统一为 `YYYY-MM-DD HH:mm:ss`（本地时区）。

---

## 三、接口总览（60 个已登记条目）

| # | 模块 | 方法 | 路由 | 鉴权 |
| --- | --- | --- | --- | --- |
| 1 | 基础 | GET | `/healthz` | 否 |
| 2 | 鉴权 | POST | `/admin/v1/auth/register` | 首个免 token |
| 3 | 鉴权 | POST | `/admin/v1/auth/login` | 否 |
| 4 | 鉴权 | POST | `/admin/v1/auth/logout` | 是 |
| 5 | 鉴权 | GET | `/admin/v1/auth/me` | 是 |
| 6 | 鉴权 | GET | `/admin/v1/menus` | 是 |
| 7 | 管理员 | GET | `/admin/v1/admins` | 仅超管 |
| 8 | 管理员 | POST | `/admin/v1/admins` | 仅超管 |
| 9 | 管理员 | PUT | `/admin/v1/admins/{id}` | 仅超管 |
| 10 | 管理员 | POST | `/admin/v1/admins/{id}/status` | 仅超管 |
| 11 | 管理员 | POST | `/admin/v1/admins/{id}/reset-password` | 仅超管 |
| 12 | 日志 | GET | `/admin/v1/operation-logs` | 是 |
| 13 | 用户 | GET | `/admin/v1/users` | 是 |
| 14 | 用户 | GET | `/admin/v1/users/{id}` | 是 |
| 15 | 用户 | GET | `/admin/v1/users/{id}/orders` | 是 |
| 16 | 用户 | GET | `/admin/v1/users/{id}/coupons` | 是 |
| 17 | 用户 | POST | `/admin/v1/users/{id}/freeze` | 是 |
| 18 | 用户 | POST | `/admin/v1/users/{id}/unfreeze` | 是 |
| 19 | 司机 | GET | `/admin/v1/driver-certifications` | 是 |
| 20 | 司机 | GET | `/admin/v1/driver-certifications/{id}` | 是 |
| 21 | 司机 | POST | `/admin/v1/driver-certifications/{id}/approve` | 是 |
| 22 | 司机 | POST | `/admin/v1/driver-certifications/{id}/reject` | 是 |
| 23 | 订单 | GET | `/admin/v1/orders` | 是 |
| 24 | 订单 | GET | `/admin/v1/orders/{id}` | 是 |
| 25 | 订单 | GET | `/admin/v1/orders/abnormal` | 是 |
| 26 | 订单 | POST | `/admin/v1/orders/{id}/cancel` | 是 |
| 27 | 优惠券 | GET | `/admin/v1/coupons` | 是 |
| 28 | 优惠券 | POST | `/admin/v1/coupons` | 是 |
| 29 | 优惠券 | PUT | `/admin/v1/coupons/{id}` | 是 |
| 30 | 优惠券 | POST | `/admin/v1/coupons/{id}/disable` | 是 |
| 31 | 优惠券 | POST | `/admin/v1/coupons/{id}/issue` | 是 |
| 32 | 优惠券 | GET | `/admin/v1/coupon-issue-tasks` | 是 |
| 33 | 计价规则 | GET | `/admin/v1/price-rules` | 是 |
| 34 | 计价规则 | POST | `/admin/v1/price-rules` | 是 |
| 35 | 计价规则 | GET | `/admin/v1/price-rules/{id}` | 是 |
| 36 | 计价规则 | PUT | `/admin/v1/price-rules/{id}` | 是 |
| 37 | 计价规则 | POST | `/admin/v1/price-rules/{id}/enable` | 是 |
| 38 | 计价规则 | POST | `/admin/v1/price-rules/{id}/disable` | 是 |
| 39 | 营销 | GET | `/admin/v1/promotion-activities` | 是 |
| 40 | 营销 | POST | `/admin/v1/promotion-activities` | 是 |
| 41 | 营销 | PUT | `/admin/v1/promotion-activities/{id}` | 是 |
| 42 | 营销 | POST | `/admin/v1/promotion-activities/{id}/publish` | 是 |
| 43 | 营销 | POST | `/admin/v1/promotion-activities/{id}/rollback` | 是 |
| 44 | 统计 | GET | `/admin/v1/statistics/overview` | 是 |
| 45 | 统计 | GET | `/admin/v1/statistics/orders` | 是 |
| 46 | 统计 | GET | `/admin/v1/statistics/drivers` | 是 |
| 47 | 统计 | GET | `/admin/v1/statistics/revenue` | 是 |
| 48 | 统计 | GET | `/admin/v1/statistics/coupons` | 是 |
| 49 | 导出 | GET | `/admin/v1/export-tasks` | 是 |
| 50 | 导出 | POST | `/admin/v1/export-tasks` | 是 |
| 51 | 导出 | GET | `/admin/v1/export-tasks/{task_no}` | 是 |
| 52 | 导出 | GET | `/admin/v1/export-tasks/{task_no}/download` | 是 |
| 53 | 工单 | GET | `/admin/v1/work-orders` | 是 |
| 54 | 工单 | POST | `/admin/v1/work-orders` | 是 |
| 53 | 工单 | GET | `/admin/v1/work-orders/{id}` | 是 |
| 54 | 工单 | POST | `/admin/v1/work-orders/{id}/actions` | 是 |
| 55 | 工单 | GET/POST | `/admin/v1/work-orders/{id}/evidence` | 是 |
| 56 | 风控 | GET | `/admin/v1/blacklist` | 是 |
| 57 | 风控 | POST | `/admin/v1/blacklist` | 是 |
| 58 | 风控 | POST | `/admin/v1/blacklist/{id}/release` | 是 |
| 59 | 风控 | GET | `/admin/v1/risk/hit-records` | 是 |
| 60 | 基础 | GET | `/` | 否 |

---

## 四、鉴权接口

### 4.1 管理员注册

`POST /admin/v1/auth/register`

鉴权：首个管理员免 token；已有管理员后必须由超管携带 token 注册。

Body 参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 是 | 登录名 |
| `password` | string | 是 | 密码（服务端 bcrypt 加密存储） |
| `real_name` | string | 是 | 真实姓名 |
| `role` | int | 是 | 1 超管，2 运营，3 客服 |

请求示例：

```json
{ "username": "admin", "password": "123456", "real_name": "系统管理员", "role": 1 }
```

响应示例：

```json
{ "code": 0, "message": "ok", "data": { "token": "xxxx", "expires_in": 86400, "admin": { "id": 1, "username": "admin", "real_name": "系统管理员", "role": 1, "status": 1 } } }
```

### 4.2 管理员登录

`POST /admin/v1/auth/login`

Body 参数：`username`、`password`（必填）。

```json
{ "username": "admin", "password": "123456" }
```

成功返回同注册接口。错误密码返回 `401 / 40004`。

### 4.3 管理员退出

`POST /admin/v1/auth/logout`

鉴权：是。删除 Redis 会话，token 立即失效。

```json
{ "code": 0, "message": "ok", "data": { "message": "ok" } }
```

### 4.4 当前管理员信息

`GET /admin/v1/auth/me`

响应 `data.admin`：`id`、`username`、`real_name`、`role`、`status`。

### 4.5 菜单权限

`GET /admin/v1/menus`

响应 `data.items`：`name`、`path`、`icon`、`perm`、`children`。按角色返回。

### 4.6 管理员管理

以下接口仅允许 `role=1` 超级管理员调用：

- `GET /admin/v1/admins`：管理员分页列表，支持 `keyword`、`role`、`status`、`page`、`page_size`。
- `POST /admin/v1/admins`：创建管理员，必填 `username`、`password`、`real_name`、`role`、`status`。
- `PUT /admin/v1/admins/{id}`：编辑 `real_name`、`role`、`status`，密码通过单独接口修改。
- `POST /admin/v1/admins/{id}/status`：请求体 `{ "status": 1 }` 启用或 `{ "status": 2 }` 停用。
- `POST /admin/v1/admins/{id}/reset-password`：请求体 `{ "password": "654321" }` 重置密码。

管理员用户名必须唯一；当前登录的超级管理员不能被自己编辑、停用或重置密码。接口复用 `admin_user`、`admin_operation_log`，不修改数据库结构。

---

## 五、用户管理

### 5.1 用户列表

`GET /admin/v1/users`

Query 参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `keyword` | string | 否 | 手机号、昵称、真实姓名模糊匹配 |
| `status` | int | 否 | 1 正常，2 冻结 |
| `start_time` / `end_time` | string | 否 | 注册时间范围 |
| `page` / `page_size` | int | 否 | 分页 |

列表项字段：`id`、`phone`、`nickname`、`avatar_url`、`gender`、`real_name`、`id_card_no`、`register_source`、`status`、`created_at`、`updated_at`。

### 5.2 用户详情

`GET /admin/v1/users/{id}`

Path：`id`（int64）。不存在返回 `404 / 40401`。

### 5.3 冻结 / 解冻用户

`POST /admin/v1/users/{id}/freeze`、`POST /admin/v1/users/{id}/unfreeze`

Body：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `reason` | string | 否 | 原因 |
| `remark` | string | 否 | 补充说明 |

```json
{ "reason": "异常下单", "remark": "同设备高频取消订单" }
```

成功：`{"code":0,"message":"ok","data":{"message":"ok"}}`。不存在：`404 / 40401`。

### 5.4 用户订单历史

`GET /admin/v1/users/{id}/orders`

Query：`page`、`page_size`、`status`（订单状态，0 或空表示全部）。

处理链路：`api/admin → adminsvc.ListUserOrders → ordersvc.ListOrders`，HTTP 层不直接读取订单表。

### 5.5 用户优惠券历史

`GET /admin/v1/users/{id}/coupons`

Query：`page`、`page_size`、`status`（用户券状态：0 全部，1 未使用，2 已使用，3 已过期，4 锁定）。

处理链路：`api/admin → adminsvc.ListUserCoupons → usersvc.ListMyCoupons`，分页由 usersvc 下游查询完成。

---

## 六、司机审核

### 6.1 认证列表

`GET /admin/v1/driver-certifications`

Query：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `keyword` | string | 否 | 司机手机号、姓名、车牌号 |
| `audit_status` | int | 否 | 1 待审核，2 通过，3 驳回 |
| `start_time` / `end_time` | string | 否 | 提交时间范围 |
| `page` / `page_size` | int | 否 | 分页 |

列表项字段：`id`、`driver_id`、`vehicle_id`、`driver_phone`、`driver_name`、`driver_status`、`plate_no`、`vehicle_status`、`id_card_front_url`、`id_card_back_url`、`driver_license_url`、`vehicle_license_url`、`audit_status`、`audit_remark`、`audited_by`、`audited_at`、`created_at`、`updated_at`。

### 6.2 认证详情

`GET /admin/v1/driver-certifications/{id}`

不存在返回 `404 / 40401`。

### 6.3 审核通过

`POST /admin/v1/driver-certifications/{id}/approve`

Body：`remark`（string，可选，默认“审核通过”）。

```json
{ "remark": "资料齐全，审核通过" }
```

处理：driversvc 事务更新认证状态为 2，并联动 `driver.status=2`、`driver_vehicle.status=2`，adminsvc 写操作日志。

### 6.4 审核驳回

`POST /admin/v1/driver-certifications/{id}/reject`

Body：`remark`（string，可选，默认“资料不完整”）。

```json
{ "remark": "身份证照片不清晰，请重新上传" }
```

处理：认证状态置 3，不激活司机/车辆，写操作日志。

---

## 七、订单管理

### 7.1 订单列表

`GET /admin/v1/orders`

Query：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `keyword` | string | 否 | 订单号 |
| `status` | int | 否 | 1 待接单，2 已接单，3 行程中，4 待支付，5 已完成，6 已取消 |
| `user_id` / `driver_id` | int64 | 否 | 按用户/司机筛选 |
| `start_time` / `end_time` | string | 否 | 下单时间范围 |
| `page` / `page_size` | int | 否 | 分页 |

列表项字段：`id`、`order_no`、`user_id`、`driver_id`、`car_type`、`from_address`、`from_longitude`、`from_latitude`、`to_address`、`to_longitude`、`to_latitude`、`estimated_distance_m`、`estimated_duration_s`、`estimated_price`、`status`、`cancel_reason`、`cancel_by`、`created_at`、`updated_at`。

### 7.2 订单详情

`GET /admin/v1/orders/{id}`

返回聚合对象 `data`：

| 字段 | 说明 |
| --- | --- |
| `order` | 订单主信息 |
| `status_logs` | 状态流转日志（`id/order_id/from_status/to_status/operator_type/operator_id/remark/created_at`） |
| `dispatch_records` | 派单记录 |
| `price` | 价格明细（`order_price`，可能为空） |
| `payment` | 支付单（可能为空） |
| `settlement` | 结算单（可能为空） |

不存在返回 `404 / 40401`。

### 7.3 异常订单

`GET /admin/v1/orders/abnormal`

Query：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `abnormal_type` | string | 否 | `cancel` / `payment` / `dispatch`，空=全部 |
| `keyword` | string | 否 | 订单号 |
| `user_id` / `driver_id` | int64 | 否 | 筛选 |
| `start_time` / `end_time` | string | 否 | 时间范围 |
| `page` / `page_size` | int | 否 | 分页 |

列表项在订单字段外补充：`abnormal_type`、`abnormal_reason`、`payment_status`、`dispatch_status`。

异常判定：`cancel` = 状态 6 或取消原因非空；`payment` = 支付单状态 3/4；`dispatch` = 派单记录状态 3/4/5。

### 7.4 后台取消订单

`POST /admin/v1/orders/{id}/cancel`

Body：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `reason` | string | 否 | 取消原因，默认“后台取消订单” |

```json
{ "reason": "司机长时间未接单" }
```

处理链路：`api/admin → adminsvc.CancelOrder → ordersvc.CancelOrder(operator_type=admin)`。

说明：ordersvc 不可用时会按统一错误格式返回 `50000`，正常链路由 ordersvc 负责订单状态机校验和订单状态日志。

---

## 八、优惠券配置

### 8.1 模板列表

`GET /admin/v1/coupons`

Query：`keyword`（名称）、`type`（1/2/3）、`status`（1/2/3）、`start_time`/`end_time`（创建时间）、`page`/`page_size`。

列表项字段：`id`、`name`、`type`、`face_value`、`discount`、`threshold_amount`、`total_count`、`received_count`、`per_user_limit`、`valid_start_at`、`valid_end_at`、`status`、`created_at`、`updated_at`。

### 8.2 新增模板

`POST /admin/v1/coupons`

Body：

| 字段 | 类型 | 必填 | 校验 |
| --- | --- | --- | --- |
| `name` | string | 是 | 非空 |
| `type` | int | 是 | 1 满减，2 折扣，3 立减 |
| `face_value` | string | 是 | 金额字符串，如 `"8.00"` |
| `discount` | string | 是 | 折扣，如 `"1.00"` |
| `threshold_amount` | string | 是 | 门槛，如 `"0.00"` |
| `total_count` | int64 | 是 | 库存，0 表示不限量 |
| `per_user_limit` | int | 是 | 单用户限领，>0 |
| `valid_start_at` / `valid_end_at` | string | 是 | `YYYY-MM-DD HH:mm:ss`，结束必须晚于开始 |
| `status` | int | 是 | 1 草稿，2 启用，3 停用 |

```json
{
  "name": "新用户立减券",
  "type": 3,
  "face_value": "8.00",
  "discount": "1.00",
  "threshold_amount": "0.00",
  "total_count": 10000,
  "per_user_limit": 1,
  "valid_start_at": "2026-08-18 00:00:00",
  "valid_end_at": "2026-12-31 23:59:59",
  "status": 1
}
```

成功返回 `data.id`（新模板 ID），并写操作日志。

### 8.3 编辑模板

`PUT /admin/v1/coupons/{id}`

Body 同新增。不存在返回 `404 / 40401`；已发放的券不建议修改核心规则。

### 8.4 停用模板

`POST /admin/v1/coupons/{id}/disable`

无 Body。将状态置 3。不存在或已停用返回 `404 / 40401`。

### 8.5 发券

`POST /admin/v1/coupons/{id}/issue`

Body：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `target_type` | string | 是 | `user` / `batch` / `crowd` |
| `target_config` | string | 是 | JSON 字符串，如 `{"user_ids":[10001,10002]}` |

```json
{ "target_type": "user", "target_config": "{\"user_ids\":[10001,10002]}" }
```

前置条件：券状态为启用（2）且未过期。响应 `data`：`task_no`、`total_count`、`success_count`、`fail_count`、`status`。

说明：发券任务依赖 `admin_coupon_issue_task` 和 `user_coupon` 表，联调前需确认业务库已按数据库发布流程应用对应迁移。

### 8.6 发券任务列表

`GET /admin/v1/coupon-issue-tasks`

Query：`coupon_id`（int64）、`status`（`pending/running/success/partial_failed/failed` 或 1-5）、`start_time`/`end_time`、`page`/`page_size`。

列表项字段：`id`、`task_no`、`coupon_id`、`target_type`、`target_config`、`total_count`、`success_count`、`fail_count`、`status`、`failure_reason`、`operator_id`、`created_at`、`updated_at`。

说明：发券任务列表依赖 `admin_coupon_issue_task` 表，支持按优惠券、状态、时间范围分页查询。

---

## 九、计价规则

计价规则管理当前链路为 `api/admin -> adminsvc -> pricesvc`。后台 HTTP 层和 adminsvc 均不直接修改 `price_rule` 表。

### 9.1 规则列表

`GET /admin/v1/price-rules`

Query：`keyword`、`city_code`、`car_type`、`status`、`page`、`page_size`。

### 9.2 规则详情

`GET /admin/v1/price-rules/{id}`

### 9.3 新增规则

`POST /admin/v1/price-rules`

Body：

```json
{
  "name": "标准快车",
  "city_code": "110100",
  "car_type": 1,
  "base_price": "12.50",
  "base_distance_km": "3.00",
  "per_km_price": "2.40",
  "per_minute_price": "0.50",
  "night_start_time": "22:00:00",
  "night_end_time": "06:00:00",
  "night_surcharge": "1.20",
  "dynamic_max_factor": "2.00",
  "status": 1,
  "effective_at": "2026-08-20 00:00:00",
  "expire_at": "2026-12-31 23:59:59"
}
```

成功返回 `data.id`（由 `pricesvc` 创建并经 `adminsvc` 透传的真实规则 ID）。规则已在 pricesvc 生效后，即使即时审计写入失败，也会写入 `admin_audit_outbox` 补偿任务并保持创建成功语义。

### 9.4 编辑规则

`PUT /admin/v1/price-rules/{id}`，Body 同新增。

### 9.5 启用 / 停用规则

`POST /admin/v1/price-rules/{id}/enable`

`POST /admin/v1/price-rules/{id}/disable`

---

## 十、营销活动

### 10.1 活动列表

`GET /admin/v1/promotion-activities`

Query：`keyword`（名称）、`type`（1/2/3）、`status`（1/2/3）、`start_time`/`end_time`、`page`/`page_size`。

列表项字段：`id`、`name`、`type`、`config`、`start_at`、`end_at`、`status`、`created_by`、`created_at`、`updated_at`。

### 10.2 新增活动

`POST /admin/v1/promotion-activities`

Body：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `name` | string | 是 | 活动名称 |
| `type` | int | 是 | 1/2/3 |
| `config` | string | 是 | 合法 JSON 字符串，如 `{"city_code":"110000","discount":"8.00"}` |
| `start_at` / `end_at` | string | 是 | 时间范围，结束晚于开始 |
| `status` | int | 是 | 1 草稿/未开始，2 运行中，3 已结束 |

```json
{
  "name": "暑期满减活动",
  "type": 3,
  "config": "{\"city_code\":\"110000\",\"discount\":\"8.00\"}",
  "start_at": "2026-08-18 00:00:00",
  "end_at": "2026-08-31 23:59:59",
  "status": 1
}
```

### 10.3 编辑活动

`PUT /admin/v1/promotion-activities/{id}`，Body 同新增。不存在返回 `404 / 40401`。

### 10.4 发布 / 回滚

`POST /admin/v1/promotion-activities/{id}/publish`、`POST /admin/v1/promotion-activities/{id}/rollback`

Body：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `publish_scope` | string | 否 | 发布范围（publish 使用） |
| `target_config` | string | 否 | 目标配置 JSON 字符串 |

```json
{ "publish_scope": "all", "target_config": "{}" }
```

发布置状态 2，回滚置状态 1，并写操作日志。不存在返回 `404 / 40401`。

---

## 十一、数据统计

统一 Query：`start_time`、`end_time`、`city_code`。当前订单、支付、结算、价格明细、司机与优惠券相关表尚未保存权威 `city_code`，因此统计接口传入非空 `city_code` 会返回 `400 / 40001`；不得将其静默降级为全量统计。订单相关指标的时间范围按订单创建时间生效，支付、结算、补贴、司机提现等财务指标按各自业务记录创建时间生效。

### 11.1 运营总览

`GET /admin/v1/statistics/overview`

响应 `data`：`user_count`、`driver_count`、`order_count`、`completed_order_count`、`abnormal_order_count`、`gmv`（string）、`coupon_issue_count`、`blacklist_count`。

### 11.2 订单统计

`GET /admin/v1/statistics/orders`

响应 `data`：`order_count`、`completed_order_count`、`canceled_order_count`、`timeout_order_count`、`payment_abnormal_count`、`completion_rate`、`cancel_rate`。

### 11.3 司机统计

`GET /admin/v1/statistics/drivers`

响应 `data`：`driver_total`、`new_driver_count`、`pending_audit_count`、`approved_driver_count`、`completed_order_count`、`driver_income`（string）、`withdraw_pending_amount`（string）、`withdraw_success_amount`（string）、`withdraw_failed_count`、`average_score`（string）、`total_complaint_count`。

说明：当前数据库没有司机在线状态和独立接单事件表，因此不返回在线司机、接单司机等无法可靠计算的指标。

### 11.4 收入统计

`GET /admin/v1/statistics/revenue`

响应 `data`：`payment_order_count`、`paid_amount`（string）、`refund_order_count`、`refund_amount`（string）、`payment_failed_count`、`settlement_order_count`、`settlement_total_amount`（string）、`platform_commission`（string）、`driver_income`（string）、`platform_subsidy`（string）。

### 11.5 优惠券统计

`GET /admin/v1/statistics/coupons`

响应 `data`：`coupon_count`、`enabled_coupon_count`、`issued_coupon_count`、`used_coupon_count`、`expired_coupon_count`、`use_rate`。

---

## 十二、导出任务

### 12.1 创建导出任务

`POST /admin/v1/export-tasks`

Body：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `export_type` | string | 是 | 当前仅支持 `orders` |
| `filters` | string | 否 | JSON 字符串；只允许 `status`、`user_id`、`driver_id`、`start_time`、`end_time` |

```json
{ "export_type": "orders", "filters": "{\"start_time\":\"2026-08-01 00:00:00\"}" }
```

当前实现：写入 `admin_export_task` 独立任务表，响应 `data`：`task_no`、`status`、`message`。后台 goroutine 异步生成 CSV 文件，并回写 `file_path`、`failure_reason`、`updated_at`、`expires_at`。

`filters` 出现未知字段、类型不符、非法时间或 `city_code` 时返回 `400 / 40001`，不会退化为全量订单导出。

### 12.2 导出任务列表

`GET /admin/v1/export-tasks`

Query：`page`、`page_size`、`export_type`。

列表项字段：`task_no`、`export_type`、`filters`、`status`、`admin_id`、`created_at`、`file_path`、`failure_reason`、`updated_at`。

### 12.3 导出任务详情

`GET /admin/v1/export-tasks/{task_no}`

Path：`task_no`（string）。响应字段同列表项。

### 12.4 下载导出文件

`GET /admin/v1/export-tasks/{task_no}/download`

鉴权：创建者本人或超级管理员。返回 `text/csv; charset=utf-8` 文件流，响应头包含 `Content-Disposition: attachment; filename="{task_no}.csv"`。

前置条件：任务状态为成功、文件仍在有效期内、服务端文件名与任务号一致。未完成、过期、无权限或文件不存在分别按统一错误格式返回。

---

## 十三、工单中心

### 13.1 工单列表 / 创建

`GET /admin/v1/work-orders`

Query：`status`、`assignee_id`、`work_order_type`、`page`、`page_size`。

`POST /admin/v1/work-orders`

Body 字段：`work_order_type`、`source_type`、`source_id`、`order_id`、`user_id`、`driver_id`、`title`、`content`、`priority`。

### 13.2 工单详情

`GET /admin/v1/work-orders/{id}`

### 13.3 工单流转

`POST /admin/v1/work-orders/{id}/actions`

Body 字段：`action`（`assign/follow/arbitrate/close/reopen`）、`assignee_id`、`content`、`arbitration_result`、`version`。

角色限制：超管可执行全部动作；运营仅可 `assign/follow`；客服仅可 `follow`；`arbitrate/close/reopen` 仅超管。

### 13.4 工单批量处理

`POST /admin/v1/work-orders/batch-actions`

请求体：

```json
{
  "ids": [10001, 10002],
  "action": "follow",
  "content": "已完成批量跟进"
}
```

字段说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `ids` | int64[] | 是 | 工单 ID 列表 |
| `action` | string | 是 | `assign/follow/arbitrate/close/reopen` |
| `assignee_id` | int64 | 否 | `assign` 时的目标管理员 |
| `content` | string | 否 | 处理内容或批量备注 |
| `arbitration_result` | string | 否 | `arbitrate` 时的仲裁结论 |

运营可执行 `assign/follow`，客服可执行 `follow`，仲裁、结案、重开仅超管可执行。响应返回 `success_count`、`fail_count`、`failure_reasons`，批量项目逐条独立处理。

### 13.5 工单证据

`GET /admin/v1/work-orders/{id}/evidence`

`POST /admin/v1/work-orders/{id}/evidence`

Body 字段：`evidence_type`（`track/audio/chat/payment/image/text`）、`evidence_url`、`content`。

---

## 十四、风控黑名单

### 14.1 黑名单列表

`GET /admin/v1/blacklist`

Query：`target_type`（`user`/`driver`/`device`/`phone`）、`target_id`（int64）、`status`（1 生效，2 已解除）、`page`/`page_size`。

列表项字段：`id`、`target_type`、`target_id`、`reason`、`operator_id`、`status`、`created_at`、`updated_at`。

### 14.2 加入黑名单

`POST /admin/v1/blacklist`

Body：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `target_type` | string | 是 | `user`/`driver`/`device`/`phone` |
| `target_id` | int64 | 是 | 目标 ID |
| `reason` | string | 是 | 原因 |

```json
{ "target_type": "user", "target_id": 10001, "reason": "高频恶意取消订单" }
```

### 14.3 解除黑名单

`POST /admin/v1/blacklist/{id}/release`

Body：`reason`（string，可选）。

```json
{ "reason": "人工复核通过" }
```

仅对状态为生效（1）的记录生效，不存在或已解除返回 `404 / 40401`。

### 14.4 风控命中记录

`GET /admin/v1/risk/hit-records`

Query：`target_type`、`target_id`、`scene`（场景）、`risk_level`、`page`/`page_size`。

列表项字段：`id`、`blacklist_id`、`target_type`、`target_id`、`scene`、`risk_level`、`hit_reason`、`request_id`、`created_at`。

说明：风控命中记录依赖 `risk_blacklist_hit_record` 表，联调前需确认业务库已应用对应迁移。

运营可查询，客服无权查询。

### 14.5 风控命中处置

`POST /admin/v1/risk/hit-records/actions`

仅超级管理员可调用。

```json
{
  "ids": [10001, 10002],
  "action": "create_work_order",
  "reason": "高风险命中，转人工复核",
  "work_order_title": "风控命中复核",
  "priority": 3
}
```

`action` 支持 `review_pass`、`add_blacklist`、`create_work_order`。由于命中表没有处理状态字段，`review_pass` 只写操作日志；其余动作分别写入黑名单或工单及审计记录。响应返回 `success_count`、`fail_count`、`work_order_ids`、`failure_reasons`。

---

## 十五、操作日志

`GET /admin/v1/operation-logs`

Query：`admin_id`（int64）、`module`、`action`、`target_type`、`target_id`（int64）、`start_time`/`end_time`、`page`/`page_size`。

列表项字段：`id`、`admin_id`、`module`、`action`、`target_type`、`target_id`、`detail`、`ip`、`created_at`。

---

## 十六、注意事项

1. **数据库迁移注意**：`admin_coupon_issue_task`、`risk_blacklist_hit_record`、`admin_export_task` 等表需按数据库发布流程应用迁移脚本，代码不会自动修改线上数据库。
2. **写操作注意**：`freeze/unfreeze`、审核通过/驳回、发券、发布/回滚、拉黑、计价规则新增/编辑/启停等会真实改动业务库并写操作日志，测试时建议使用测试账号/可回滚数据。
3. **文档规划但未开放的路由**（返回 404）：`GET /drivers`、`GET /drivers/{id}`、`POST /drivers/{id}/freeze`、`GET /orders/{id}/track`、`POST /orders/{id}/redispatch`、`POST /orders/{id}/refund`、`GET /user-coupons`、`GET /dashboard/overview`、`GET /statistics/{revenue,drivers,users}`。
4. **自动化验证**：当前已通过 `go test ./api/admin/... ./rpc/adminsvc/... ./rpc/pricesvc/... ./rpc/driversvc/...`；路由 smoke 用例需继续补齐管理员管理、用户历史、导出下载和工单接口覆盖。
