# 乘客端与订单真实数据接入实现计划

## 1. 目标

将管理后台的用户、订单、订单状态日志和派单记录接入乘客端真实业务链路，保持：

```text
web/user -> api/passenger -> usersvc/ordersvc/dispatchsvc
web/admin -> api/admin -> adminsvc -> 领域 RPC
```

本计划不包含司机管理、支付、营销、风控和统计改造。

## 2. 约束

- 不修改数据库结构；
- 不直接修改乘客端业务数据；
- `api/admin` 不新增直接 SQL；
- 订单状态只能由 `ordersvc` 状态机决定；
- 管理员鉴权、审计和幂等沿用现有后台规范；
- 现有用户和订单数据库直读作为兼容路径时，必须隔离在 `adminsvc` 内部；
- 所有新增 Go 代码添加中文模块、结构体、函数和关键逻辑注释；
- 变更只涉及用户、订单、派单和相关测试文档。

## 3. 阶段一：真实链路基线

### 3.1 核对乘客端运行模式

检查：

- `api/passenger/internal/svc/service_context.go`
- `api/passenger/etc/*`
- `docs/passenger/乘客端真实RPC联调记录.md`

任务：

1. 确认联调环境设置 `PASSENGER_CLIENT_MODE=grpc`。
2. 确认 usersvc、ordersvc、pricesvc、dispatchsvc 地址可配置。
3. 确认未配置真实服务时不会把验收环境静默切换到本地内存客户端。
4. 补充或修正配置测试，覆盖真实 gRPC 模式。

验收：

- 乘客端启动后使用真实 RPC client；
- 本地模式只能通过显式配置启用；
- 配置错误能明确暴露，而不是返回伪造业务数据。

### 3.2 核对乘客订单接口契约

检查：

- `api/passenger/passenger.api`
- `api/passenger/internal/router/router.go`
- `api/passenger/internal/logic/order_logic.go`
- `web/user/src/api/order.js`

任务：

1. 以实际 passenger API 路由作为联调标准。
2. 核对订单状态、金额、时间和 ID 类型。
3. 核对乘客端创建订单后 `ordersvc` 返回的数据。
4. 将发现的接口字段不一致记录为独立修复项，避免在后台适配层猜测字段。

验收：

- 真实乘客登录成功；
- 真实创建订单成功；
- 乘客端订单列表和详情能查到同一订单；
- 订单取消后状态可以重新查询到。

## 4. 阶段二：后台查询抽象

### 4.1 用户查询接口抽象

主要文件：

- `rpc/adminsvc/internal/logic/adminservice/listuserslogic.go`
- `rpc/adminsvc/internal/logic/adminservice/*user*`
- `rpc/adminsvc/internal/svc/service_context.go`
- `rpc/usersvc/proto/usersvc.proto`

任务：

1. 定义后台用户查询的内部接口。
2. 保持现有 HTTP 参数和返回结构不变。
3. 将当前用户数据库查询封装为兼容实现。
4. 预留 usersvc 后台 RPC adapter。
5. 对手机号、身份证号和实名信息执行服务端脱敏。
6. 保持用户状态和不存在错误的现有语义。

验收：

- `GET /admin/v1/users` 返回真实乘客；
- `GET /admin/v1/users/{id}` 与 usersvc 数据一致；
- 筛选、分页和状态过滤保持兼容；
- 不在 api/admin 中出现业务表 SQL。

### 4.2 订单查询接口抽象

主要文件：

- `rpc/adminsvc/internal/logic/adminservice/listorderslogic.go`
- `rpc/adminsvc/internal/logic/adminservice/order_helpers.go`
- `rpc/adminsvc/internal/logic/adminservice/*order*`
- `rpc/ordersvc/proto/ordersvc.proto`
- `rpc/dispatchsvc/proto/dispatchsvc.proto`

任务：

1. 定义后台订单列表、详情和状态日志查询接口。
2. 保留现有分页和筛选参数。
3. 将订单主信息查询与详情聚合拆分。
4. 将兼容数据库读取封装在 adminsvc repository 内。
5. 预留 ordersvc 和 dispatchsvc gRPC adapter。
6. 统一订单状态、金额和时间字段转换。

验收：

- `GET /admin/v1/orders` 可以查到乘客端真实订单；
- 订单号、用户 ID、地址、状态、金额与 ordersvc 一致；
- 状态筛选、用户筛选、时间筛选和分页有效；
- 不在 HTTP 层直接读 `ride_order`。

## 5. 阶段三：订单详情聚合

### 5.1 聚合结构

主要文件：

- `rpc/adminsvc/internal/logic/adminservice/getorderlogic.go`
- `rpc/adminsvc/adminsvc/admin.proto`
- `api/admin/internal/types/types.go`
- `api/admin/internal/handler/router.go`

任务：

1. 保持现有订单详情接口路径。
2. 设计订单主信息、用户摘要、派单记录、状态日志的聚合响应。
3. 订单主信息失败时直接返回错误。
4. 关联数据失败时返回主信息和降级标识。
5. 为每个下游调用设置独立超时。
6. 避免泄露乘客敏感信息。

验收：

- 订单详情包含主订单信息；
- 有派单记录时能展示；
- 有状态流转时能展示；
- dispatchsvc 或关联查询不可用时主订单仍能返回；
- 响应能明确说明被降级的模块。

### 5.2 异常订单

主要文件：

- `rpc/adminsvc/internal/logic/adminservice/listabnormalorderslogic.go`
- `api/admin/internal/handler/*`

任务：

1. 保留 `cancel`、`payment`、`dispatch` 类型筛选兼容性。
2. 一期只保证订单取消和派单异常的真实判断。
3. 支付异常字段只保留已有数据映射，不扩展 paysvc。
4. 对无法确认的异常不进行静默推断。

验收：

- 取消订单能被异常订单筛选到；
- 派单拒绝、超时、取消记录能正确展示；
- 不存在关联记录时不会生成伪异常。

## 6. 阶段四：后台取消订单

主要文件：

- `api/admin/internal/handler/*order*`
- `api/admin/internal/logic/order_logic.go`
- `rpc/adminsvc/internal/logic/adminservice/cancelorderlogic.go`
- `rpc/ordersvc/proto/ordersvc.proto`

任务：

1. 保持 `POST /admin/v1/orders/{id}/cancel`。
2. 校验管理员会话、订单 ID、取消原因和请求标识。
3. adminsvc 调用 ordersvc 后台取消能力。
4. 由 ordersvc 状态机判断是否允许取消。
5. 成功后写入后台操作日志。
6. 重复请求返回幂等结果，不重复变更订单。
7. 下游失败时禁止返回业务成功。

验收：

- 未登录或无权限不能取消；
- 可取消状态取消成功；
- 不可取消状态返回业务错误；
- 乘客端重新查询能看到最新状态；
- 审计日志包含管理员、订单、原因、IP 和 request_id。

## 7. 阶段五：测试与联调

### 7.1 单元测试

覆盖：

- gRPC 配置读取；
- 用户字段脱敏；
- 用户筛选和分页；
- 订单筛选和分页；
- 订单状态转换；
- 订单详情聚合；
- 关联服务降级；
- 取消幂等；
- 非法状态取消拦截。

### 7.2 集成测试

启动：

```text
usersvc
ordersvc
pricesvc
dispatchsvc
api/passenger
rpc/adminsvc
api/admin
```

流程：

1. 乘客发送短信验证码；
2. 乘客登录；
3. 乘客创建真实订单；
4. 后台查询用户列表；
5. 后台查询订单列表；
6. 后台查询订单详情；
7. 查询状态日志和派单记录；
8. 后台取消订单；
9. 乘客端重新查询订单状态；
10. 检查后台操作日志。

### 7.3 回归测试

执行：

```powershell
go test ./api/passenger/...
go test ./api/admin/...
go test ./rpc/adminsvc/...
go test ./rpc/ordersvc/...
go test ./rpc/dispatchsvc/...
```

禁止在本期执行会改变数据库结构或污染共享业务数据的迁移操作。

## 8. 交付顺序

1. 配置和真实 gRPC 基线；
2. 用户查询抽象；
3. 订单列表查询抽象；
4. 订单详情聚合；
5. 状态日志和派单记录；
6. 后台取消订单；
7. 单元测试；
8. 多服务联调；
9. 回归验证和文档更新。

每个阶段完成后先运行对应测试，再进入下一阶段。

## 9. 风险与处理

| 风险 | 处理 |
| --- | --- |
| 乘客端误用本地客户端 | 联调环境强制 `PASSENGER_CLIENT_MODE=grpc`，增加配置测试 |
| 后台与 ordersvc 状态不一致 | 订单状态以 ordersvc 为准，禁止后台直改 |
| 详情关联服务超时 | 主订单优先返回，使用 `degraded` 标识 |
| 旧数据库读取逻辑影响切换 | 通过 repository/adapter 隔离，逐步替换 |
| 字段文档与实际接口不一致 | 以 proto、`.api` 和实际 handler 为准 |
| 共享数据库被测试污染 | 使用只读查询、独立测试数据和显式清理策略 |
