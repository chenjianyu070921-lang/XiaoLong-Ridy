# 管理后台优惠券与活动状态一致性修复设计

## 范围

本设计仅修复 `rpc/adminsvc` 中优惠券启用状态和营销活动审计事务的一致性问题。不修改 HTTP 路由、protobuf、数据库结构或业务数据。

## 优惠券规则

优惠券模板状态统一为：`1=草稿`、`2=启用`、`3=停用`。

`IssueCoupon` 仅允许状态为 `2` 的模板发券；草稿和停用模板返回 `FailedPrecondition`，不创建 `user_coupon` 或 `admin_coupon_issue_task`。优惠券统计的启用数量仅统计 `status=2`。

## 活动审计

创建和编辑活动时，`promotion_activity` 的新增或更新与 `admin_operation_log` 的写入使用同一 MySQL 事务。任一步失败均回滚，调用方不会收到业务失败但活动数据已经提交的结果。

## 测试

使用 sqlmock 覆盖草稿券拒绝、启用券可发放、启用券统计 SQL 口径，以及活动创建/编辑的审计失败回滚。执行 `go test ./api/admin/... ./rpc/adminsvc/...` 并使用仓库内独立 `GOCACHE`。
