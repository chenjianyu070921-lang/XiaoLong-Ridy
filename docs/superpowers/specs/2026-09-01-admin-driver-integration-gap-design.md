# 管理后台接入司机端缺口扫描与修复设计

## 1. 背景与目标

管理后台负责用户管理、司机管理、订单管理、营销活动、风控审核、数据统计等后台运营能力。司机端已完成司机注册、登录、车辆、资质、在线状态、接单与行程服务等主体开发。当前项目文档和代码已形成 `api/admin -> rpc/adminsvc -> rpc/driversvc/ordersvc/pushsvc` 的接入边界，本轮目标是在不改变数据库结构、不跨服务直连内部实现的前提下，对后台接入司机端能力做一次缺口扫描，并对发现的未对齐问题做最小范围修复。

本设计覆盖后台司机列表、司机详情、司机资质审核、司机冻结、订单司机关联、司机通知补偿、前端字段展示和自动化验证。所有新增 Go 代码必须添加中文注释，说明模块、结构体、函数、关键逻辑、入参和返回值的用途。

## 2. 设计原则

- 严格遵循 README 中的目录结构和服务边界：HTTP 网关放在 `api/admin`，后台业务编排放在 `rpc/adminsvc`，司机域数据和状态变更由 `rpc/driversvc` 负责。
- 后台不得直接调用司机端 HTTP 接口，也不得绕过 `driversvc` 直查或修改司机域表。
- 不新增数据库迁移，不通过代码改动数据库结构，不写入无关业务数据。
- 所有变更最小化、隔离化，只修复后台接入司机端相关缺口。
- 查询类接口保持只读语义，写操作必须保留权限校验、审计日志和失败返回，不返回假成功。
- 敏感字段默认脱敏，只有通过后台敏感字段权限校验后才返回完整手机号、身份证号和驾驶证号，并写入操作审计。

## 3. 当前已确认接入边界

`api/admin` 只负责后台 HTTP 接入、登录态校验、参数转换和统一响应包装。司机相关 HTTP 接口包括：

- `GET /admin/v1/drivers`
- `GET /admin/v1/drivers/{id}`
- `POST /admin/v1/drivers/{id}/freeze`
- `GET /admin/v1/driver-certifications`
- `GET /admin/v1/driver-certifications/{id}`
- `POST /admin/v1/driver-certifications/{id}/approve`
- `POST /admin/v1/driver-certifications/{id}/reject`

`rpc/adminsvc` 负责后台业务编排，不直接持有司机状态权威。司机列表、详情、冻结、审核通过和审核驳回应通过 `driversvc` RPC 完成；后台只做参数校验、权限校验、字段适配、审计日志和通知补偿。

`rpc/driversvc` 是司机域权威服务。`ListDrivers` 和 `GetDriver` 返回司机基础信息，并聚合车辆、资质和 Redis 在线状态；`FreezeDriver` 更新司机状态并同步在线状态；`ApproveCertification` 和 `RejectCertification` 在司机域内维护认证、司机和车辆状态。

订单行程状态机属于 `ordersvc`。司机端接单、到达、开始行程、结束行程通过 `api/driver` 调用 `ordersvc`，后台订单管理只展示和筛选司机关联信息，不复制司机端行程状态机逻辑。

## 4. 缺口扫描范围

### 4.1 司机列表与详情

扫描 `rpc/driversvc/internal/logic/list_drivers_logic.go`、`get_driver_logic.go`、`admin_driver_helpers.go`，确认后台列表和详情返回以下字段：

- 司机基础字段：`id`、`phone`、`real_name`、`id_card_no`、`driver_license_no`、`avatar_url`、`status`。
- 在线状态字段：`online_status`，以 Redis 在线状态为权威，Redis 无记录时视为离线。
- 车辆字段：`vehicle_id`、`plate_no`、`vehicle_status`。
- 认证字段：`certification_id`、`audit_status`、`audit_remark`。
- 时间字段：`created_at`、`updated_at`，由 `adminsvc` 统一转换为后台时间字符串。

如果发现 `driversvc` 聚合字段缺失，应优先在 `driversvc` 内补齐聚合逻辑，再由 `adminsvc` 做字段适配。不得让 `adminsvc` 直接查询司机、车辆或认证表。

### 4.2 敏感字段脱敏与审计

扫描 `rpc/adminsvc/internal/logic/adminservice/getdriverlogic.go` 和 `listdriverslogic.go`，确认司机列表默认脱敏，详情在 `sensitive=true` 时调用敏感字段权限校验。权限通过后返回完整字段，并写入 `admin_operation_log`，动作建议使用 `driver/view_sensitive`。

如果发现列表或详情绕过脱敏，应在 `adminsvc` 的适配层修复，不修改 `driversvc` 的权威数据返回。

### 4.3 司机资质审核

扫描 `approvedrivercertificationlogic.go`、`rejectdrivercertificationlogic.go`、`driversvc` 的审核逻辑和相关测试，确认：

- 审核通过调用 `driversvc.ApproveCertification`。
- 审核驳回调用 `driversvc.RejectCertification`。
- 审核通过由 `driversvc` 在本地事务内更新 `driver_certification`、`driver`、`driver_vehicle` 状态。
- 审核驳回只更新认证审核状态和备注，不激活司机和车辆。
- `adminsvc` 在司机域操作成功后写操作日志；审计失败不得返回业务假成功。

如果发现审核列表仍由 `adminsvc` 直读司机域表，本轮仅记录为后续迁移项；除非接口字段实际错误，否则不扩大到大规模迁移。

### 4.4 司机冻结与通知补偿

扫描 `freezedriverlogic.go`、`driversvc/internal/logic/freeze_driver_logic.go`、`pushsvc` 调用和 `admin_audit_outbox` 逻辑，确认：

- 后台冻结司机必须通过 `driversvc.FreezeDriver`。
- `driversvc` 将司机状态置为冻结，并将 Redis 在线状态置为离线。
- `adminsvc` 写操作日志，并通过 `pushsvc.SendNotice/SendPush` 通知司机。
- 通知失败写入 `admin_audit_outbox`，由 job 重试，司机冻结本身不因通知失败回滚。
- 冻结接口权限至少限制为超级管理员或符合现有后台授权规则的角色。

如果发现冻结只改后台本地表或未联动在线状态，应优先修复 RPC 调用链。

### 4.5 订单司机关联

扫描后台订单列表、详情、异常订单和订单轨迹逻辑，确认：

- 订单列表支持 `driver_id` 筛选。
- 订单详情展示 `driver_id`、派单记录、支付、结算和降级标识。
- 后台不实现司机端接单、到达、开始行程、结束行程接口。
- 订单轨迹通过 `locationsvc` 读取轨迹点，不从司机端接口取数。

如果发现后台页面缺少司机关联展示，仅修复 DTO 或前端展示；不改变订单状态机。

### 4.6 前端字段展示

扫描 `web/admin/src/api/modules.js`、`web/admin/src/views/driver/index.vue`、`web/admin/src/views/driver/certification.vue` 和订单详情页，确认前端只调用 `/admin/v1/...` 后台接口，展示字段与后台 DTO 对齐。

司机列表建议展示司机 ID、手机号、姓名、司机状态、在线状态、车牌、车辆状态、认证状态、注册时间。详情页可展示完整 DTO 中的非空字段。审核页应展示司机姓名、手机号、车牌、证件图片 URL、审核状态、审核备注、审核人和审核时间。

## 5. 修复策略

修复顺序按风险由低到高推进：

1. 先补只读字段映射和前端展示，因为该类变更无业务副作用。
2. 再补 `adminsvc` 与 `driversvc` 的 RPC 参数适配、脱敏和审计。
3. 最后补冻结、审核和通知补偿等写操作链路，确保失败时返回明确错误。

每个修复点都必须保留现有接口路径和响应结构。若发现文档与代码不一致，以 README 架构边界和真实 proto 字段为准，同时更新对应文档说明。

## 6. 测试设计

后端测试优先覆盖以下内容：

- `rpc/driversvc`：司机列表和详情聚合车辆、认证、在线状态；Redis 无在线记录时返回离线；聚合依赖异常时不影响司机主信息返回。
- `rpc/adminsvc`：司机列表调用 `driversvc.ListDrivers`；司机详情调用 `driversvc.GetDriver`；敏感字段权限不足返回错误；权限通过时写审计。
- `rpc/adminsvc`：审核通过和驳回转发到 `driversvc`，审计失败不返回假成功。
- `rpc/adminsvc`：冻结司机转发到 `driversvc`，通知失败写 outbox。
- `api/admin`：司机相关 HTTP 路由冒烟测试仍全部注册并返回统一结构。

前端验证优先运行管理后台现有构建或单元检查；如依赖未安装，则至少完成字段和接口路径静态核对，并在交付说明中明确未运行原因。

建议执行命令：

```powershell
go test ./rpc/driversvc/... ./rpc/adminsvc/... ./api/admin/...
```

如果前端依赖可用，再执行：

```powershell
cd web/admin
npm run build
```

## 7. 非目标

- 不新增或修改数据库迁移脚本。
- 不实现司机端新的 HTTP 接口。
- 不把司机端订单行程状态机复制到后台。
- 不重构管理后台前端整体布局和视觉风格。
- 不改动乘客端、支付、营销、风控等无关模块。
- 不调整线上中间件配置和真实业务数据。

## 8. 验收标准

- 后台司机列表和详情展示的司机、车辆、认证、在线状态字段与 `driversvc` 权威数据一致。
- 后台审核通过、审核驳回和司机冻结均通过 RPC 调用司机域服务完成，不存在后台直改司机域状态。
- 敏感字段默认脱敏，查看明文字段有权限校验和审计。
- 后台订单管理能按司机 ID 查询和展示司机关联信息，但不越界实现司机端行程动作。
- 相关 Go 代码具备中文注释，测试覆盖核心接入链路。
- 未修改数据库结构，未污染业务数据，未触碰无关模块。
