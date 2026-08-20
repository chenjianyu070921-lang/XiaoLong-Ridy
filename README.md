# XiaoLong-Ridy 花小猪打车（仿）

基于 Go 微服务架构的花小猪打车仿制项目，采用 go-zero/goctl 风格的目录组织，由六位组员按六个模块协作开发。

> 当前状态：工程骨架。目录已按微服务拆分完毕，具体接口、数据模型和业务代码待各模块逐步补齐。

## 技术栈（规划）

- 语言：Go 1.26（`go.mod` 当前声明 `go 1.26.5`）
- 框架：go-zero 风格网关 + gRPC 微服务
- 存储：MySQL、Redis
- 中间件：etcd（服务注册发现）、Kafka / RabbitMQ（消息队列）
- 部署：Docker、`deploy/configs` 集中管理配置

## 目录结构

```text
XiaoLong-Ridy
├── api/                          # API 网关层
│   ├── admin/                    # 管理后台网关
│   ├── driver/                   # 司机端网关
│   ├── passenger/                # 乘客端网关
│   └── middleware/               # 网关公共中间件
├── common/                       # 公共工具、常量、错误码
├── deploy/                       # 部署相关
│   ├── configs/                  # 服务配置文件
│   └── docker/                   # Docker 构建与编排
├── job/                          # 定时任务
├── mq-consumer/                  # 消息队列消费者
│   ├── location-consumer/        # 位置消息消费者
│   └── order-event-consumer/     # 订单事件消费者
├── rpc/                          # RPC 微服务
│   ├── dispatchsvc/              # 派单调度服务
│   ├── locationsvc/              # 位置服务
│   ├── ordersvc/                 # 订单服务
│   ├── paysvc/                   # 支付服务
│   ├── pricesvc/                 # 计价服务
│   ├── pushesvc/                 # 消息推送服务
│   └── usersvc/                  # 用户服务
├── scripts/
│   └── sql/
│       └── migrate/              # 数据库迁移脚本
├── .dockerignore
├── .gitignore
├── go.mod
└── README.md
```

## 六模块分工

| 模块 | 成员 | 主要目录 | 职责 |
| --- | --- | --- | --- |
| 模块一：乘客端 | 成员 1 | `api/passenger` + `rpc/usersvc` | 注册登录、叫车下单、行程展示、支付入口、个人中心 |
| 模块二：司机端 | 成员 2 | `api/driver` | 司机认证、接单、行程服务、收入、成长体系 |
| 模块三：管理后台 | 成员 3 | `api/admin` | 用户/司机/订单管理、营销、风控、数据统计 |
| 模块四：订单与派单 | 成员 4 | `rpc/ordersvc` + `rpc/dispatchsvc` + `mq-consumer/order-event-consumer` | 订单状态机、派单调度、订单事件处理 |
| 模块五：计价与支付 | 成员 5 | `rpc/pricesvc` + `rpc/paysvc` | 计价规则、优惠、支付回调、结算对账 |
| 模块六：位置与基础设施 | 成员 6 | `rpc/locationsvc` + `rpc/pushesvc` + `mq-consumer/location-consumer` + `job` + `common` | 地图定位、轨迹、消息推送、公共基础能力 |

## 核心业务流程

```text
乘客叫车 → 订单创建 → 派单匹配 → 司机接单 → 司机接驾 → 行程开始 → 行程结束 → 乘客支付 → 平台结算 → 双方评价
```

## 开发约定

- API 网关统一放在 `api/<端>`，内部按 `handler`、`logic`、`svc`、`types` 分层。
- RPC 服务统一放在 `rpc/<svc>`，内部按 `config`、`logic`、`model`、`repository`、`server` 分层。
- 服务间只通过 RPC 或消息队列通信，不直接依赖其他服务内部实现。
- 数据库变更脚本统一放到 `scripts/sql/migrate`，按时间或版本顺序命名。
- 环境配置放 `deploy/configs` 或各服务的 `etc/`，密钥类配置不得提交到仓库。
- 公共代码（错误码、工具函数、通用类型）放入 `common`，禁止跨服务复制粘贴。
- 消息类消费者放在 `mq-consumer`，定时任务放在 `job`。

## 本地运行

当前为骨架阶段，尚未生成可执行入口；各服务补全 `main` 入口后：

```bash
# 1. 安装依赖
go mod tidy

# 2. 先启动依赖中间件：MySQL、Redis、etcd、消息队列
#    本地配置参考 deploy/configs 或各服务 etc/ 下的 yaml

# 3. 依次启动 RPC 服务（示例）
cd rpc/usersvc && go run .
cd rpc/ordersvc && go run .
cd rpc/dispatchsvc && go run .

# 4. 启动 API 网关（示例）
cd api/passenger && go run .
cd api/driver && go run .
cd api/admin && go run .

# 5. 启动消息消费者与定时任务（示例）
cd mq-consumer/orderclient-event-consumer && go run .
cd mq-consumer/location-consumer && go run .
cd job && go run .
```

## 建议里程碑

1. 基础设施：`common`、配置中心、网关鉴权、数据库迁移脚本
2. 用户与乘客端：`usersvc` + `api/passenger`
3. 订单与派单：`ordersvc` + `dispatchsvc` + 订单事件消费者
4. 司机端：`api/driver`
5. 计价与支付：`pricesvc` + `paysvc`
6. 管理后台、联调与部署：`api/admin` + Docker
