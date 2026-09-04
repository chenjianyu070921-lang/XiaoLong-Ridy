// 模块五答辩稿 .docx 生成脚本
// 用法：node scripts/gen_defense_docx.js
const fs = require('fs');
const path = require('path');
const {
  Document, Packer, Paragraph, TextRun, Table, TableRow, TableCell,
  AlignmentType, LevelFormat, BorderStyle, WidthType, ShadingType, PageNumber,
  PageBreak, Footer, HeadingLevel
} = require('C:\\Users\\hjy\\AppData\\Roaming\\npm\\node_modules\\docx');

const OUT = 'c:\\Users\\hjy\\Desktop\\XiaoLong-Ridy\\docs\\module5\\答辩\\模块五答辩稿_完整版.docx';

// ─────────── 样式常量 ───────────
const FONT = '微软雅黑';
const CONTENT_W = 9026; // A4 1英寸边距
const ORANGE = 'E55E00';
const HEAD_FILL = 'FFF3E9';
const GREY_FILL = 'F2F2F2';
const border = { style: BorderStyle.SINGLE, size: 1, color: 'CCCCCC' };
const borders = { top: border, bottom: border, left: border, right: border };

// ─────────── 辅助函数 ───────────
function p(text, opts = {}) {
  const runs = [];
  if (opts.bold) runs.push(new TextRun({ text, bold: true, font: FONT, size: opts.size || 21 }));
  else runs.push(new TextRun({ text, font: FONT, size: opts.size || 21 }));
  return new Paragraph({
    children: runs,
    spacing: { after: opts.after ?? 120, line: 300 },
    alignment: opts.align,
    shading: opts.shade ? { fill: opts.shade, type: ShadingType.CLEAR } : undefined,
    indent: opts.indent ? { left: opts.indent } : undefined,
  });
}

function h1(text) { return new Paragraph({ heading: HeadingLevel.HEADING_1, children: [new TextRun({ text, font: FONT, size: 30, bold: true, color: '1F4E79' })], spacing: { before: 320, after: 160 } }); }
function h2(text) { return new Paragraph({ heading: HeadingLevel.HEADING_2, children: [new TextRun({ text, font: FONT, size: 24, bold: true, color: ORANGE })], spacing: { before: 240, after: 120 } }); }
function h3(text) { return new Paragraph({ heading: HeadingLevel.HEADING_3, children: [new TextRun({ text, font: FONT, size: 21, bold: true })], spacing: { before: 160, after: 80 } }); }

function talk(text) {
  // 【口述】强调块
  return new Paragraph({
    children: [new TextRun({ text: '【口述】', bold: true, font: FONT, size: 21, color: ORANGE }), new TextRun({ text, font: FONT, size: 21 })],
    spacing: { after: 160, line: 300 },
    shading: { fill: 'FFF9F0', type: ShadingType.CLEAR },
  });
}

function note(text) {
  return new Paragraph({
    children: [new TextRun({ text: '【提示】', bold: true, font: FONT, size: 20, color: '888888' }), new TextRun({ text, font: FONT, size: 20, color: '555555' })],
    spacing: { after: 140 },
  });
}

function cell(text, w, opts = {}) {
  const runs = Array.isArray(text) ? text.map(t => new TextRun({ text: t, font: FONT, size: 19, bold: opts.bold })) : [new TextRun({ text, font: FONT, size: 19, bold: opts.bold })];
  return new TableCell({
    borders,
    width: { size: w, type: WidthType.DXA },
    shading: opts.fill ? { fill: opts.fill, type: ShadingType.CLEAR } : undefined,
    margins: { top: 60, bottom: 60, left: 100, right: 100 },
    children: [new Paragraph({ children: runs, spacing: { after: 0, line: 260 } })],
  });
}

function table(headers, rows, widths) {
  const total = widths.reduce((a, b) => a + b, 0);
  const scale = CONTENT_W / total;
  const ws = widths.map(w => Math.round(w * scale));
  const head = new TableRow({ tableHeader: true, children: headers.map((h, i) => cell(h, ws[i], { bold: true, fill: HEAD_FILL })) });
  const body = rows.map(r => new TableRow({ children: r.map((c, i) => cell(c, ws[i])) }));
  return new Table({
    width: { size: CONTENT_W, type: WidthType.DXA },
    columnWidths: ws,
    rows: [head, ...body],
  });
}

function pageBreak() { return new Paragraph({ children: [new PageBreak()] }); }

// ─────────── 文档内容 ───────────
const children = [];

// ═══ 封面 ═══
children.push(
  new Paragraph({ spacing: { before: 2400 }, alignment: AlignmentType.CENTER, children: [new TextRun({ text: '花小猪打车（仿）', font: FONT, size: 56, bold: true, color: ORANGE })] }),
  new Paragraph({ alignment: AlignmentType.CENTER, spacing: { after: 120 }, children: [new TextRun({ text: '模块五 · 计价与支付', font: FONT, size: 40, bold: true })] }),
  new Paragraph({ alignment: AlignmentType.CENTER, spacing: { after: 800 }, children: [new TextRun({ text: '—— 答辩主讲稿 ——', font: FONT, size: 28, color: '888888' })] }),
  new Paragraph({ alignment: AlignmentType.CENTER, spacing: { after: 80 }, children: [new TextRun({ text: '主讲人：乔宇翔（成员 5）', font: FONT, size: 26 })] }),
  new Paragraph({ alignment: AlignmentType.CENTER, spacing: { after: 80 }, children: [new TextRun({ text: '日期：2026-08-28', font: FONT, size: 24 })] }),
  new Paragraph({ alignment: AlignmentType.CENTER, spacing: { after: 80 }, children: [new TextRun({ text: '分支：module5-qyx', font: FONT, size: 22, color: '888888' })] }),
  new Paragraph({ alignment: AlignmentType.CENTER, children: [new TextRun({ text: '时长：主讲 30 分钟 + 老师提问', font: FONT, size: 22, color: '888888' })] }),
  pageBreak(),
);

// ═══ 第2章 摘要 ═══
children.push(
  h1('2. 一句话项目定位'),
  p('本项目是「花小猪打车」的微服务仿制项目，Go 语言、go-zero zrpc + gRPC + GORM + MySQL + Redis + Kafka 的分布式架构。六位成员按六个业务模块并行开发，我负责的是**模块五：计价与支付**（rpc/pricesvc + rpc/paysvc）。', { bold: false }),
  talk('各位老师好，我是本项目模块五「计价与支付」的负责人。一句话概括项目：用 Go 微服务复刻网约车的核心交易链路，其中我负责的是乘客叫车之后最关键的环节——钱怎么算、怎么收、怎么分。我接下来的讲解会围绕这条业务链路展开，最后会有 8 分钟的现场演示。'),
  p('模块五解决的是网约车交易中最敏感的四个问题：价格怎么算（计价）、优惠怎么减（优惠抵扣）、钱怎么收（支付闭环）、钱怎么分（司机结算），以及一旦出错怎么办（退款与幂等）。', {}),
  pageBreak(),
);

// ═══ 第3章 项目架构 ═══
children.push(
  h1('3. 项目架构'),
  h2('3.1 全项目六大模块'),
  table(
    ['模块', '目录', '核心职责'],
    [
      ['模块一：乘客端', 'api/passenger + rpc/usersvc', '注册登录、叫车下单、支付入口、个人中心'],
      ['模块二：司机端', 'api/driver', '司机认证、接单、行程服务、收入'],
      ['模块三：管理后台', 'api/admin', '用户/司机/订单管理、营销、数据统计'],
      ['模块四：订单与派单', 'rpc/ordersvc + dispatchsvc + mq-consumer', '订单状态机、派单调度、订单事件'],
      ['模块五：计价与支付（本组）', 'rpc/pricesvc + rpc/paysvc', '计价规则、优惠、支付回调、结算对账'],
      ['模块六：位置与基础设施', 'rpc/locationsvc + pushsvc + job + common', '地图定位、消息推送、公共基础能力'],
    ],
    [1700, 3300, 4026],
  ),
  h2('3.2 模块五分层架构（go-zero zrpc 标准）'),
  table(
    ['层', '组件', '说明'],
    [
      ['接入层', 'API 网关 / gRPC 客户端', '乘客端、订单模块、管理后台通过 gRPC 调用'],
      ['服务层', 'pricesvc :50053 / paysvc :50054', 'go-zero zrpc，proto 定义接口'],
      ['逻辑层', 'internal/logic', '每个 RPC 一个 logic，负责业务流程编排'],
      ['领域层', 'internal/rule（纯函数）', '计价引擎、夜间/高峰判断、优惠、结算，全部可单测'],
      ['数据层', 'internal/model + repository', 'GORM 模型 + 数据访问'],
      ['渠道层', 'internal/channel', '支付渠道抽象：mock / alipay / wechat'],
      ['中间件', 'MySQL / Redis / Kafka', '4 张表存储、缓存、order.paid 事件'],
    ],
    [1200, 3600, 4226],
  ),
  talk('架构上我们严格对齐 go-zero 的标准分层。特别要提一个设计点：把计价和结算这种纯计算逻辑抽成 internal/rule 的纯函数，不依赖数据库，这样每个公式都能写单测、都能被复用——包括我后面演示页也是直接调用这一层，保证演示结果和真实代码 100% 一致。'),
  pageBreak(),
);

// ═══ 第4章 项目一览 ═══
children.push(
  h1('4. 项目一览（三张核心图）'),
  p('答辩时在「图表预览.html」中打开并讲解以下三张图。', {}),
  h2('4.1 全项目业务泳道图'),
  p('乘客端（模块一）→ 订单与派单（模块四）→ 计价与支付（模块五）→ 位置与推送（模块六），一次出行六大模块协作完成。', {}),
  p('模块五占据「价格预估 → 优惠计算 → 支付预下单 → 支付回调 → 司机结算 → 退款」五个泳道节点。', {}),
  note('图表见 图表预览.html 图①'),
  h2('4.2 模块五功能结构图'),
  table(
    ['服务', '功能', 'RPC 接口'],
    [
      ['pricesvc（计价）', '行程价格预估', 'EstimatePrice'],
      ['pricesvc（计价）', '优惠券抵扣计算', 'CalculateDiscount'],
      ['pricesvc（计价）', '实际费用落库', 'SaveActualOrderPrice'],
      ['paysvc（支付）', '支付预下单 / 回调 / 查询', 'CreatePayment / NotifyPayment / GetPayment'],
      ['paysvc（支付）', '退款', 'RefundPayment'],
      ['paysvc（支付）', '司机结算', 'SettleOrder'],
    ],
    [1500, 3300, 4226],
  ),
  h2('4.3 模块五业务闭环流程图'),
  p('行程配置 → 价格预估 → 优惠抵扣 → 创建支付 → 支付回调 → 司机结算 → 退款，共 7 步，也是演示页的完整流程。', {}),
  note('图表见 图表预览.html 图③'),
  talk('模块五虽然只负责计价和支付，但它贯穿整个交易闭环。从乘客下单那一刻起，预估价格、算优惠、发起支付、收到回调、给司机结算，到订单异常退款，全部经过我们的服务。这也是这个模块答辩时最好讲清楚的一点：它既是终点（收钱），也是起点（给司机分钱）。'),
  pageBreak(),
);

// ═══ 第5章 本组负责项目 ═══
children.push(
  h1('5. 本组本月负责项目（模块五范围）'),
  h2('5.1 负责范围'),
  table(
    ['子模块', '文件', '说明'],
    [
      ['计价服务 pricesvc', 'rpc/pricesvc', '计价规则、价格预估、优惠抵扣、费用落库'],
      ['支付服务 paysvc', 'rpc/paysvc', '支付预下单、回调、查询、退款、司机结算'],
      ['数据库 4 张表', 'scripts/sql/migrate/05_trade_module.sql', 'price_rule / order_price / payment / settlement'],
      ['计价规则种子数据', 'scripts/sql/migrate/07_price_rule_seed.sql', '北京三条真实规则'],
      ['接口文档', 'docs/api/模块五-计价与支付接口文档.md', '供全组联调'],
      ['对接说明', 'docs/module5/给成员4的对接说明.md', '与订单模块的协作约定'],
      ['答辩演示页', 'docs/module5/demo/', '真实引擎驱动的可视化演示'],
    ],
    [1700, 3600, 3726],
  ),
  h2('5.2 数据库四张表'),
  table(
    ['表', '关键字段', '用途'],
    [
      ['price_rule', 'city_code、car_type、base_price、per_km_price、per_minute_price、night_start/end、night_surcharge、dynamic_max_factor', '计价规则配置（按城市+车型）'],
      ['order_price', 'order_id、actual_price、base_fee、distance_fee、time_fee、night_fee、dynamic_fee、status', '订单费用明细快照'],
      ['payment', 'payment_no、order_id、amount、channel、status、transaction_id、refund_amount、event_sent', '支付单（唯一 + 幂等）'],
      ['settlement', 'settlement_no、order_id、driver_id、platform_commission、driver_income、status', '司机结算单'],
    ],
    [1300, 4900, 2826],
  ),
  talk('数据库是我们第一天就定稿的。四张表字段都很克制，关键点有两个：金额在数据库里用 decimal(10,2) 存「元」，但接口和计算全部用 int64「分」，两层换算由 common/priceutil 统一处理，这是整个模块防金额精度问题的根基。'),
  pageBreak(),
);

// ═══ 第6章 需求分析 ═══
children.push(
  h1('6. 项目背景与需求分析'),
  h2('6.1 业务背景'),
  p('网约车的计价和支付是交易闭环的核心，也是出问题最敏感的环节。需求来源包括：', {}),
  table(
    ['需求来源', '具体内容'],
    [
      ['乘客端（模块一）', '下单前展示预估价格、选择优惠券、发起支付、查询支付结果'],
      ['订单模块（模块四）', '行程结束后调用实际费用落库、支付成功后回调订单状态'],
      ['组长任务分配', '行程价格预估、优惠券抵扣计算、支付预下单三个基础接口'],
      ['平台运营需求', '夜间附加费、高峰动态调价、优惠门槛与上限、司机结算抽成'],
    ],
    [2100, 6926],
  ),
  h2('6.2 需求分析方法论（拿到原型后怎么做）'),
  p('接到需求/原型图后，按五个维度拆解，确保不遗漏异常场景：', {}),
  table(
    ['维度', '要回答的问题', '在模块五的落地'],
    [
      ['信息架构', '有哪些实体？字段？状态？', '四张表建模，payment 状态机 1→2→4'],
      ['业务主流程', '正常路径怎么走？', '预估→优惠→支付→回调→结算'],
      ['异常流程', '失败/重复/篡改怎么办？', '验签、金额比对、幂等条件更新'],
      ['边界条件', '边界值/时段/金额上限？', '夜间跨天 23:00-05:00、动态调价上限 1.5、退款上限'],
      ['可测试性', '每个规则能不能被验证？', 'rule 纯函数 + 单测 + e2e 客户端'],
    ],
    [1500, 4000, 3526],
  ),
  h2('6.3 拿到原型图后的具体工作清单'),
  p('① 澄清问题（问清金额单位、状态流转、幂等要求）→ ② 拆任务（DDL / proto / rule 引擎 / logic / channel / 测试）→ ③ 排期（按日 check 分天）→ ④ 与上下游模块确认接口契约 → ⑤ 评审（金额算例逐条核对）→ ⑥ 开发（TDD）→ ⑦ 联调验证。', {}),
  h2('6.4 需求分析用到的 skills'),
  p('本项目答辩准备过程中用到的 AI 能力：需求解构（把业务诉求转成接口契约）、流程图生成（mermaid：泳道图/时序图/状态机/甘特图）、代码生成（rule 引擎/单测）、文档产出（接口文档/对接说明）、答辩演示页（前端可视化）。', {}),
  talk('关于需求分析，我总结一句话：原型图只是起点，真正的需求在异常路径里。比如支付回调，主流程谁都会写，但「重复回调」「回调金额被篡改」「并发更新」这些场景才是考验。我在需求分析阶段就专门列了一个异常流检查清单，后面做实现时发现几乎每个坑都在清单里。'),
  pageBreak(),
);

// ═══ 第7章 功能实现路径 ═══
children.push(
  h1('7. 功能实现路径'),
  h2('7.1 开发顺序（TDD 驱动）'),
  table(
    ['阶段', '产出', '验证'],
    [
      ['1 数据库', 'DDL 迁移脚本 + 种子数据', 'SQL 可重复执行（幂等）'],
      ['2 服务骨架', 'pricesvc / paysvc（go-zero zrpc）', 'go build ./... 通过'],
      ['3 计价引擎', 'rule/price_engine.go（纯函数）', '单测断言分项金额'],
      ['4 三个基础 RPC', 'EstimatePrice / CalculateDiscount / CreatePayment', '单测 + gRPC 客户端联调'],
      ['5 支付闭环', 'NotifyPayment / GetPayment / RefundPayment / SettleOrder', 'sqlmock 全链路 e2e 测试'],
      ['6 对接文档', '接口文档 + 给成员4的对接说明', '跨模块契约确认'],
      ['7 演示交付', '答辩演示页（真实引擎）', '现场演示与代码一致'],
    ],
    [1800, 4200, 3026],
  ),
  h2('7.2 核心计算逻辑（全部以「分」int64 计算）'),
  p('计价公式（对应 rule/price_engine.go）：', { bold: true }),
  p('· 起步价：命中规则的 base_price，直接计起步费；', { indent: 360 }),
  p('· 里程费 = (总里程 - 起步包含里程) × 每公里价，仅当超过起步里程，不足 1 公里按比例计；', { indent: 360 }),
  p('· 时长费 = 总时长(秒) × 每分钟价 ÷ 60，不足 1 分钟按比例计；', { indent: 360 }),
  p('· 夜间费：夜间时段（23:00-05:00，支持跨天）加收固定附加费；', { indent: 360 }),
  p('· 动态调价费 = 基础价(起步+里程+时长) × (factor - 1)，factor 受车型动态上限约束，高峰时段自动取车型上限（快车 1.5）；', { indent: 360 }),
  p('· 总价 = 起步 + 里程 + 时长 + 夜间 + 动态。', { indent: 360 }),
  p('优惠公式（对应 rule/discount.go）：固定金额券优惠 = min(面额, 订单总额)；折扣券优惠 = min(订单总额 × (1 - 折扣/100), 最大优惠上限)。', {}),
  p('结算公式（对应 paysvc/rule/settlement.go）：平台抽成 = round(总金额 × 抽成率 / 100)，司机实收 = 总金额 - 抽成。', {}),
  talk('实现路径的核心是把「计算」和「流程」分离。计算全部收敛在 rule 纯函数里，logic 只做编排：取规则、算价、落库、发事件。这样做的好处是——答辩老师现场让我口算金额，我能拿源码里的公式一步步算给他听，因为公式就是代码本身。'),
  pageBreak(),
);

// ═══ 第8章 技术方案 ═══
children.push(
  h1('8. 技术方案（关键技术决策）'),
  table(
    ['决策点', '选型', '理由'],
    [
      ['金额表示', '接口 int64「分」+ DB decimal(10,2)「元」', '避免 float 精度漂移；priceutil 统一换算'],
      ['计算方式', 'rule 纯函数（不依赖 DB）', '可单测、可复用、演示页直接调用'],
      ['微服务框架', 'go-zero zrpc + gRPC', '与项目全组统一；proto 即契约'],
      ['ORM', 'GORM', '与 common/datasource 统一，模型即建表参考'],
      ['消息队列', 'Kafka（IBM/sarama）', '支付成功事件 order.paid 异步解耦'],
      ['支付渠道', 'channel 抽象：mock / alipay SDK / wechat', '密钥缺失降级 mock，沙箱可用走真实支付宝'],
      ['验签', 'smartwalle/alipay v3 SDK（RSA2）', '真实验签而非模拟，防伪造'],
      ['幂等', 'payment_no 唯一 + WHERE status=待支付 条件更新', '重复回调安全；事务内比对金额'],
    ],
    [1700, 3600, 3726],
  ),
  talk('技术方案这里最值得讲的是「分」与「元」的双层表示。很多同学用 float 存钱，跑几次加法就出现 0.30000000000000004 这种问题。我们用 int64 分做计算、decimal 元做存储，两边用 priceutil 换算，整个模块没有一处浮点金额运算。'),
  pageBreak(),
);

// ═══ 第9章 项目时间线 ═══
children.push(
  h1('9. 项目内容与实训计划（时间线）'),
  table(
    ['日期', '实训任务（依据组长日 check）', '本人产出'],
    [
      ['08-13', '计价规则梳理 + DDL + 三个基础接口', '服务骨架、计价引擎、EstimatePrice / CalculateDiscount / CreatePayment、迁移脚本'],
      ['08-14', '支付闭环 + 对接订单状态回调 + 接口文档', 'NotifyPayment / GetPayment / RefundPayment / SettleOrder、给成员4对接说明、接口文档'],
      ['08-17', '全链路联调验证', 'sqlmock e2e 测试 + 真实 gRPC 联调客户端'],
      ['08-27', '答辩演示 + 日 check', '答辩演示页（真实引擎 + 内置降级）、接口文档归档'],
      ['08-28', '答辩准备（本次）', '答辩稿、图表预览、速查手册'],
    ],
    [1200, 4200, 3626],
  ),
  note('甘特图见 图表预览.html 图⑦⑧。'),
  talk('开发节奏完全按实训计划表推进：第一天打通骨架和三个基础接口，第二天补完支付闭环，中间穿插和订单模块的对接，答辩前做了可视化演示页。每一步都有文档和测试兜底，这也是为什么我敢在现场演示——因为每一步都是被验证过的。'),
  pageBreak(),
);

// ═══ 第10章 难点与解决 ═══
children.push(
  h1('10. 难点与解决方案（答辩重点）'),
  h2('难点一：夜间时段跨天判断'),
  p('计价规则里的夜间时段是 23:00-05:00，跨天了。不能简单判断「时间在 8 点以后」，否则 0 点的订单会漏判。', {}),
  p('解决方案（rule/night.go IsNightTime）：把 start/end 解析成「当天第几分钟」，当 start ≥ end 时判定为跨天，走 cur ≥ start 或 cur < end 的规则。23:00-05:00 表示 23:00 至次日 05:00。', {}),
  table(['用例', '结果'], [['23:30 → 跨天命中', '加收夜间费'], ['03:30 → 跨天命中（cur < end）', '加收夜间费'], ['12:00 → 不命中', '不加收'], ['05:00 整 → 不命中（闭开区间）', '不加收']], [3300, 5726]),
  h2('难点二：支付回调幂等与并发安全'),
  p('支付宝回调可能重复投递（网络超时重发），也可能并发到达。如果处理两次，会出现重复入账、重复发事件。', {}),
  p('解决方案（notify_payment_logic.go）：① 先验签；② 事务内读取支付单，若已是「支付成功」直接幂等返回 success；③ 仅「待支付」可流转；④ 用 WHERE id=? AND status=待支付 条件更新做乐观锁，并发下最多一条生效；⑤ 金额比对：回调金额必须等于支付单金额，否则拒绝。', {}),
  h2('难点三：金额精度'),
  p('计价、优惠、结算涉及大量乘法与除法，用 float 会漂移。', {}),
  p('解决方案：全链路 int64「分」+ priceutil.Add 累加 + math.Round 四舍五入；数据库 decimal(10,2)。单测里对 5 个算例逐项断言分项金额。', {}),
  h2('难点四：跨模块协作（订单模块依赖）'),
  p('支付回调成功后要生成司机结算单，需要司机 ID，但司机 ID 在订单模块；同时订单状态「待支付→已完成」需要通知订单模块。', {}),
  p('解决方案：回调成功 → 发 Kafka order.paid 事件（异步解耦）→ 调 ordersvc.GetOrder 取 driver_id → 生成结算单。给成员 4 的对接说明里明确了 GetOrder 实现与消费约定，接口化隔离，不阻塞本模块进度。', {}),
  talk('四个难点我挑两个细讲：夜间跨天和支付幂等。夜间判断大家容易忽略「跨天」这个边界；支付幂等则是生产环境最容易翻车的地方。我们的方案不是简单的「查一下状态」，而是用条件更新做乐观锁，配合金额比对，从数据库层面保证重复回调绝对安全。'),
  pageBreak(),
);

// ═══ 第11章 优化 ═══
children.push(
  h1('11. 项目优化（AI 辅助开发）'),
  table(
    ['维度', '优化前', '优化后', '收益'],
    [
      ['演示方式', '讲 PPT / 贴代码', '真实引擎驱动的交互演示页', '现场可验证、与代码一致'],
      ['演示页计算', 'JS 前端复刻公式', '直接调用 Go 真实 rule 引擎，失败自动降级', '结果 100% 一致、零依赖兜底'],
      ['跨服务复用', 'internal 无法跨服务引用', '服务内建导出适配包（paysvc/demoapi）', '复用真实结算/渠道代码'],
      ['回调可靠性', '发 Kafka 失败可能丢事件', 'outbox-lite：DB 先提交 + event_sent 标记 + 对账补发', '事件不丢、可对账'],
      ['测试', '零散手动验证', 'sqlmock e2e + 单测 + 联调客户端', '全链路可回归'],
    ],
    [1400, 2800, 3400, 1426],
  ),
  talk('优化主要体现在「让演示和真实代码零距离」。答辩演示页不是画出来的假数据，而是启动一个 Go 服务直接调用模块五的计价引擎。这样老师问「这个 76 块 1 毛 3 是怎么来的」，我当场就能用源码公式算给他看。另外回调链路我们做了 outbox-lite 改造，事件不丢、失败可对账，这是从「能用」到「生产可用」的关键优化。'),
  pageBreak(),
);

// ═══ 第12章 测试结果 ═══
children.push(
  h1('12. 测试结果'),
  table(
    ['测试类型', '内容', '结果'],
    [
      ['单元测试（rule 包）', '计价引擎（含夜间/高峰/上限/非法输入）、优惠、结算、退款', '全部通过，金额逐项断言'],
      ['logic 层测试', 'EstimatePrice / CalculateDiscount / SaveActualOrderPrice', '通过'],
      ['sqlmock 全链路 e2e', '创建支付单→回调→自动结算→查询→退款', '通过（无需 MySQL/Kafka）'],
      ['真实 gRPC 联调', 'scripts/e2e 客户端直连 50054 跑全链路', '通过（本地 MySQL）'],
      ['演示页验收', '5 个算例与 rule 引擎输出一致', '通过'],
    ],
    [2100, 4500, 2426],
  ),
  table(
    ['验收算例', '期望（分）', '期望（元）'],
    [
      ['北京快车 12.5km/30min 白天 10:00', '1200+2375+1500=5075', '50.75'],
      ['同上，早高峰 08:00（factor=1.5）', '5075+2538=7613', '76.13'],
      ['北京特惠 5km/10min 夜间 23:30', '800+540+400+500=2240', '22.40'],
      ['8 折券（满20、最高减10）用于 50.75', '抵扣 1000（受上限）', '实付 40.75'],
      ['结算 40.75 × 20%', '平台 815 / 司机 3260', '平台 8.15 / 司机 32.60'],
    ],
    [4000, 2700, 2326],
  ),
  talk('测试上我最看重的是「算例可复现」。上面这张表里的每一行，都可以在演示页里现场复现、也可以在源码单测里找到对应断言。老师问我怎么保证正确性，我说：因为每个金额都有单测兜底，现场还能再跑一遍。'),
  pageBreak(),
);

// ═══ 第13章 上线流程 ═══
children.push(
  h1('13. 发布上线流程（Linux / 多服务）'),
  h2('13.1 服务启动与配置'),
  p('每个服务是独立进程：config（etc/*.yaml）→ svc（ServiceContext 组装 DB/Redis/Channel/Repo）→ server（zrpc 监听）。yaml 里的密钥类配置支持环境变量覆盖。', {}),
  h2('13.2 Linux / 可执行文件 / 多服务开发'),
  p('本地开发与部署方式（Windows / Mac / Linux 通用）：', { bold: true }),
  table(
    ['场景', '命令', '说明'],
    [
      ['直接运行（调试用）', 'go run ./rpc/pricesvc/paysvc.go -f etc/paysvc.yaml', '源码级启动，方便改完即跑'],
      ['Makefile 多服务', 'make run-pricesvc & make run-paysvc &', '项目根 Makefile 已提供所有 RPC 的 run-* target'],
      ['生成可执行文件', 'go build -o bin/pricesvc ./rpc/pricesvc', '输出单一二进制，部署无需带源码'],
      ['Linux 交叉编译', 'GOOS=linux GOARCH=amd64 go build -o bin/pricesvc ./rpc/pricesvc', '在 Windows/Mac 上编出 Linux 可执行文件'],
      ['后台启动 + 日志', 'nohup ./bin/pricesvc -f etc/praysvc.yaml > logs/pricesvc.log 2>&1 &', '生产环境标准做法'],
      ['进程查看/停止', 'ps aux | grep pricesvc；kill <pid>', '或 pkill pricesvc'],
      ['依赖中间件', 'docker compose -f deploy/docker/infra.yml up -d', '一次性拉起 MySQL / Redis / Kafka / etcd'],
    ],
    [2400, 4400, 2226],
  ),
  p('健康检查与验证：', { bold: true }),
  p('· 端口监听：`ss -tlnp | grep 50053` 确认 pricesvc 监听中；', { indent: 360 }),
  p('· gRPC 健康：grpcurl -plaintext 127.0.0.1:50053 list 或直接跑联调客户端；', { indent: 360 }),
  p('· 演示服务：`go run ./rpc/pricesvc/demo` 浏览器访问 http://127.0.0.1:8787/。', { indent: 360 }),
  talk('Linux 与多服务开发上，我们遵循项目统一的 Makefile 约定：每个服务是一个独立进程、一个独立端口。开发期用 go run 直接跑，部署期用交叉编译产 Linux 二进制，进程管理用 nohup + 端口监听做探活。整套流程在 Makefile 里一行命令就能起所有服务。'),
  h2('13.3 依赖中间件（deploy/docker/infra.yml）'),
  p('MySQL、Redis、etcd（服务注册发现）、Kafka 通过 docker compose 一键拉起。', {}),
  h2('13.4 一键联调脚本'),
  p('scripts/e2e/run_pay_e2e.ps1：自动启动 paysvc → 跑联调客户端 → 停止 paysvc，本地可回归验证全链路。', {}),
  talk('上线流程上，我们遵循项目统一的部署约定：中间件容器化、服务独立进程、配置集中管理。本地有 Docker 就能拉起全部依赖，加上一键联调脚本，从拉代码到跑通支付闭环不超过 5 分钟。'),
  pageBreak(),
);

// ═══ 第14章 监控规划 ═══
children.push(
  h1('14. 后期监控与运维规划'),
  table(
    ['监控点', '监控方式', '告警阈值/说明'],
    [
      ['支付单异常状态', '查询 payment 表 status=3（支付失败）', '失败率异常告警'],
      ['事件投递', 'event_sent=false 的支付单', '对账任务补发 order.paid'],
      ['Kafka 积压', 'topic 消费 lag', 'lag 超阈值告警'],
      ['结算异常', 'settlement 生成失败数', '人工介入'],
      ['演示服务', 'demo_server 健康检查', '内置模式自动降级，答辩不翻车'],
    ],
    [2300, 3400, 3326],
  ),
  talk('监控规划这块，我们特别设计了「事件对账」机制：每笔支付单有 event_sent 标记，如果 Kafka 事件没发出去，对账任务会扫描并补发，保证订单模块一定能收到支付结果。这是支付系统的可靠性底线。'),
  pageBreak(),
);

// ═══ 第15章 AI 防幻觉 ═══
children.push(
  h1('15. AI 辅助开发与防幻觉措施'),
  p('本项目大量使用 AI 辅助编码与文档，但所有产出都经过以下八条防幻觉措施验证，确保内容真实可查：', {}),
  table(
    ['#', '措施', '落地示例'],
    [
      ['1', '金额算例逐项断言', '答辩稿附录 B 每个金额都附「分项计算过程」，与单测断言一致'],
      ['2', '接口字段照抄源码', '所有 proto 字段与 rpc/*/proto 一致，不做「记得大概是这样」'],
      ['3', '图表与代码可对账', '演示页直接调真实 rule 引擎，不是 JS 复刻的假数据'],
      ['4', '跨模块依赖留文档指针', '给成员4的对接说明明确 GetOrder/事件契约，不凭印象描述'],
      ['5', '边界诚实', '只讲做过的功能；未做的（优惠券落库、对账报表）明确标注延后'],
      ['6', '代码引用可 grep', '答辩稿引用的函数/接口全部能 git grep 到'],
      ['7', '现场兜底', '演示页内置降级模式，服务挂了也能讲'],
      ['8', '内部问题不包装', 'ordersvc.GetOrder 待成员4补全等依赖如实说明，反而体现协作边界清晰'],
    ],
    [700, 4200, 4126],
  ),
  talk('关于 AI 防幻觉，我的原则是：AI 负责写，人类负责验。具体做法是「三个必须」——必须能 grep 到、必须能跑出结果、必须能口算验证。每个金额我都能在单测和演示页里现场复现，这就是防幻觉最硬的保障。'),
  pageBreak(),
);

// ═══ 第16章 答辩口述稿 ═══
children.push(
  h1('16. 答辩口述稿（30 分钟分段话术）'),
  p('按 30 分钟主讲 + 演示编排。语速 150-180 字/分钟。', {}),

  h2('16.1 开场（0:00-2:00）'),
  talk('各位老师好。我是本组负责「计价与支付」模块的成员。本项目是基于 Go 微服务架构的花小猪打车仿制系统，六大模块协作完成一次出行的完整交易链路。我负责的模块五是整个交易里最敏感的部分：价格怎么算、钱怎么收、钱怎么分。接下来我先讲架构与需求，中间会有一场真实代码驱动的演示，最后留出时间请老师提问。'),
  p('—— 此时切换屏幕到「图表预览.html」，先展示图②模块五功能结构图。', { shade: GREY_FILL }),

  h2('16.2 项目架构与一览（2:00-6:00）'),
  talk('先看项目架构。全项目六个模块，乘客端、司机端、管理后台是三层入口，底层是订单派单、计价支付、位置推送六个微服务，加 Kafka、Redis、MySQL、etcd 做基础设施。我负责的模块五在这张泳道图里占据五个环节：预估、优惠、下单、回调、结算。这也是为什么我说模块五是交易的「收口」。'),
  p('—— 依次展示图①泳道图、图②功能结构图、图③业务流程图。', { shade: GREY_FILL }),

  h2('16.3 需求分析（6:00-10:00）'),
  talk('需求分析上，我拿到原型后的第一件事不是写代码，而是列异常流检查清单。举个例：支付回调这个接口，主流程是「验签、改状态、发事件」，但真正要处理的是重复回调、金额篡改、并发覆盖三种异常。我把它们全列在需求清单里，实现时逐条对应代码。原型图分析我看五个维度：信息架构、主流程、异常流、边界条件、可测试性。每个维度都要求能落地成代码和单测。'),
  p('—— 展示「原型图分析五维度」表（答辩稿 6.2 节）。', { shade: GREY_FILL }),

  h2('16.4 功能实现与技术方案（10:00-14:00）'),
  talk('实现路径是标准的 TDD：先建表、再写纯函数引擎、再包 logic、最后上渠道层。最核心的技术决策是金额用 int64 分、数据库 decimal 元，两层换算统一封装，全模块零浮点金额运算。支付渠道做成抽象层：沙箱密钥齐了走真实支付宝 SDK 验签下单，密钥缺失自动降级 mock，保证任何环境都能演示。'),
  p('—— 展示分层架构图（图④）与关键技术决策表（答辩稿第 8 章）。', { shade: GREY_FILL }),

  h2('16.5 现场演示（14:00-22:00）★核心'),
  talk('下面进入演示环节。我启动模块五的真实演示服务，它直接调用生产代码里的计价引擎。第一步行程配置：我选北京快车，12.5 公里、30 分钟，把时刻切到早高峰 8 点。点预估，大家看五项费用明细：起步 12、里程 23 块 75、时长 15 块，高峰动态调价 25 块 38，合计 76 块 13。这个 25 块 38 就是基础价 50 块 75 乘以 0.5 的溢价，对应源码里 factor=1.5 的动态上限。'),
  p('—— 打开演示页，演示第 1、2 步。', { shade: GREY_FILL }),
  talk('第二步优惠：我用一张 8 折券，门槛 20、最高减 10。大家看，50 块 75 的 8 折优惠本来是 10 块 15，但被上限压到 10 块，实付 40 块 75。这体现的是「先算优惠、再卡上限」的规则顺序。'),
  p('—— 演示第 3 步优惠。', { shade: GREY_FILL }),
  talk('第三步支付：我选支付宝，弹出模拟收银台，确认支付。回调返回，支付单从「待支付」变成「支付成功」。这里关键操作来了——我再点一次「模拟重复回调」，大家看，系统提示「已忽略重复回调」，没有重复入账、没有重复发事件。这就是我们条件更新乐观锁的幂等效果。'),
  p('—— 演示第 4、5 步，重点演示「重复回调」的幂等提示。', { shade: GREY_FILL }),
  talk('第四步结算：40 块 75 按 20% 抽成，平台 8 块 15、司机实收 32 块 6。最后演示退款，全额退款后状态变成「已退款」，累计退款金额正确回写。'),
  p('—— 演示第 6、7 步。', { shade: GREY_FILL }),

  h2('16.6 难点与解决（22:00-25:00）'),
  talk('演示之外，我讲两个最有代表性的难点。一是夜间跨天：规则是 23 点到次日 5 点，不能用简单的时间比较。我们把时段解析成「当天第几分钟」，start 大于 end 就判定为跨天，走「当前分钟 ≥ start 或 < end」的逻辑，0 点的订单也能正确加收夜间费。二是支付幂等，刚才演示已经看到了：重复回调在事务内被条件更新挡住，从数据库层面保证安全。'),
  p('—— 展示时序图（图⑥）讲解回调幂等链路。', { shade: GREY_FILL }),

  h2('16.7 优化与测试（25:00-27:00）'),
  talk('优化方面，最大的改进是把演示从「讲 PPT」变成「真实引擎演示」，并且给演示服务加了内置降级，服务挂了页面也能算。测试方面，全链路有 sqlmock 的 e2e 测试，不需要数据库就能回归；真实环境有一键联调脚本。验收表里五个算例，每一个都能现场复现。'),
  p('—— 展示测试算例表（答辩稿第 12 章）。', { shade: GREY_FILL }),

  h2('16.8 上线与监控（27:00-28:00）'),
  talk('上线流程遵循项目统一约定：中间件 docker compose 一键拉起，服务独立进程，配置集中管理。监控上我们设计了事件对账：每笔支付单有 event_sent 标记，事件没发出去对账任务会自动补发，这是支付可靠性的底线。'),

  h2('16.9 AI 使用与防幻觉（28:00-29:00）—— 简短主动讲，细节留给老师提问'),
  talk('开发过程中我们大量使用 AI 辅助编码与文档。这里我主动只讲一句原则：三个必须——必须能 grep 到、必须能跑出结果、必须能口算验证。每个金额在单测、演示页、源码三处都能对齐，这是防幻觉的硬保障。', {}),
  note('【预留策略】本节主动部分不超过 1 分钟，不展开具体 AI 写了哪些代码，留给老师追问时再展开。老师可能追问："AI 帮你写了多少？""你最担心 AI 什么地方出错？""你怎么验证 AI 写的金额是对的？"——准备好 8 条防幻觉措施对应回答（详见 15 章）。'),

  h2('16.10 结束语（29:30-29:50）'),
  talk('邹老师，我这边讲完了。本次讲解覆盖了项目架构、需求分析、功能实现、技术方案、四个难点、测试算例、上线监控和 AI 防幻觉八个方面，并现场演示了计价与支付的完整闭环。您还有什么需要提问的吗？'),
  p('—— 此刻切换到演示页或图表 HTML 静默等待；目光注视老师，不要先说话。', { shade: GREY_FILL }),

  h2('16.11 自我总结（29:50-30:00）—— 做得好的、存在的不足、态度谦卑'),
  p('【做得好的地方】', { bold: true }),
  table(
    ['维度', '实际表现'],
    [
      ['接口完整度', '七个 RPC（Estimate/Calculate/Save/Creat/Notify/Get/Refund/Settle）全部真实实现，非空壳'],
      ['金额可验证', '五个验收算例在演示页和单测里逐项对齐，老师可现场口算复核'],
      ['跨模块协作', '给订单模块留了 Kafka 事件 + 接口契约，不阻塞本模块进度'],
      ['可视化交付', '演示页直接复用真实引擎，结果与代码 100% 一致'],
    ],
    [1500, 7526],
  ),
  p('【存在的不足（非致命问题）】', { bold: true }),
  table(
    ['不足点', '现状', '改进计划'],
    [
      ['优惠券未落库', '本期作入参传入，模块三优惠券配置联调后再落库', '等模块三上线后接入 coupon 表'],
      ['ordersvc.GetOrder 空壳', '结算 driver_id 由 GetOrder 返回，目前靠 mock', '等成员 4 补全即可打通（已留对接说明）'],
      ['演示服务降级', '已内置 JS 同公式兜底，保证现场不翻车', '已覆盖'],
      ['对账报表未做', '本期仅做事件补发，未做完整对账', '下期接入 reportsvc'],
    ],
    [1700, 4200, 3126],
  ),
  p('【态度谦卑】', { bold: true }),
  talk('以上是我的模块总结。模块五能走到今天离不开组长的任务分派、订单模块同学的协作，以及项目统一规范的约束。我本人能持续改进的地方还有不少，比如测试覆盖率还可以再高一点、跨服务的错误码设计还可以更统一一些，欢迎各位老师多指正。'),
  pageBreak(),
);

// ═══ 第17章 附录A 图表清单 ═══
children.push(
  h1('17. 附录 A：图表清单'),
  table(
    ['#', '图名', '类型', '对应页面'],
    [
      ['①', '全项目业务泳道图', '泳道图', '图表预览.html'],
      ['②', '模块五功能结构图', '结构图', '图表预览.html'],
      ['③', '模块五业务闭环流程', '流程图', '图表预览.html'],
      ['④', '模块五分层架构', '架构图', '图表预览.html'],
      ['⑤', '全项目微服务调用图', '架构图', '图表预览.html'],
      ['⑥', '支付回调时序图', '时序图', '图表预览.html'],
      ['⑦', '模块五开发里程碑', '甘特图', '图表预览.html'],
      ['⑧', '实训计划落地对照', '甘特图', '图表预览.html'],
    ],
    [900, 3600, 1700, 2826],
  ),
);

// ═══ 第18章 附录B 算例表 + Q&A ═══
children.push(
  h1('18. 附录 B：金额算例明细 + 高频 Q&A'),
  h2('B.1 金额算例明细（全部与源码/单测一致）'),
  table(
    ['场景', '分项计算（分）', '合计'],
    [
      ['快车 12.5km/30min 白天', '起步 1200；里程 (12500-3000)×250/1000=2375；时长 1800×50/60=1500', '5075（50.75 元）'],
      ['同上 早高峰 08:00', '基础 5075；动态 round(5075×0.5)=2538', '7613（76.13 元）'],
      ['特惠 5km/10min 夜间', '起步 800；里程 540；时长 400；夜间 500', '2240（22.40 元）'],
      ['8 折券（满20/最高减10）用于 50.75', '原始优惠 5075-5075×80/100=1015 → 受上限取 1000', '实付 4075（40.75 元）'],
      ['结算 40.75 × 20%', '平台 round(4075×20/100)=815', '司机 3260（32.60 元）'],
    ],
    [3300, 4000, 1726],
  ),
  h2('B.2 高频追问 Q&A（20 题）'),
  table(
    ['问题', '参考回答'],
    [
      ['为什么金额用「分」？', 'float 有精度漂移，int64 分无误差；DB 用 decimal(10,2) 存元，priceutil 统一换算'],
      ['夜间跨天怎么判断？', 'IsNightTime：时段解析为当天第几分钟，start≥end 视为跨天，cur≥start 或 cur<end'],
      ['重复回调怎么防？', '事务内读取，已支付直接幂等返回；仅待支付可流转；WHERE id=? AND status=待支付 条件更新防并发'],
      ['回调金额被篡改怎么办？', '金额比对：回调 total_amount_cents 必须等于支付单金额，不一致直接拒绝'],
      ['动态调价的 factor 从哪来？', '高峰时段取车型的 DynamicMaxFactor（快车 1.5），平时 1.0，且有上限兜底'],
      ['为什么优惠券要上限？', '折扣券 MaxDiscountCents，防过度补贴，min(原优惠, 上限)'],
      ['平台补贴为什么是 0？', '本期简化为 0，字段已预留，后续对接活动运营扩展'],
      ['和订单模块怎么协作？', '回调成功发 order.paid 事件（Kafka）+ GetOrder 取 driver_id；契约见给成员4的对接说明'],
      ['事件发失败了怎么办？', 'outbox-lite：DB 先提交 + event_sent=false，对账任务扫描补发'],
      ['ordersvc.GetOrder 是空壳怎么办？', '属于模块四，本模块已接口化 + mock 自测；对接说明已给成员4，实现后即可打通'],
      ['支付渠道为什么 mock？', '无沙箱密钥时降级 mock 保证本地可演示；密钥齐了走真实支付宝 SDK（验签 RSA2）'],
      ['为什么做演示页？', '项目无前端，答辩可视化需要；且直接复用真实引擎，结果与代码一致'],
      ['演示页降级怎么回事？', '启动时探测服务，失败则用内置 JS 同公式计算，保证现场零依赖'],
      ['为什么要 DDL 单独拆 seed？', '建表与初始数据分离，seed 幂等可重复执行，避免联调重复插数据'],
      ['一笔支付能部分退款吗？', '能：校验累计已退+本次≤已支付，部分退保持支付成功，全额退流转已退款'],
      ['退款前提条件？', 'ValidateRefund：状态必须已支付、金额为正、不超过可退余额'],
      ['动态调价会超过上限吗？', '不会：Estimate 里 cappedFactor 超上限被截断到 DynamicMaxFactor'],
      ['时长费不足 1 分钟怎么算？', '按比例：durationS×每分钟分/60，秒级精度折算'],
      ['起步价和起步里程什么关系？', '起步里程内只收起步价，超出部分才按每公里计里程费'],
      ['结算抽成率从哪来？', 'SettleOrder 入参 commission_rate（如 20%），由平台运营配置'],
    ],
    [2600, 6426],
  ),
  note('Q&A 完整版还有 30+ 追问可在答辩前通读接口文档与联调记录，防止临场被问倒。'),
);

// ─────────── 组装文档 ───────────
const doc = new Document({
  styles: {
    default: { document: { run: { font: FONT, size: 21 } } },
    paragraphStyles: [
      { id: 'Heading1', name: 'Heading 1', basedOn: 'Normal', next: 'Normal', quickFormat: true, run: { size: 30, bold: true, font: FONT, color: '1F4E79' }, paragraph: { spacing: { before: 320, after: 160 }, outlineLevel: 0 } },
      { id: 'Heading2', name: 'Heading 2', basedOn: 'Normal', next: 'Normal', quickFormat: true, run: { size: 24, bold: true, font: FONT, color: ORANGE }, paragraph: { spacing: { before: 240, after: 120 }, outlineLevel: 1 } },
      { id: 'Heading3', name: 'Heading 3', basedOn: 'Normal', next: 'Normal', quickFormat: true, run: { size: 21, bold: true, font: FONT }, paragraph: { spacing: { before: 160, after: 80 }, outlineLevel: 2 } },
    ],
  },
  sections: [{
    properties: {
      page: {
        size: { width: 11906, height: 16838 }, // A4
        margin: { top: 1134, right: 1134, bottom: 1134, left: 1134 }, // 约2cm
      },
    },
    footers: {
      default: new Footer({
        children: [new Paragraph({
          alignment: AlignmentType.CENTER,
          children: [new TextRun({ text: '模块五答辩稿 · ', font: FONT, size: 16, color: '999999' }), new TextRun({ children: [PageNumber.CURRENT], font: FONT, size: 16, color: '999999' })],
        })],
      }),
    },
    children,
  }],
});

Packer.toBuffer(doc).then(buffer => {
  fs.mkdirSync(path.dirname(OUT), { recursive: true });
  fs.writeFileSync(OUT, buffer);
  console.log('答辩稿已生成：' + OUT + '（' + (buffer.length / 1024).toFixed(1) + ' KB）');
});
