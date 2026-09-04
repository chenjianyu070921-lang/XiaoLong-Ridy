# 模块五答辩演示页面

可视化展示「计价 → 优惠 → 支付 → 回调 → 结算 → 退款」完整业务闭环，答辩共享屏幕用。
计算结果与模块五真实代码 100% 一致（复用真实 `rule` 引擎）。

## 文件结构

| 文件 | 说明 |
| --- | --- |
| `docs/module5/demo/index.html` | 演示页面（单文件，双击即可用） |
| `rpc/pricesvc/demo/demo_server.go` | 真实演示服务（复用计价引擎等真实代码） |
| `rpc/paysvc/demoapi/demoapi.go` | 支付侧适配包（包装 paysvc internal，供演示服务复用） |
| `rpc/pricesvc/demo/demo_server_test.go` | 演示服务单测（数值与真实引擎断言对齐） |

## 两种运行模式

### 模式一：真实模式（推荐答辩使用）

```bash
# 在项目根目录启动
go run ./rpc/pricesvc/demo
# 浏览器打开 http://127.0.0.1:8787/
```

页面顶部徽标显示「真实计价引擎」。所有计算由 Go 服务完成：
复用 `rpc/pricesvc/internal/rule` 的 `Estimate`（计价引擎，含夜间跨天/高峰判断）、
`CalculateDiscount`（优惠）、`rpc/paysvc` 的 `CalcSettlement`（结算）与 `MockChannel`（支付渠道），
计价规则取自 `scripts/sql/migrate/07_price_rule_seed.sql` 的北京三条真实规则。

### 模式二：内置模拟模式（零依赖兜底）

不启动服务，直接双击打开 `docs/module5/demo/index.html`。
页面自动检测不到演示服务时，降级为内置 JS 同公式计算，功能完全一致。

## 答辩演示步骤

1. **行程配置**：默认北京快车 12.5km / 30 分钟；点快捷场景按钮切换 白天/早高峰/晚高峰/夜间跨天。
2. **价格预估**：展示命中的计价规则 + 五项费用明细；夜间时段与高峰溢价会自动亮徽标。
3. **优惠抵扣**：选「8 折券（满 20 元）」演示门槛 + 最大优惠上限约束；或「立减 5 元」。
4. **创建支付**：选支付宝渠道 → 弹出模拟收银台。
5. **支付回调**：点「确认支付」→ 状态变为支付成功；**再点一次「模拟重复回调」演示幂等**（核心亮点）。
6. **司机结算**：按 20% 抽成展示平台抽成 + 司机实收。
7. **退款**：全额退款 → 状态流转为「已退款」。

底部「≣」按钮打开**接口流水日志**，每步记录接口名 + 请求/响应，可与代码对照讲解。

## 演示数据与真实代码的对应关系

| 页面步骤 | 演示服务接口 | 真实代码位置 |
| --- | --- | --- |
| 价格预估 | `POST /api/estimate` | `rpc/pricesvc/internal/rule/price_engine.go`、`night.go`、`peak.go` |
| 优惠抵扣 | `POST /api/discount` | `rpc/pricesvc/internal/rule/discount.go` |
| 创建支付 | `POST /api/payment/create` | `rpc/paysvc/internal/channel/mock.go` |
| 支付回调（幂等） | `POST /api/payment/notify` | `rpc/paysvc/internal/logic/notify_payment_logic.go` |
| 司机结算 | `POST /api/settle` | `rpc/paysvc/internal/rule/settlement.go` |
| 退款 | `POST /api/payment/refund` | `rpc/paysvc/internal/logic/refund_payment_logic.go` |

金额全程「分」(int64) 计算、展示时转元，与 `common/priceutil` 约定一致。

## 验收数值（答辩可现场核对）

| 场景 | 期望 |
| --- | --- |
| 北京快车 12.5km/30min 白天 10:00 | 总价 50.75 元（1200+2375+1500） |
| 同上，早高峰 08:00 | factor=1.5，动态调价 25.38 元，总价 76.13 元 |
| 北京特惠快车 5km/10min 夜间 23:30 | 夜间费 5 元，总价 22.40 元 |
| 8 折券（满 20，最高减 10）用于 50.75 元 | 抵扣 10 元（受上限约束），实付 40.75 元 |
| 结算 40.75 元 × 20% | 平台 8.15 元 / 司机 32.60 元 |

## 常见问题

- **页面顶部显示「内置模拟模式」**：说明演示服务没启动。回到项目根执行 `go run ./rpc/pricesvc/demo`，刷新页面即可。
- **端口被占用**：`go run ./rpc/pricesvc/demo -addr 127.0.0.1:8788`，同时把页面里 `API_BASE` 改成对应端口（默认 8787）。
- **支付宝真实收银台**：演示服务用 mock 渠道生成模拟支付参数；真实支付宝沙箱收银台展示方式见 `docs/module5/计价支付接口文档.md` 第 4.6 节（配置沙箱密钥后 `CreatePayment` 返回真实跳转链接）。
