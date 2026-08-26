# 乘客端与订单真实数据接入设计

## 1. 文档目的

本文档用于指导管理后台模块接入乘客端和订单服务的真实业务数据。

本期范围仅包括：

- 乘客用户查询；
- 乘客订单查询；
- 订单详情聚合；
- 订单状态日志查询；
- 订单派单记录查询；
- 管理后台取消订单。

本期不包括司机管理、支付管理、营销活动、风控管理和数据统计的独立接入。

## 2. 当前系统事实

乘客端位于 `api/passenger` 和 `web/user`，当前真实业务链路如下：

```text
web/user
  -> api/passenger
  -> usersvc
  -> pricesvc
  -> ordersvc
  -> dispatchsvc
```

主要真实数据表和所属服务：

| 数据 | 真实来源 |
| --- | --- |
| 乘客基本资料 | `usersvc`，对应 `user` 表 |
| 订单主信息 | `ordersvc`，对应 `ride_order` 表 |
| 订单状态流转 | `ordersvc`，对应 `order_status_log` 表 |
| 派单记录 | `dispatchsvc`，对应 `dispatch_record` 表 |
| 价格预估 | `pricesvc` |

当前管理后台的用户列表和订单列表由 `adminsvc` 直接查询业务数据库，属于过渡实现。后续应通过后台专用 RPC 逐步解耦业务表结构。

## 3. 设计目标

### 3.1 业务目标

乘客端创建真实订单后，管理后台能够查询并展示：

- 乘客基本资料；
- 实名认证状态；
- 订单编号和订单状态；
- 起点、终点及经纬度；
- 车型、距离、时长和金额；
- 当前司机 ID；
- 订单状态日志；
- 派单记录；
- 取消原因、取消方和更新时间。

### 3.2 架构目标

统一保持以下边界：

```text
api/admin -> rpc/adminsvc -> 领域服务 RPC
```

其中：

- `api/admin` 只负责 HTTP 路由、管理员鉴权、参数解析和响应包装；
- `rpc/adminsvc` 负责后台业务编排、聚合、审计和错误转换；
- `usersvc` 负责用户领域数据；
- `ordersvc` 负责订单主数据和状态机；
- `dispatchsvc` 负责派单记录；
- 管理后台不得直接修改其他服务的业务表。

## 4. 总体架构

```text
web/admin
    |
    | HTTP /admin/v1/*
    v
api/admin
    |
    | AdminService RPC
    v
rpc/adminsvc
    |-- usersvc：用户查询
    |-- ordersvc：订单查询与取消
    |-- dispatchsvc：派单记录
    `-- 本地聚合：后台响应、审计、降级标识
```

一期允许保留现有数据库直读作为兼容路径，但必须通过内部接口隔离，不能在 HTTP 层新增直接 SQL。后续切换到领域 RPC 时，HTTP 接口和前端数据结构保持不变。

## 5. 后台接口范围

继续使用现有 HTTP 路径：

| HTTP 接口 | 用途 |
| --- | --- |
| `GET /admin/v1/users` | 乘客列表 |
| `GET /admin/v1/users/{id}` | 乘客详情 |
| `GET /admin/v1/orders` | 订单列表 |
| `GET /admin/v1/orders/{id}` | 订单详情 |
| `GET /admin/v1/orders/abnormal` | 异常订单列表 |
| `POST /admin/v1/orders/{id}/cancel` | 管理员取消订单 |

管理后台 HTTP 层不得复用乘客端 JWT。管理员必须使用后台自身会话和 RBAC 权限。

## 6. RPC 能力设计

### 6.1 用户查询

目标能力：

```text
usersvc.AdminListUsers
usersvc.AdminGetUser
usersvc.AdminGetUserHistory
```

查询字段：

- 用户 ID；
- 手机号；
- 昵称；
- 头像；
- 性别；
- 实名信息及认证状态；
- 账号状态；
- 注册来源；
- 创建时间；
- 更新时间。

敏感字段处理：

- 手机号、身份证号默认脱敏；
- 只有具备对应权限的管理员才允许查看完整实名信息；
- 后台服务负责脱敏，前端不自行处理；
- 用户不存在统一返回资源不存在错误。

### 6.2 订单查询

目标能力：

```text
ordersvc.AdminListOrders
ordersvc.AdminGetOrder
ordersvc.AdminListOrderStatusLogs
ordersvc.AdminCancelOrder
```

订单列表筛选：

- 订单号；
- 乘客 ID；
- 司机 ID；
- 订单状态；
- 创建时间范围；
- 页码；
- 每页数量。

订单详情主字段：

- 订单 ID；
- 订单号；
- 用户 ID；
- 司机 ID；
- 车型；
- 起终点地址；
- 起终点经纬度；
- 预计距离；
- 预计时长；
- 预计金额；
- 订单状态；
- 取消原因；
- 取消方；
- 创建时间；
- 更新时间。

### 6.3 派单查询

目标能力：

```text
dispatchsvc.AdminListDispatchRecords
```

派单记录字段：

- 记录 ID；
- 订单 ID；
- 司机 ID；
- 派单类型；
- 派单状态；
- 匹配分数；
- 备注；
- 创建时间；
- 更新时间。

## 7. 订单详情聚合

`adminsvc.GetOrder` 返回订单详情聚合结果：

```json
{
  "order": {},
  "user": {},
  "dispatch": {
    "currentDriverId": 0,
    "dispatchStatus": 0,
    "records": []
  },
  "statusLogs": [],
  "degraded": []
}
```

处理规则：

1. 订单主信息是必需数据，查询失败直接返回错误。
2. 用户摘要查询失败时，保留订单主信息并记录降级原因。
3. 派单记录查询失败时，保留订单主信息并记录降级原因。
4. 状态日志查询失败时，保留订单主信息并记录降级原因。
5. `degraded` 只描述关联数据缺失，不得伪造空业务结果。
6. 所有下游 RPC 使用独立超时，避免单个依赖拖垮整个详情接口。

## 8. 订单取消

管理员取消订单的调用链：

```text
api/admin
  -> adminsvc.CancelOrder
  -> ordersvc.AdminCancelOrder
  -> ordersvc 状态机校验
  -> 更新订单及状态日志
  -> adminsvc 写操作审计
```

请求必须包含：

- 订单 ID；
- 取消原因；
- 管理员 ID；
- 客户端 IP；
- `request_id` 或等价幂等号。

约束：

- `adminsvc` 不直接更新 `ride_order`；
- 订单状态是否允许取消由 `ordersvc` 决定；
- 下游失败时禁止返回业务成功；
- 审计失败时按现有后台审计策略处理；
- 重复请求必须具备幂等行为。

## 9. 运行配置与真实链路

乘客端联调必须明确使用真实 gRPC：

```powershell
$env:PASSENGER_CLIENT_MODE = "grpc"
```

并配置：

```powershell
$env:PASSENGER_USERSVC_ADDR
$env:PASSENGER_ORDERSVC_ADDR
$env:PASSENGER_PRICESVC_ADDR
$env:PASSENGER_DISPATCHSVC_ADDR
```

当前代码中的默认客户端模式需要在实现阶段重点核对，避免未配置环境变量时误用本地内存客户端。

乘客端实际接口以 `api/passenger/passenger.api` 和 `api/passenger/internal/router` 为准：

```text
/api/passenger/v1/profile/me
/api/passenger/v1/orders/create
/api/passenger/v1/orders/list
/api/passenger/v1/orders/detail
/api/passenger/v1/orders/status
/api/passenger/v1/orders/cancel
```

## 10. 错误处理与降级

错误分类：

| 场景 | 处理 |
| --- | --- |
| 参数非法 | HTTP 400 |
| 管理员未登录 | HTTP 401 |
| 无后台权限 | HTTP 403 |
| 用户或订单不存在 | HTTP 404 |
| 下游 RPC 不可用 | HTTP 502 |
| 订单状态不允许取消 | 返回业务错误，不伪造成功 |
| 关联详情超时 | 返回主信息并写入 `degraded` |

禁止：

- 下游失败时返回空列表并伪装成功；
- 通过后台直接修改订单状态；
- 使用乘客端本地内存订单作为后台真实数据；
- 在 HTTP 层直接读取业务数据库。

## 11. 数据一致性

一期采用以下策略：

- 用户和订单详情以领域服务返回为准；
- 订单状态以 `ordersvc` 状态机为准；
- 派单状态以 `dispatchsvc` 为准；
- 后台查询不自行推导或覆盖业务状态；
- 允许关联详情短暂不可用，但必须明确返回降级信息；
- 暂不引入独立统计读模型和新的数据库表。

## 12. 实施顺序

### 阶段一：真实链路基线

1. 确认乘客端使用真实 gRPC 模式。
2. 启动 `usersvc`、`ordersvc`、`pricesvc`、`dispatchsvc`。
3. 通过乘客端完成登录和真实下单。
4. 核对订单主表、订单状态和派单记录。

### 阶段二：后台查询接入

1. 抽象后台用户查询接口。
2. 抽象后台订单查询接口。
3. 接入订单详情聚合。
4. 接入状态日志和派单记录。
5. 将后台前端列表和详情页面切换到统一接口。

### 阶段三：后台操作接入

1. 保留并核对管理员取消订单链路。
2. 增加幂等号和请求超时。
3. 完善操作审计。
4. 验证订单状态机拒绝非法取消。

## 13. 验收标准

### 用户

- 乘客端登录成功后，后台用户列表可以查询到该乘客；
- 用户详情中的手机号、昵称和实名状态与 `usersvc` 一致；
- 用户不存在时返回明确的 404 业务错误。

### 订单

- 乘客端创建订单后，后台订单列表可以查询到订单；
- 订单号、用户 ID、金额、状态、起终点与 `ordersvc` 一致；
- 后台订单详情可以显示状态日志；
- 有派单记录时可以显示派单记录；
- 关联服务不可用时主订单仍可返回并包含降级信息。

### 取消

- 后台取消请求必须经过管理员鉴权；
- 取消结果由 `ordersvc` 状态机决定；
- 取消成功后乘客端查询到的订单状态同步变化；
- 非法状态取消不能返回成功；
- 重复取消请求具备幂等行为；
- 操作日志包含管理员、订单、原因和请求标识。

## 14. 非本期范围

- 司机管理和司机详情；
- 支付和退款；
- 优惠券和营销活动；
- 风控规则和黑名单；
- 运营统计读模型；
- 新增或修改数据库表结构；
- 乘客端支付参数统一改造。
