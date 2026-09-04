// 模块五答辩：渲染 8 张 mermaid 图为 PNG + 复制交付物到桌面「答辩」文件夹
// 用法：node scripts/gen_defense_png.js
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const REPO = 'c:/Users/hjy/Desktop/XiaoLong-Ridy';
const DESK = 'C:/Users/hjy/Desktop/答辩'; // 目标文件夹（UTF-8 由脚本保证正确）
const TMP = path.join(REPO, 'scripts', '_mmd_tmp');

fs.mkdirSync(DESK, { recursive: true });
fs.mkdirSync(TMP, { recursive: true });

// 8 张图：文件名 + mermaid 源码
const charts = [
  {
    file: '图1-全项目业务泳道图.png',
    src: `flowchart LR
  subgraph S0["乘客端（模块一）"]
    A1["叫车下单"]
    A2["查看预估"]
    A3["确认支付"]
  end
  subgraph S1["订单与派单（模块四）"]
    B1["创建订单"]
    B2["派单匹配"]
    B3["司机接单"]
    B4["行程开始/结束"]
  end
  subgraph S2["计价与支付（模块五）"]
    C1["价格预估 EstimatePrice"]
    C2["优惠计算 CalculateDiscount"]
    C3["支付预下单 CreatePayment"]
    C4["支付回调 NotifyPayment"]
    C5["司机结算 SettleOrder"]
    C6["退款 RefundPayment"]
  end
  subgraph S3["位置与推送（模块六）"]
    D1["位置上报 location-report"]
    D2["消息推送 pushsvc"]
  end
  A1 --> B1
  B1 --> C1
  C1 --> A2
  A2 --> C2
  C2 --> A3
  A3 --> C3
  C3 -->|"order.paid"| C4
  C4 --> C5
  B4 --> C6
  B2 --> D1
  C4 --> D2
  C5 --> D2`,
  },
  {
    file: '图2-模块五功能结构图.png',
    src: `flowchart TD
  M5["模块五 计价与支付"] --> PS["pricesvc 计价服务"]
  M5 --> PA["paysvc 支付服务"]
  M5 --> COM["common 公共层"]
  PS --> PS1["行程价格预估 EstimatePrice"]
  PS --> PS2["优惠券抵扣计算 CalculateDiscount"]
  PS --> PS3["实际费用落库 SaveActualOrderPrice"]
  PS --> PS4["计价规则引擎 rule/price_engine"]
  PS1 --> PS4
  PS2 --> D2_["rule/discount"]
  PA --> PA1["支付预下单 CreatePayment"]
  PA --> PA2["支付回调（幂等+验签） NotifyPayment"]
  PA --> PA3["支付查询 GetPayment"]
  PA --> PA4["退款 RefundPayment"]
  PA --> PA5["司机结算 SettleOrder"]
  PA1 --> CH["渠道抽象 channel mock/alipay/wechat"]
  PA4 --> CH
  PA5 --> ST["rule/settlement"]
  COM --> COM1["priceutil 金额工具 分/元+Add+Round"]
  COM --> COM2["errorx 统一错误码"]
  COM --> COM3["constants 状态/Key/Topic"]
  COM --> COM4["mq 事件定义 order.paid"]`,
  },
  {
    file: '图3-模块五业务闭环流程.png',
    src: `flowchart TD
  S1["① 行程配置 城市/车型/里程/时长/时刻"] --> S2["② 价格预估 EstimatePrice"]
  S2 --> S3{"是否夜间/高峰？"}
  S3 -->|"23:00-05:00"| S4["加收夜间附加费"]
  S3 -->|"早7-9 晚17-19"| S5["高峰动态调价 ×1.3~1.5"]
  S3 -->|"否"| S6["基础价计算"]
  S4 --> S7["③ 优惠抵扣 CalculateDiscount"]
  S5 --> S7
  S6 --> S7
  S7 --> S8["④ 创建支付 CreatePayment"]
  S8 --> S9["模拟收银台弹窗"]
  S9 --> S10["⑤ 支付回调 NotifyPayment"]
  S10 --> S11{"重复回调？"}
  S11 -->|"是"| S12["幂等：直接返回已处理"]
  S11 -->|"否"| S13["状态 待支付→支付成功 发 order.paid 事件"]
  S12 --> S14["⑥ 司机结算 SettleOrder"]
  S13 --> S14
  S14 --> S15["⑦ 退款 RefundPayment"]
  S15 --> S16["状态流转 已退款"]`,
  },
  {
    file: '图4-模块五分层架构.png',
    src: `flowchart TB
  subgraph L1["接入层"]
    G1["API 网关 / gRPC 客户端"]
  end
  subgraph L2["zrpc 服务层（go-zero）"]
    S1["pricesvc :50053"]
    S2["paysvc :50054"]
  end
  subgraph L3["logic 业务逻辑层"]
    LG1["Estimate / Discount SaveActualOrderPrice"]
    LG2["CreatePayment / NotifyPayment GetPayment / Refund / Settle"]
  end
  subgraph L4["领域层（纯函数，可单测）"]
    R1["rule 计价引擎 price_engine/night/peak/discount"]
    R2["rule 支付结算 settlement/refund"]
  end
  subgraph L5["基础设施层"]
    D1["MySQL 4张表 price_rule/order_price/payment/settlement"]
    D2["Redis 缓存"]
    D3["Kafka topic order.paid"]
    D4["渠道抽象 channel mock/alipay/wechat"]
  end
  G1 --> S1
  G1 --> S2
  S1 --> LG1
  S2 --> LG2
  LG1 --> R1
  LG2 --> R2
  R1 --> D1
  R2 --> D1
  LG2 --> D2
  LG2 --> D3
  LG2 --> D4
  D4 --> D1`,
  },
  {
    file: '图5-全项目微服务调用图.png',
    src: `flowchart LR
  subgraph API["API 网关"]
    AP["api/passenger 乘客端"]
    AD["api/driver 司机端"]
    AA["api/admin 管理后台"]
  end
  subgraph BIZ["业务服务"]
    U["usersvc"]
    O["ordersvc 订单"]
    D["dispatchsvc 派单"]
    L["locationsvc 位置"]
    P["pushsvc 推送"]
  end
  subgraph M5["模块五（本组）"]
    PR["pricesvc 计价"]
    PA["paysvc 支付"]
  end
  AP --> O
  AP --> PR
  AP --> PA
  AP --> U
  AD --> D
  AD --> O
  AA --> O
  AA --> PR
  AA --> PA
  O -->|"FinishTrip 实际费用落库"| PR
  O -->|"订单金额/状态"| PA
  PA -->|"GetOrder 取司机ID"| O
  PA -->|"order.paid 事件"| P
  D --> L
  O --> L`,
  },
  {
    file: '图6-支付回调时序图.png',
    src: `sequenceDiagram
  participant ALI as 支付宝/模拟渠道
  participant PAY as paysvc.NotifyPayment
  participant DB as MySQL(payment)
  participant KF as Kafka(order.paid)
  participant ORD as ordersvc.GetOrder
  participant ST as 结算单 settlement

  ALI->>PAY: 回调（payment_no + sign + 金额）
  PAY->>PAY: 验签（RSA2，失败直接拒绝）
  PAY->>DB: 事务开始
  DB-->>PAY: 读支付单
  alt 状态已支付（重复回调）
    PAY->>PAY: 幂等：直接返回 success，不发事件
  else 状态=待支付
    PAY->>DB: 金额比对（回调额=单金额）
    PAY->>DB: 条件更新 WHERE id=? AND status=待支付
    DB-->>PAY: 行受影响=1 → 状态=支付成功
  end
  PAY->>PAY: 事务提交
  PAY->>KF: 发 order.paid（outbox-lite，失败可对账补发）
  PAY->>ORD: GetOrder(order_id) 取 driver_id
  ORD-->>PAY: driver_id
  PAY->>ST: CalcSettlement 生成结算单
  PAY-->>ALI: success`,
  },
  {
    file: '图7-模块五开发里程碑.png',
    src: `gantt
  title 模块五开发里程碑（2026-08）
  dateFormat  YYYY-MM-DD
  axisFormat  %m-%d
  section 需求与设计
    计价规则梳理（组长日check）  :a1, 2026-08-13, 1d
    DDL + 接口设计             :a2, 2026-08-13, 1d
  section 研发
    服务骨架 + 3基础接口        :b1, 2026-08-13, 1d
    支付闭环（回调/结算/退款）   :b2, 2026-08-14, 1d
    高峰溢价 rule 扩展          :b3, 2026-08-14, 1d
    全链路联调                 :b4, 2026-08-17, 1d
  section 交付与答辩
    答辩演示页面（真实引擎）     :c1, 2026-08-27, 1d
    接口文档/对接说明/联调记录   :c2, 2026-08-27, 1d
    答辩材料（本次）            :c3, 2026-08-28, 1d`,
  },
  {
    file: '图8-实训计划落地对照.png',
    src: `gantt
  title 实训计划落地对照
  dateFormat  YYYY-MM-DD
  axisFormat  %m-%d
  section 组长要求
    行程价格预估接口           :t1, 2026-08-13, 1d
    优惠券抵扣计算接口         :t2, 2026-08-13, 1d
    支付预下单接口             :t3, 2026-08-13, 1d
    对接订单状态回调           :t4, 2026-08-14, 1d
    对接推送模块通知接口        :t5, 2026-08-14, 1d
    产出接口文档/结算SQL/代码   :t6, 2026-08-14, 1d
  section 本人落地
    计价+优惠+支付 三个RPC     :d1, 2026-08-13, 1d
    支付回调+结算+退款闭环     :d2, 2026-08-14, 1d
    GetOrder+consumer 对接说明 :d3, 2026-08-14, 1d
    接口文档 + DDL + seed      :d4, 2026-08-14, 1d
    全链路联调验证             :d5, 2026-08-17, 1d
    答辩演示页（真实引擎）      :d6, 2026-08-27, 1d`,
  },
];

// 渲染每张图
let done = 0;
for (const c of charts) {
  const mmd = path.join(TMP, c.file.replace('.png', '.mmd'));
  fs.writeFileSync(mmd, c.src, 'utf8');
  const out = path.join(DESK, c.file);
  try {
    execSync(
      `npx mmdc -i "${mmd}" -o "${out}" -w 1600 -b white -s 2`,
      { stdio: 'inherit', cwd: REPO, encoding: 'utf8' }
    );
    console.log('OK  ' + c.file);
    done++;
  } catch (e) {
    console.log('FAIL ' + c.file + ': ' + (e.message || e));
  }
}

// 复制其他交付物
const copies = [
  [path.join(REPO, 'docs/module5/答辩/模块五答辩稿_完整版.docx'), path.join(DESK, '模块五答辩稿_完整版.docx')],
  [path.join(REPO, 'docs/module5/答辩/图表预览.html'), path.join(DESK, '图表预览.html')],
  [path.join(REPO, 'docs/module5/答辩/README.md'), path.join(DESK, '答辩速查手册.md')],
  [path.join(REPO, 'docs/module5/demo/index.html'), path.join(DESK, '演示页.html')],
];
for (const [src, dst] of copies) {
  try {
    fs.copyFileSync(src, dst);
    console.log('COPY ' + path.basename(dst));
  } catch (e) {
    console.log('COPY FAIL ' + path.basename(dst) + ': ' + e.message);
  }
}

console.log(`\n完成：${done}/${charts.length} 张图渲染成功，交付物已复制到 ${DESK}`);
