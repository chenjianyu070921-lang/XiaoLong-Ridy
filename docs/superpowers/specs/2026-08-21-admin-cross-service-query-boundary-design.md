# 管理后台跨服务查询边界设计

## 目标

消除 `adminsvc` 对用户、订单及其关联业务表的直接读取。`adminsvc` 保留后台鉴权、审计和 HTTP/RPC 适配职责；数据归属服务负责查询、筛选、分页和聚合。

## 服务契约

`usersvc` 新增只读后台查询 RPC：`AdminListUsers`、`AdminGetUser`。请求承载当前后台已有的关键字、状态、注册时间和分页字段；响应保持现有 `adminsvc.User` 可映射字段与总数。

`ordersvc` 新增只读后台查询 RPC：`AdminListOrders`、`AdminGetOrder`、`AdminListAbnormalOrders`、`AdminGetStatistics`。请求承载订单号关键字、用户/司机 ID、状态、时间范围、异常类型和分页字段。订单服务在自身仓储内查询订单、状态流水、派单和支付相关数据；统计在订单服务一次快照内汇总订单指标。

优惠券、黑名单和后台运营表目前归 adminsvc 管理，保持在 adminsvc；支付金额、超时派单、订单异常等订单域指标在 ordersvc 契约中统一返回，避免 adminsvc 再访问 `payment`、`dispatch_record`。

## 迁移步骤

1. 在 usersvc/ordersvc proto 和服务端实现后台只读 RPC，并为筛选、分页、空结果和数据库错误建立测试。
2. 在 adminsvc 配置 usersvc 客户端，调用新 RPC 并映射为现有 admin proto，HTTP 接口不变。
3. 删除 adminsvc 中针对 `user`、`ride_order`、`order_status_log`、`dispatch_record`、`payment`、`settlement` 的读取 SQL 及其 SQL mock 用例。
4. 在集成环境验证 adminsvc 仅通过下游 RPC 获取用户和订单域数据；下游不可用时返回明确的 gRPC 错误，不降级为不完整数据。

## 一致性与性能

单服务统计由数据所属服务使用单次聚合查询完成。跨服务运营总览允许标注为最终一致的组合视图；若产品要求严格一致，应建设独立的运营读模型，而不是由 adminsvc 跨库事务读取。

## 非目标

本阶段不移动表、不修改数据库结构、不将后台权限规则下放到 usersvc/ordersvc，也不让网关直连任意业务服务。
