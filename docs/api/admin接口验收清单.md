# 管理后台接口验收清单

> 适用范围：`api/admin`
>
> 基础路径：`/admin/v1`
>
> 当前结论基于本地联调结果与现有路由代码。

## 1. 验收原则

1. 除登录、注册外，所有接口都必须携带 `Authorization: Bearer <token>`。
2. 所有敏感操作必须写入 `admin_operation_log`。
3. 仅验证接口行为，不做破坏性数据修改。
4. 验收结果分为：
   - `已通过`
   - `参数校验正常`
   - `资源不存在`
   - `待实现`

## 2. 已实现接口验收清单

### 2.1 鉴权类

| 接口 | 方法 | 状态 | 验收结果 |
| --- | --- | ---: | --- |
| `/admin/v1/auth/register` | POST | 已通过 | 无效请求体返回 `400`，注册链路可用 |
| `/admin/v1/auth/login` | POST | 已通过 | `admin/123456` 登录成功，返回 token |
| `/admin/v1/auth/logout` | POST | 已通过 | 登录态可正常注销 |
| `/admin/v1/auth/me` | GET | 已通过 | 返回当前管理员信息 |
| `/admin/v1/menus` | GET | 已通过 | 返回角色菜单权限 |

### 2.2 操作日志

| 接口 | 方法 | 状态 | 验收结果 |
| --- | --- | ---: | --- |
| `/admin/v1/operation-logs` | GET | 已通过 | 可正常查询日志列表 |

### 2.3 用户管理

| 接口 | 方法 | 状态 | 验收结果 |
| --- | --- | ---: | --- |
| `/admin/v1/users` | GET | 已通过 | 列表可查 |
| `/admin/v1/users/{id}` | GET | 已通过 | 当前测试库无数据，返回 `404` |
| `/admin/v1/users/{id}/freeze` | POST | 已通过 | `999999` 返回 `404`，错误映射正常 |
| `/admin/v1/users/{id}/unfreeze` | POST | 已通过 | `999999` 返回 `404`，错误映射正常 |

### 2.4 司机审核

| 接口 | 方法 | 状态 | 验收结果 |
| --- | --- | ---: | --- |
| `/admin/v1/driver-certifications` | GET | 已通过 | 列表可查 |
| `/admin/v1/driver-certifications/{id}` | GET | 已通过 | 当前测试库无数据，返回 `404` |
| `/admin/v1/driver-certifications/{id}/approve` | POST | 已通过 | `999999` 返回 `400`，参数/前置校验正常 |
| `/admin/v1/driver-certifications/{id}/reject` | POST | 已通过 | `999999` 返回 `400`，参数/前置校验正常 |

### 2.5 订单处理

| 接口 | 方法 | 状态 | 验收结果 |
| --- | --- | ---: | --- |
| `/admin/v1/orders` | GET | 已通过 | 列表可查 |
| `/admin/v1/orders/{id}` | GET | 已通过 | 当前测试库无数据，返回 `404` |
| `/admin/v1/orders/abnormal` | GET | 已通过 | 支持异常订单查询 |

### 2.6 优惠券配置

| 接口 | 方法 | 状态 | 验收结果 |
| --- | --- | ---: | --- |
| `/admin/v1/coupons` | GET | 已通过 | 列表可查 |
| `/admin/v1/coupons` | POST | 待补充正式样本 | 当前未做有效新增样本验证 |
| `/admin/v1/coupons/{id}` | PUT | 已通过 | `4` 返回 `400`，请求体校验正常 |
| `/admin/v1/coupons/{id}/disable` | POST | 已通过 | `999999` 返回 `404`，错误映射正常 |

## 3. 当前验收结论

### 3.1 已确认通过

- 登录鉴权链路正常。
- token 校验正常。
- 用户、司机、订单、优惠券、日志的列表接口可用。
- 详情、冻结、审核、下架等路由均已接通。
- `404` / `400` 的错误映射符合预期。
- `api/admin` 到 `rpc/adminsvc` 的边界已生效。
- `go test ./api/admin/... ./rpc/adminsvc/...` 已通过。

### 3.2 当前限制

- 本地测试库数据量较少，部分详情接口只能验证 `404`。
- `POST /admin/v1/coupons` 需要正式样本做完整新增验收。
- 司机审核、用户冻结、优惠券下架等改写类接口建议后续补充一条可回滚测试数据。

## 4. 文档规划但当前未验收接口

以下接口在设计文档中存在，但当前路由未必已完整开放，建议标记为 `待实现`。

| 模块 | 接口 | 状态 |
| --- | --- | ---: |
| 用户管理 | `/admin/v1/users/{id}/orders` | 待实现 |
| 用户管理 | `/admin/v1/users/{id}/coupons` | 待实现 |
| 司机管理 | `/admin/v1/drivers` | 待实现 |
| 司机管理 | `/admin/v1/drivers/{id}/freeze` | 待实现 |
| 订单处理 | `/admin/v1/orders/{id}/track` | 待实现 |
| 订单处理 | `/admin/v1/orders/{id}/cancel` | 待实现 |
| 订单处理 | `/admin/v1/orders/{id}/redispatch` | 待实现 |
| 订单处理 | `/admin/v1/orders/{id}/refund` | 待实现 |
| 风控黑名单 | `/admin/v1/blacklist` | 待实现 |
| 风控黑名单 | `/admin/v1/blacklist/{id}/release` | 待实现 |
| 数据统计 | `/admin/v1/dashboard/overview` | 待实现 |
| 数据统计 | `/admin/v1/statistics/*` | 待实现 |
| 导出任务 | `/admin/v1/export-tasks` | 待实现 |

## 5. 验收标准

### 5.1 成功标准

- 返回 `200`。
- 响应体符合统一格式：
  - `code = 0`
  - `message = ok`
  - `data` 正确。

### 5.2 失败标准

- 参数错误返回 `40001`。
- 未登录或 token 失效返回 `40004`。
- 无权限返回 `40003`。
- 资源不存在返回 `40401`。
- 冲突返回 `40902`。

## 6. 建议补测项

1. 用一条可回滚测试数据，补测 `POST /admin/v1/coupons` 与 `PUT /admin/v1/coupons/{id}`。
2. 补一条存在的用户、司机审核、订单记录，再测详情接口。
3. 如果后续接入更完整的权限体系，再补测 `403` 场景。
