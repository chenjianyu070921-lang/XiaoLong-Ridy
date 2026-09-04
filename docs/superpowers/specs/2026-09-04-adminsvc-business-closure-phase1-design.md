# 管理后台 adminsvc 第一阶段业务闭环设计

日期：2026-09-04

## 一、目标与范围

本设计用于在不新增独立 RPC 服务的前提下，优先在现有 `adminsvc` 内完成管理后台高风险业务闭环。

第一阶段覆盖：

1. 司机违规处罚、处罚申诉、处罚规则管理；
2. 支付退款通道补偿；
3. 活动与计价服务的真实冲突检测、发布与回滚下发；
4. 大批量异步发券。

本阶段不覆盖独立报表服务、完整通知 MQ 重构、生产级导出存储、完整浏览器端到端验收。这些能力在第一阶段接口和事件稳定后继续实施。

## 二、架构原则

### 2.1 服务边界

管理后台保持既有调用结构：

```text
web/admin -> api/admin -> adminsvc -> 下游 RPC / Kafka
```

`adminsvc` 是后台流程编排、后台状态机、权限校验、审计与可靠事件落库的唯一入口。

下游领域服务保持数据所有权：

| 服务 | 数据与执行权 |
| --- | --- |
| `driversvc` | 司机冻结、禁接单、扣分、降权等司机状态动作 |
| `ordersvc` | 订单状态、退款状态机、退款事件 |
| `paysvc` | 支付通道退款结果和通道补偿执行 |
| `pricesvc` | 计价规则、活动冲突检测和活动规则实际生效 |
| `usersvc` | 用户券实际发放、领取上限和用户券状态 |
| `pushsvc` | 司机和用户通知实际发送 |
| `job` | outbox 抢占、Kafka 投递、失败重试和状态回写 |

`adminsvc` 不直接修改下游领域服务的数据表。

### 2.2 可靠事件模式

所有跨服务写操作采用“本地事务表 + job 重试 + Kafka 事件”模式：

1. `adminsvc` 在同一 MySQL 事务内写入后台业务状态、操作日志和 outbox 事件；
2. 事务提交后，HTTP 接口返回“任务已创建”或“请求已受理”；
3. `job` 以租约抢占 outbox 事件并投递 Kafka；
4. 下游消费者按事件号幂等执行；
5. 下游执行结果回写后台任务或由 `job` 查询并更新；
6. 失败事件使用指数退避重试，超过上限后标记失败并保留人工处理入口。

任何下游调用失败不得被包装为业务成功；任何已提交的后台业务请求也不得因异步执行暂时失败而丢失。

### 2.3 幂等规则

所有高风险写接口必须携带 `request_id`。

幂等键为：

```text
业务类型 + 聚合对象类型 + 聚合对象 ID + request_id
```

相同幂等键再次请求时，返回已有任务或处罚单的当前状态，不重复生成下游动作。

## 三、数据模型与迁移约束

新增 SQL 迁移脚本仅创建新表、索引和必要约束：

| 表 | 用途 |
| --- | --- |
| `admin_driver_punishment_rule` | 处罚规则、动作组合、启停状态和版本 |
| `admin_driver_punishment` | 处罚单、关联司机/订单、处罚动作、金额、状态与幂等号 |
| `admin_punishment_appeal` | 司机申诉、证据、审核意见和复核状态 |
| `admin_refund_compensation_task` | 退款失败补偿任务、重试、通道结果和人工处理记录 |
| `admin_coupon_issue_batch` | 异步发券批次、目标规则、分片统计和状态 |
| `admin_promotion_publish_task` | 活动发布/回滚任务、规则版本、冲突检查结果和投递状态 |
| `admin_domain_outbox` | 统一领域事件、载荷、状态、租约、重试次数和下次重试时间 |

迁移约束：

1. 代码不自动执行迁移；
2. 服务启动时不得建表或修改表结构；
3. 迁移脚本不执行已有业务数据的 `INSERT`、`UPDATE` 或 `DELETE`；
4. 新表必须包含创建时间、更新时间、状态和必要索引；
5. 迁移执行前必须按发布流程预检目标库状态。

## 四、业务能力设计

### 4.1 司机处罚与申诉

处罚对象仅限司机。第一期支持以下动作：

| 动作 | 下游执行方式 |
| --- | --- |
| 禁接单 | `driversvc` 处理接单资格 |
| 冻结 | `driversvc` 处理司机状态和在线状态 |
| 扣分 | `driversvc` 处理司机服务分 |
| 降权 | `driversvc` 处理派单权重或接单能力 |
| 罚款 | 仅创建待结算处罚记录，不直接扣减司机余额 |

处罚状态机：

```text
pending -> processing -> effective
                     -> failed
                     -> cancelled
```

处罚申诉状态机：

```text
pending -> reviewing -> upheld
                     -> revoked
                     -> rejected
```

申诉通过或处罚撤销必须创建反向 outbox 事件。不得直接跨库回写司机状态。

新增 HTTP 接口：

```text
GET  /admin/v1/punishment-rules
POST /admin/v1/punishment-rules
PUT  /admin/v1/punishment-rules/{id}
POST /admin/v1/punishment-rules/{id}/enable
POST /admin/v1/punishment-rules/{id}/disable

GET  /admin/v1/punishments
POST /admin/v1/punishments
GET  /admin/v1/punishments/{id}
POST /admin/v1/punishments/{id}/cancel

GET  /admin/v1/punishment-appeals
POST /admin/v1/punishment-appeals/{id}/review
```

### 4.2 退款补偿

现有退款入口保持不变。订单服务或支付服务返回明确失败、超时或通道不可用时，`adminsvc` 创建 `admin_refund_compensation_task` 和对应 outbox 事件。

退款补偿状态机：

```text
pending -> processing -> success
                     -> retrying
                     -> manual_review
                     -> failed
```

规则：

1. 使用 `request_id` 和 `refund_no` 保证幂等；
2. 禁止退款金额超过原支付成功金额；
3. 禁止对同一退款单重复发起补偿；
4. 每次尝试保存通道响应、失败原因、执行时间和操作人；
5. 达到最大重试次数后进入 `manual_review`，不得继续自动扣款或退款。

### 4.3 活动冲突检测与发布

活动发布流程：

1. `adminsvc` 校验时间范围、活动类型、目标人群、城市和优惠参数；
2. `adminsvc` 同步调用 `pricesvc` 进行真实冲突检测；
3. 无冲突时创建 `admin_promotion_publish_task`、审计日志和 outbox；
4. `job` 投递发布或回滚事件；
5. `pricesvc` 以规则版本幂等执行生效或回滚；
6. 下游结果回写发布任务。

发布任务状态：

```text
pending -> checking -> publishing -> success
                                  -> retrying
                                  -> failed
```

回滚使用独立任务和版本号，不得直接覆盖已生效规则。

### 4.4 大批量异步发券

发券 HTTP 请求只负责创建批次，不在请求内同步逐用户发券。

批次状态机：

```text
pending -> processing -> partial_success
                     -> success
                     -> failed
                     -> cancelled
```

规则：

1. 批次按固定大小分片；
2. 每个用户券操作使用批次号、用户 ID 和券 ID 幂等；
3. 重复用户、无效用户、已达领取上限分别记录失败原因；
4. 单用户失败不得回滚其他已成功用户；
5. 批次统计总量、成功数、失败数、处理中数量和最后处理位置；
6. 消费者必须由 `usersvc` 执行实际用户券写入。

## 五、Kafka 事件契约

第一期新增下列事件类型：

```text
admin.driver.punishment.requested
admin.driver.punishment.reversed
admin.driver.punishment.appeal.reviewed
admin.refund.compensation.requested
admin.coupon.issue.requested
admin.promotion.publish.requested
admin.promotion.rollback.requested
admin.notification.requested
```

每个事件至少包含：

```json
{
  "event_id": "唯一事件号",
  "event_type": "事件类型",
  "aggregate_type": "聚合类型",
  "aggregate_id": "聚合 ID",
  "request_id": "请求幂等号",
  "occurred_at": "事件生成时间",
  "payload": {}
}
```

消费者必须以 `event_id` 和领域业务幂等键双重去重。

## 六、权限与审计

| 角色 | 权限范围 |
| --- | --- |
| 超级管理员 | 所有操作，包括资金补偿、处罚撤销、规则启停、活动正式发布和回滚 |
| 运营 | 查询、创建处罚、初审申诉、保存处罚规则或活动草稿 |
| 客服 | 处罚和申诉只读查询 |

以下操作必须写入 `admin_operation_log`：

1. 创建、取消、撤销处罚；
2. 处罚规则创建、编辑、启停；
3. 申诉审核；
4. 创建、重试和人工处理退款补偿；
5. 创建、发布和回滚活动；
6. 创建、取消和重试发券批次。

审计与业务状态必须在同库操作中使用同一事务。跨服务执行后的补充审计失败时，写入现有或统一 outbox 补偿记录。

## 七、错误处理与任务重试

1. outbox 使用租约和状态比较更新，避免多实例重复抢占；
2. 任务采用指数退避，并设置最大尝试次数；
3. payload 不得保存密码、令牌、完整身份证号或银行卡号；
4. 下游超时与明确业务拒绝必须区分：超时可重试，业务拒绝直接失败；
5. 所有人工处理动作必须重新校验最新任务状态，避免覆盖异步执行结果；
6. 任务状态回写必须使用版本或条件更新，避免旧事件覆盖新状态。

## 八、测试与验收

### 8.1 单元测试

覆盖：

1. 状态机合法与非法迁移；
2. 退款金额、重复退款和退款补偿幂等；
3. 处罚规则校验、处罚创建、撤销和申诉审核；
4. 活动冲突、版本发布和回滚；
5. 异步发券分片、部分成功和重复消费；
6. outbox 租约、重试、最大次数和失败状态；
7. 超级管理员、运营、客服的角色边界。

### 8.2 集成测试

覆盖：

1. MySQL 事务中业务表、审计表和 outbox 的原子性；
2. Kafka 事件 payload 与消费者幂等；
3. `driversvc`、`paysvc`、`pricesvc`、`usersvc` 暂时不可用时的补偿行为；
4. 下游成功、超时、业务拒绝和重复回调；
5. 新增 HTTP 路由的参数校验与统一错误码。

### 8.3 发布验收

发布顺序：

1. 合入迁移脚本，不自动执行；
2. 预检数据库、Kafka topic、消费者配置和索引；
3. 部署 `adminsvc`、`job` 和消费者；
4. 先开放查询和任务创建；
5. 按灰度顺序开放处罚、发券、活动发布和退款补偿；
6. 监控 outbox 积压、任务失败率、重复消费率与审计完整率；
7. 完成真实数据联调后开放正式写操作。

## 九、非目标与后续阶段

第一期不拆分 `punishsvc`、`reportsvc`、`couponsvc`、`workordersvc` 或 `auditsvc`。

后续阶段按依赖顺序推进：

1. 独立报表服务和汇总指标体系；
2. 通知 MQ 全链路；
3. 生产级导出文件存储与任务治理；
4. 订单轨迹、支付、结算真实联调；
5. 浏览器端到端自动化验收；
6. 根据实际压力与数据所有权评估服务拆分。
