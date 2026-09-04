# XiaoLong-Ridy 花小猪打车（仿）

基于 Go 微服务架构的花小猪打车仿制项目，采用 go-zero/goctl 风格的目录组织，由六位组员按六个模块协作开发。

> 当前状态：六模块已完成主体开发，主链路「乘客叫车 → 派单 → 司机接单 → 接驾 → 行程 → 支付 → 结算」代码已打通，处于联调收口阶段。具体完成度与已知卡点见「当前开发进度」。

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

### 已知卡点

| 级别 | 卡点 | 说明 | 责任 |
| --- | --- | --- | --- |
| P0 | 支付回调网关缺失 | api 层无 HTTP 回调路由，支付闭环当前依赖 `scripts/e2e/pay_e2e_client.go` 模拟回调 | 成员 5 |
| P0 | 定时任务接线待验证 | 订单超时关闭 / 派单重试补偿依赖 `job` 实际运行 | 成员 4 + 成员 6 |
| P1 | pushesvc 未接入主流程 | 派单通知实际走 Redis Pub/Sub，推送服务独立可用但未接线 | 成员 6 |
| P1 | 司机端 WebSocket 实时接单未端到端验证 | 断连时依赖 `/orders/available` 轮询兜底 | 成员 2 |

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
7. 🔄 当前阶段：主链路全流程联调收口（支付回调网关、定时任务接线、前端端到端验证）











