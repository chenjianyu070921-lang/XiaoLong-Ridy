# 管理后台 P0 闭环设计

## 目标

关闭管理后台的数据库迁移可用性、管理员注册越权防护和后台取消订单联调三个 P0 问题。

## 数据库迁移

迁移目标为远程 `xiaolong_ridy` 数据库。执行前先使用只读查询核验 `07_admin_operation_business.sql` 定义的六张表是否存在；仅在缺失时执行该脚本。导出任务和审计补偿表由 `09_admin_export_audit_task.sql` 定义，须单独授权执行。脚本只包含 `CREATE TABLE IF NOT EXISTS`，不修改已有表或业务数据。执行后核验表、索引和最小查询，并记录执行结果。

远程数据库结构变更须在执行前获得明确授权。

## 注册权限

当系统已有管理员时，注册请求必须从 gRPC metadata 解析已登录操作者，并同时满足：操作者角色为超级管理员、metadata 中的管理员 ID 等于请求中的 `operator_admin_id`。新增单元测试覆盖首管理员、非超级管理员、伪造操作者 ID 和合法超级管理员四种情况。

## 订单服务配置

为 ordersvc 增加配置加载回归测试，确认 Redis 配置不会与 go-zero `RpcServerConf` 产生解析冲突。Redis 仍使用订单服务独立配置，并由 `ServiceContext` 初始化。完成后运行 ordersvc 与 adminsvc 取消订单测试，验证管理员取消请求继续传递 `operator_type=admin` 和 `operator_id=admin_id`。

实际 RPC 联调需要本机 MySQL、Redis、etcd、ordersvc 与 adminsvc 均可用；该联调不写入远程业务数据。

## 验收

- `07` 中六张运营表存在且可查询；`09` 中两张表按单独授权迁移。
- 越权注册请求返回权限拒绝，合法超级管理员注册成功。
- ordersvc 可加载配置并通过测试。
- adminsvc 取消订单转发测试验证管理员操作者身份字段。
