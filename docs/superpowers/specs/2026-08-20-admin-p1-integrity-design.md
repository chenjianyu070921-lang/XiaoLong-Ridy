# 管理后台 P1 数据一致性与过滤设计

## 目标

修复管理后台优惠券、计价规则、统计和导出功能中的一致性、审计关联和筛选失效问题。变更遵循 `api/admin -> adminsvc -> pricesvc` 边界，不修改业务主表结构。

## 约束

1. `api/admin` 仅负责 HTTP 参数转换和响应映射。
2. `adminsvc` 不直接读写 `price_rule`，计价规则始终由 `pricesvc` 管理。
3. 现有 `admin_audit_outbox` 用于跨服务成功后的审计补偿；审计补偿不得导致调用方重试已经成功的业务操作。
4. 统计的 `city_code` 仅作用于订单及订单派生指标。用户、司机、优惠券和黑名单指标没有权威城市归属，保持全局统计。
5. 导出仅支持白名单过滤字段，未知字段、类型不匹配和非法时间格式必须返回参数错误。

## 优惠券创建与审计

`AdminService.CreateCoupon` 的响应由 `CommonResponse` 调整为 `CreateCouponResponse`，包含 `id` 和 `message`。

adminsvc 在一个数据库事务中完成以下两步：

1. 创建 `coupon` 模板并取得自增主键。
2. 使用同一个事务连接写入 `admin_operation_log`，目标 ID 为新优惠券 ID。

任一步失败均回滚，HTTP 层直接映射 RPC 返回 ID，不再通过优惠券名称查询最新记录。此设计消除同名并发创建时返回错误 ID 的问题。

## 计价规则创建与审计补偿

pricesvc 的 `CreatePriceRule` 返回 `CreatePriceRuleResponse{id,message}`，其中 `id` 为持久化后的规则 ID。adminsvc 同步转发该 ID，`AdminService.CreatePriceRule` 同样返回 `CreatePriceRuleResponse`。

计价规则创建、更新、启用和停用均遵循以下结果语义：

1. pricesvc 写入成功后，adminsvc 尝试写操作审计。
2. 审计写入成功，直接返回成功。
3. 审计写入失败，adminsvc 将审计事件写入 `admin_audit_outbox`，并返回业务成功。
4. outbox 写入也失败时返回内部错误，并记录结构化错误日志，供值班人员人工补审计；不得尝试回滚已经提交的 pricesvc 规则。

审计事件目标 ID 使用 pricesvc 的返回 ID，创建操作不再使用请求中的零值 ID。

## 统计过滤

所有订单及订单派生指标共享一个订单过滤器：

- `city_code`：使用订单记录的城市字段精确匹配。
- `start_time`、`end_time`：使用订单创建时间过滤。

运营总览中的订单数、完成订单数、异常订单数和 GMV 使用该过滤器。订单统计中的订单数、完成数、取消数、超时数和支付异常数同样使用该过滤器；后两者通过订单 ID 关联派单或支付表，避免遗漏时间和城市条件。

用户数、司机数、优惠券数、黑名单数及其派生指标保持全局聚合，即使请求含有 `city_code` 也不将其错误映射到其他表。

## 导出过滤

`ExportTaskRequest.filters` 仍使用 JSON 字符串传输，但在 adminsvc 解析为类型化结构。当前仅支持 `orders` 导出，并接受以下字段：

- `status`：正整数订单状态。
- `city_code`：非空城市编码。
- `start_time`、`end_time`：`YYYY-MM-DD HH:mm:ss`，结束时间不得早于开始时间。
- `user_id`、`driver_id`：正整数。

`writeExportCSV` 接收解析后的订单过滤器并将其绑定到 SQL 参数。非对象 JSON、未知字段、错误类型、非法时间或不支持的 `export_type` 返回 `InvalidArgument`，不创建任务，不生成全量文件。

## 测试与验收

1. 优惠券创建成功时响应包含真实 ID，操作日志与优惠券记录在同一事务提交；审计失败时两者均回滚。
2. 同名优惠券并发创建各自返回对应 ID，HTTP 层无名称回查。
3. pricesvc 创建规则返回 ID；adminsvc 审计使用该 ID；审计失败时写入 outbox 并返回成功。
4. 城市和时间范围能同时限制订单、超时和支付异常统计；全局指标不拼接城市条件。
5. 订单 CSV 只包含符合筛选条件的数据；未知筛选字段和不支持导出类型返回参数错误。
6. 执行 `go test ./api/admin/... ./rpc/adminsvc/... ./rpc/pricesvc/...`。
