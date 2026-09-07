# XiaoLong-Ridy 花小猪打车（仿）

基于 Go 微服务架构的花小猪打车仿制项目，采用 go-zero/goctl 风格的目录组织，由六位组员按六个模块协作开发。

> 当前状态：六模块主体开发已完成，2026-09-07 订单与派单模块完成 6 个 P0 高危 BUG 修复（优惠券事务原子化、退款链路打通、支付创建失败补偿、派单候选为空重试、geo availability 强制校验），代码可构建、核心单元测试全部通过。主链路「乘客叫车 → 派单 → 司机接单 → 行程 → 支付 → 结算」处于联调收口阶段。历史已知卡点与 2026-09-07 修复明细见「当前开发进度」。

## 技术栈（已落地 / 规划）

- 语言：Go 1.25（`go.mod` 当前声明 `go 1.25.0`）
- 框架：go-zero 风格 API 网关 + gRPC 微服务（服务间通过 RPC 直连 `Target`，etcd 注册为可选）
- 存储：MySQL、Redis（GEO 派单、分布式锁、验证码、实时推送）
- 中间件：Kafka（订单 / 位置事件，基于 IBM/sarama）、Redis Pub/Sub（司机实时接单推送）
- 前端：Vue 3（`web/user` 乘客/司机端 + `web/admin` 管理后台）
- 部署：Docker Compose（`deploy/docker/infra.yml` 中间件编排）

## 目录结构

```text
XiaoLong-Ridy
├── api/                          # API 网关层
│   ├── admin/                    # 管理后台网关
│   ├── driver/                   # 司机端网关（含 WebSocket 实时通道）
│   ├── passenger/                # 乘客端网关
│   └── middleware/               # 网关公共中间件
├── common/                       # 公共工具、常量、错误码（28 个子包）
├── deploy/
│   └── docker/                   # 中间件编排 infra.yml
├── job/                          # 定时任务（位置清理 / 每日报表 / 派单重试）
├── mq-consumer/                  # 消息队列消费者
│   ├── location-consumer/        # 位置消息消费者
│   └── order-event-consumer/     # 订单事件消费者（订单 / 派单 / 支付事件）
├── rpc/                          # RPC 微服务（共 9 个）
│   ├── adminsvc/                 # 管理后台服务
│   ├── dispatchsvc/              # 派单调度服务
│   ├── driversvc/                # 司机服务
│   ├── locationsvc/              # 位置服务
│   ├── ordersvc/                 # 订单服务（状态机）
│   ├── paysvc/                   # 支付服务
│   ├── pricesvc/                 # 计价服务
│   ├── pushesvc/                 # 消息推送服务
│   └── usersvc/                  # 用户服务
├── scripts/
│   ├── admin-local/              # 管理后台本地启动脚本
│   ├── e2e/                      # 端到端联调客户端（支付回调模拟等）
│   ├── sql/migrate/              # 数据库迁移脚本（17 份）
│   └── start_all_dev.ps1         # 全量开发环境启动脚本
├── web/                          # 前端
│   ├── admin/                    # 管理后台（Vue 3）
│   └── user/                     # 乘客 / 司机端（Vue 3 + Vant）
├── agent/ 与 cmd/                # AI 智能体能力（react-agent）
├── go.mod
└── README.md
```

## 六模块分工

| 模块 | 成员 | 主要目录 | 职责 | 完成度 |
| --- | --- | --- | --- | --- |
| 模块一：乘客端 | 成员 1 | `api/passenger` + `rpc/usersvc` + `web/user` | 注册登录、叫车下单、行程展示、支付入口、个人中心 | 90% |
| 模块二：司机端 | 成员 2 | `api/driver` + `rpc/driversvc` + `web/user` | 司机认证、接单、行程服务、收入、成长体系 | 90% |
| 模块三：管理后台 | 成员 3 | `api/admin` + `rpc/adminsvc` + `web/admin` | 用户/司机/订单管理、营销、风控、数据统计 | 95% |
| 模块四：订单与派单 | 成员 4 | `rpc/ordersvc` + `rpc/dispatchsvc` + `mq-consumer/order-event-consumer` | 订单状态机、派单调度、订单事件处理 | 90% |
| 模块五：计价与支付 | 成员 5 | `rpc/pricesvc` + `rpc/paysvc` | 计价规则、优惠、支付回调、结算对账 | 85% |
| 模块六：位置与基础设施 | 成员 6 | `rpc/locationsvc` + `rpc/pushesvc` + `mq-consumer/location-consumer` + `job` + `common` | 地图定位、轨迹、消息推送、公共基础能力 | 80% |

## 核心业务流程

```text
乘客叫车 → 订单创建 → 派单匹配 → 司机接单 → 司机接驾 → 行程开始 → 行程结束 → 乘客支付 → 平台结算 → 双方评价
```

## 当前开发进度

### 主链路状态（代码已打通，联调收口阶段）

乘客端 `orders/create` → `ordersvc.CreateOrder`（调 `pricesvc` 预估金额快照 + 发 `order.created` 事件）→ `dispatchsvc` Redis GEO 派单 → `order-event-consumer` 写司机待接单集合 + Redis Pub/Sub 推送 → 司机端 `orders/accept` 接单 → 接驾/开始/结束行程 → `paysvc` 支付回调 → `ordersvc.ConfirmPaid`（订单已完成）+ `SettleOrder`（司机结算）。全链路单元测试与支付 e2e 脚本均已通过。

### 已知卡点（含 2026-09-07 已修复项）

| 级别 | 卡点 | 说明 | 责任 | 状态 |
| --- | --- | --- | --- | --- |
| P0 | ~~ConfirmPaid 优惠券事务外核销~~ | ~~订单状态 Completed 后 ConsumeByOrder 失败，优惠券永久卡 Locked~~ | 成员 4 | ✅ 已修复 09-07 |
| P0 | ~~RefundOrder 未触发 paysvc 真实退款~~ | ~~正常退款只改状态不调通道，用户支付不会退回~~ | 成员 4 | ✅ 已修复 09-07 |
| P0 | ~~FinishTrip createPayment 失败无补偿~~ | ~~订单先改 WaitPay 再调 paysvc，后者挂则订单永久卡待支付~~ | 成员 4 | ✅ 已修复 09-07 |
| P0 | ~~CancelOrder 双重释放优惠券~~ | ~~事务内 CancelWithCoupon 已释放，logic 层又调一次 ReleaseCoupon~~ | 成员 4 | ✅ 已修复 09-07 |
| P0 | ~~geo 引擎 availability nil 绕过~~ | ~~未注入 busy/online 过滤时，忙碌司机也会进派单候选~~ | 成员 4 | ✅ 已修复 09-07 |
| P0 | ~~Dispatch 候选为空不触发重试~~ | ~~dispatchsvc 返回空列表被 ordersvc 视为成功，订单一直 WaitAccept 直到超时~~ | 成员 4 | ✅ 已修复 09-07 |
| P0 | job 服务扫描 PaymentRetryQueueKey | enqueuePaymentRetry 已实现入队端，需 job 侧消费者扫描重调 CreatePayment | 成员 6 | ❌ 待开发 |
| P0 | paysvc 测试 mockOrderClient 缺 GetUserId 方法 | 原有问题，`rpc/paysvc/internal/logic` 测试 build failed | 成员 5 | ❌ 待修复 |
| P0 | 支付回调网关缺失 | api 层无 HTTP 回调路由，支付闭环当前依赖 `scripts/e2e/pay_e2e_client.go` 模拟回调 | 成员 5 | ❌ 待开发 |
| P0 | 定时任务接线待验证 | 订单超时关闭 / 派单重试补偿依赖 `job` 实际运行 | 成员 4 + 成员 6 | ❌ 待联调 |
| P1 | pushesvc 未接入主流程 | 派单通知实际走 Redis Pub/Sub，推送服务独立可用但未接线 | 成员 6 | ❌ 待接线 |
| P1 | 司机端 WebSocket 实时接单未端到端验证 | 断连时依赖 `/orders/available` 轮询兜底 | 成员 2 | ❌ 待联调 |
| P1 | job 扫描 RefundRetryQueueKey / DispatchRetryQueueKey | 订单/退款/支付三个重试队列的消费者待实现 | 成员 6 | ❌ 待开发 |

## 开发约定

- API 网关统一放在 `api/<端>`，内部按 `handler`、`logic`、`svc`、`types` 分层。
- RPC 服务统一放在 `rpc/<svc>`，内部按 `config`、`logic`、`model`、`repository`、`server` 分层。
- 服务间只通过 RPC 或消息队列通信，不直接依赖其他服务内部实现。
- 数据库变更脚本统一放到 `scripts/sql/migrate`，按时间或版本顺序命名。
- 环境配置放各服务的 `etc/`，密钥类配置不得提交到仓库。
- 公共代码（错误码、工具函数、通用类型、Redis/Kafka 键与事件常量）放入 `common`，禁止跨服务复制粘贴。
- 消息类消费者放在 `mq-consumer`，定时任务放在 `job`。
- 跨服务联调脚本（如支付回调模拟）放在 `scripts/e2e`。

## 本地运行

推荐方式：先启动中间件编排 `docker compose -f deploy/docker/infra.yml up -d`（MySQL、Redis、Kafka），再执行 `.\scripts\start_all_dev.ps1` 一键启动全部服务。

管理后台模块单独启动：`.\scripts\admin-local\start.ps1`
（需先设置环境变量 `ADMINSVC_MYSQL_DSN` 与 `ADMINSVC_REDIS_PASSWORD`）。

手动多服务启动参考：

```bash
# 1. 安装依赖
go mod tidy

# 2. 先启动依赖中间件：MySQL、Redis、Kafka
#    配置参考各服务 etc/ 下的 yaml（RPC 服务间通过 Target 直连）

# 3. 依次启动 RPC 服务（示例）
cd rpc/usersvc && go run .
cd rpc/ordersvc && go run .
cd rpc/dispatchsvc && go run .

# 4. 启动 API 网关（示例）
cd api/passenger && go run .
cd api/driver && go run .
cd api/admin && go run .

# 5. 启动消息消费者与定时任务（示例）
cd mq-consumer/order-event-consumer && go run .
cd mq-consumer/location-consumer && go run .
cd job && go run .

# 6. 启动前端（示例）
cd web/user && npm install && npm run dev
cd web/admin && npm install && npm run dev

# 7. 支付回调联调（当前 api 层回调网关未落地时）
cd scripts/e2e && .\run_pay_e2e.ps1
```

## 里程碑

1. ✅ 基础设施：`common`、网关鉴权、数据库迁移脚本（17 份 SQL）
2. ✅ 用户与乘客端：`usersvc` + `api/passenger`
3. ✅ 订单与派单：`ordersvc` 状态机 + `dispatchsvc` Redis GEO 派单 + 订单事件消费者
4. ✅ 司机端：`api/driver` + `driversvc`
5. ✅ 计价与支付：`pricesvc` + `paysvc`（支付 e2e 已通过）
6. ✅ 管理后台：`adminsvc` + `web/admin`
7. ✅ 订单与派单 P0 高危 BUG 修复（2026-09-07）：6 项资金/数据一致性问题，详见下方
8. 🔄 当前阶段：三个重试队列消费者落地 + 支付回调网关 + 定时任务接线 + 前端端到端验证

## 订单与派单 P0 修复明细（2026-09-07）

针对订单与派单调度模块的代码级审查，发现并修复 6 个 P0 高危 BUG。全量 `go build ./...` 通过，`ordersvc` / `dispatchsvc` 核心模块单元测试全部通过。

| 编号 | BUG 描述 | 根因 | 修复方案 | 改动文件 |
| --- | --- | --- | --- | --- |
| P0-1 | **ConfirmPaid 优惠券泄漏** | `CompleteOrder` MySQL 事务提交后才调 `CouponConsumer.ConsumeByOrder`，后者失败则订单已 Completed、优惠券永久卡 Locked | 将核销 SQL 纳入 `CompleteOrder` 同一事务，直接用 `tx.Table("user_coupon")...Updates(...)` 原子化 | `order_repository.go` 接口签名、`order_ext.go` GORM+Memory 实现、`confirm_paid_logic.go` 调用 + 删除独立 Consume 调用 |
| P0-2 | **RefundOrder 未触发 paysvc 真实退款** | 正常退款只改订单状态为 Refunded + 释放优惠券，无事件发布、无通道调佣——用户支付不会退回 | 末尾增加发布 `TopicOrderRefunded` 事件 + `enqueueRefundRetry` 写入 Redis ZSet，复用 ForceRefund 的 `orderRefundedEvent` 结构体 | `refund_order_logic.go` 新增 json/redis import + 事件发布 + `enqueueRefundRetry` 方法 |
| P0-3 | **FinishTrip createPayment 失败无补偿** | 订单先改 WaitPay 再调 `paysvc.CreatePayment`，后者挂则订单永久卡"待支付但无支付单" | 新增 `PaymentRetryQueueKey`（Redis ZSet），`createPayment` 返回空串时入队，待 job 服务扫描重调 | `constants.go` 加常量、`finish_trip_logic.go` 加入队逻辑 + `paymentRetryEvent` 结构体 + `enqueuePaymentRetry` 方法 |
| P0-4 | **CancelOrder 双重释放优惠券** | `CancelWithCoupon` 事务内已释放券，logic 层又调了一次 `ReleaseCoupon`（虽然第二次 WHERE status=Locked 命中 0 行不报错，但易混淆维护者） | 删除 logic 层冗余调用，统一由 Repository 事务内完成 | `cancel_order_logic.go` 删除 L103-107 |
| P0-5 | **geo 引擎 availability nil 绕过** | `filterAvailable` 在 `availability == nil` 时直接返回原切片不做过滤，忙碌司机也会进候选 | 改为 panic 而非静默放行——geo 引擎必须注入 Redis busy/online 检查 | `geo_dispatch_engine.go` filterAvailable + 同步更新单元测试验证 panic |
| P0-6 | **Dispatch 候选为空不触发重试** | dispatchsvc 返回"成功但空列表"被 ordersvc 视为派单成功，订单一直 WaitAccept 直到超时 | dispatchsvc 定义 `ErrNoAvailableDriver`，候选为空时返回此 error；ordersvc 已有 `DispatchOrder error → enqueueDispatchRetry` 链路，自动入队延迟重试 | `errors.go` 加错误定义、`dispatch_order_logic.go` 加候选为空判断 |

### 修复后待联动开发项

1. `job` 服务扫描 `PaymentRetryQueueKey` 重调 `CreatePayment`（当前只有入队端）
2. `job` 服务扫描 `RefundRetryQueueKey` / `DispatchRetryQueueKey` / `PaymentRetryQueueKey` 三个队列统一重试框架
3. 验证 `order-event-consumer` 已消费 `TopicOrderRefunded` 并调 `paysvc.RefundPayment`
4. 修复 `rpc/paysvc/internal/logic` 测试 build failed（mockOrderClient 缺 GetUserId）


