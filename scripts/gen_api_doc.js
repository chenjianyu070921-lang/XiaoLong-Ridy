const docx = require('C:\\Users\\hjy\\AppData\\Roaming\\npm\\node_modules\\docx');
const fs = require('fs');

const {
  Document, Packer, Table, TableCell, TableRow, Paragraph, TextRun,
  WidthType, AlignmentType, BorderStyle, HeadingLevel
} = docx;

const YH = 'Microsoft YaHei';

function text(t, o = {}) {
  return new TextRun({ text: t, font: YH, size: 21, bold: !!o.bold, color: o.color });
}

function heading(t, level = 1) {
  const size = level === 1 ? 32 : level === 2 ? 26 : level === 3 ? 24 : 22;
  const spacing = { before: 200, after: 100 };
  return new Paragraph({
    children: [new TextRun({ text: t, font: YH, size, bold: true, color: '1F4E79' })],
    heading: level === 1 ? HeadingLevel.HEADING_1 : level === 2 ? HeadingLevel.HEADING_2 : level === 3 ? HeadingLevel.HEADING_3 : undefined,
    spacing
  });
}

function para(t, o = {}) {
  return new Paragraph({
    children: Array.isArray(t) ? t : [text(t, o)],
    spacing: { after: 80 },
    bullet: o.bullet || false,
    indent: o.indent ? { left: 360 } : undefined
  });
}

function bullet(t) {
  return para(t, { bullet: { level: 0 } });
}

function numbered(t) {
  return para(t, { numbering: { reference: 'num', level: 0 } });
}

function codeBlock(lines) {
  return lines.map(l => new Paragraph({
    children: [new TextRun({ text: l, font: 'Consolas', size: 18, color: '1A1A1A' })],
    spacing: { after: 20 },
    shading: { type: 'clear', fill: 'F5F5F5' },
    indent: { left: 180 }
  }));
}

const headerShading = 'D9E1F2';
const labelShading = 'F2F2F2';
const border = {
  top: { style: BorderStyle.SINGLE, size: 4, color: '7F7F7F' },
  bottom: { style: BorderStyle.SINGLE, size: 4, color: '7F7F7F' },
  left: { style: BorderStyle.SINGLE, size: 4, color: '7F7F7F' },
  right: { style: BorderStyle.SINGLE, size: 4, color: '7F7F7F' },
  insideHorizontal: { style: BorderStyle.SINGLE, size: 2, color: 'BFBFBF' },
  insideVertical: { style: BorderStyle.SINGLE, size: 2, color: 'BFBFBF' }
};

function cell(t, o = {}) {
  const { bold = false, width = 100, colspan = 1, shading = null, align = 'left', size = 20 } = o;
  return new TableCell({
    children: [new Paragraph({
      children: [new TextRun({ text: t, font: YH, size, bold })],
      spacing: { before: 40, after: 40 },
      alignment: align === 'center' ? AlignmentType.CENTER : AlignmentType.LEFT
    })],
    width: { size: width, type: WidthType.PERCENTAGE },
    columnSpan: colspan,
    shading: shading ? { fill: shading } : undefined,
    verticalAlign: 'center'
  });
}

function hcell(t, width = 100) {
  return cell(t, { bold: true, width, shading: headerShading, align: 'center' });
}

// 通用参数表：表头为 字段/类型/必填/说明
function paramTable(rows, widths = [22, 18, 10, 50]) {
  const header = new TableRow({ children: ['字段', '类型', '必填', '说明'].map((h, i) => hcell(h, widths[i])) });
  const body = rows.map(r => new TableRow({
    children: r.map((c, i) => {
      const align = (i === 2) ? 'center' : 'left';
      const bold = (i === 0);
      return cell(String(c), { bold, width: widths[i], align });
    })
  }));
  return new Table({ rows: [header, ...body], width: { size: 100, type: WidthType.PERCENTAGE }, borders: border });
}

// 普通信息表
function infoTable(rows, widths) {
  return new Table({
    rows: rows.map(r => new TableRow({
      children: r.map((c, i) => {
        if (typeof c !== 'object') return cell(c, { width: widths[i] });
        if (c.bold) return cell(c.text, { bold: true, width: widths[i], shading: labelShading });
        return cell(c.text, { width: widths[i] });
      })
    })),
    width: { size: 100, type: WidthType.PERCENTAGE },
    borders: border
  });
}

// ---------------- 文档内容 ----------------
const children = [];

// 封面标题
children.push(new Paragraph({
  children: [new TextRun({ text: '模块五：计价与支付服务接口文档', font: YH, size: 44, bold: true, color: '1F4E79' })],
  alignment: AlignmentType.CENTER,
  spacing: { after: 60 }
}));
children.push(new Paragraph({
  children: [new TextRun({ text: 'pricesvc + paysvc  ·  v1.0  ·  供全组开发联调使用', font: YH, size: 24, color: '595959' })],
  alignment: AlignmentType.CENTER,
  spacing: { after: 300 }
}));

// ================= 一、服务概览 =================
children.push(heading('一、服务概览', 1));
children.push(infoTable([
  [hcell('服务', 16), hcell('目录', 30), hcell('监听端口', 14), hcell('职责', 40)],
  ['pricesvc', 'rpc/pricesvc', '50053', '计价规则、行程价格预估、优惠券抵扣、实际费用落库'],
  ['paysvc', 'rpc/paysvc', '50054', '支付预下单、支付回调、支付查询、退款、司机结算'],
], [16, 30, 14, 40]));
children.push(para(''));
children.push(bullet('通信方式：gRPC（go-zero zrpc），proto 文件见对应 rpc/<svc>/proto/ 目录'));
children.push(bullet('金额单位：接口与计算统一使用「分」(int64)；数据库存 decimal(10,2) 元。以下所有 *_cents 字段均为分'));
children.push(bullet('调用方：订单模块（ordersvc）、派单模块、乘客端网关（api/passenger）等'));

// ================= 二、公共约定 =================
children.push(heading('二、公共约定', 1));
children.push(heading('2.1 错误码', 2));
children.push(infoTable([
  [hcell('错误码', 15), hcell('含义', 30), hcell('说明', 55)],
  ['0', 'OK', '成功'],
  ['400', 'PARAM', '参数错误（如计价参数非法）'],
  ['401', 'NOTLOG', '未登录'],
  ['500', 'FAIL', '服务内部错误'],
], [15, 30, 55]));
children.push(para(''));
children.push(heading('2.2 枚举定义', 2));
children.push(heading('车型 car_type', 3));
children.push(infoTable([
  [hcell('值', 15), hcell('含义', 85)],
  ['1', '特惠快车'],
  ['2', '快车'],
  ['3', '拼车'],
], [15, 85]));
children.push(para(''));
children.push(heading('优惠券类型 CouponType', 3));
children.push(infoTable([
  [hcell('值', 15), hcell('含义', 85)],
  ['1', '固定金额券'],
  ['2', '折扣券（如 8 折 discount=80）'],
], [15, 85]));
children.push(para(''));
children.push(heading('支付渠道 PayChannel', 3));
children.push(infoTable([
  [hcell('值', 15), hcell('含义', 85)],
  ['1', '微信'],
  ['2', '支付宝（已接入真实手机网站支付渠道）'],
  ['3', '余额'],
], [15, 85]));
children.push(para(''));
children.push(heading('支付状态 status', 3));
children.push(infoTable([
  [hcell('值', 15), hcell('含义', 85)],
  ['1', '待支付'],
  ['2', '支付成功'],
  ['3', '支付失败'],
  ['4', '已退款'],
], [15, 85]));

// ================= 三、pricesvc =================
children.push(heading('三、pricesvc 计价服务', 1));
children.push(para('服务地址：127.0.0.1:50053。以下接口按 gRPC 调用，字段名与 proto 对齐。'));

// 3.1 EstimatePrice
children.push(heading('3.1 EstimatePrice — 行程价格预估', 2));
children.push(para('按计价规则 + 里程/时长估算费用，返回各分项费用与总价。'));
children.push(heading('请求参数', 3));
children.push(paramTable([
  ['user_id', 'int64', '是', '用户ID（本期不参与计价，仅透传）'],
  ['city_code', 'string', '是', '城市编码，如 110000（北京），须在 price_rule 表存在'],
  ['car_type', 'int32', '是', '车型：1特惠快车 2快车 3拼车'],
  ['distance_m', 'int64', '是', '预估里程（米），≥0'],
  ['duration_s', 'int64', '是', '预估时长（秒），≥0'],
  ['timestamp', 'int64', '否', '预估时刻（Unix 秒），用于判断夜间/高峰；0 表示当前时间'],
]));
children.push(para(''));
children.push(heading('响应字段', 3));
children.push(paramTable([
  ['price_rule_id', 'int64', '-', '命中的计价规则ID'],
  ['detail.base_fee_cents', 'int64', '-', '起步价（分）'],
  ['detail.distance_fee_cents', 'int64', '-', '里程费（分）'],
  ['detail.time_fee_cents', 'int64', '-', '时长费（分）'],
  ['detail.night_fee_cents', 'int64', '-', '夜间附加费（分）'],
  ['detail.dynamic_fee_cents', 'int64', '-', '动态调价费（分，含高峰溢价）'],
  ['detail.total_cents', 'int64', '-', '合计（分，不含优惠）'],
  ['total_cents', 'int64', '-', '预估总价（分）'],
]));
children.push(para(''));
children.push(heading('请求示例', 3));
children.push(...codeBlock([
  '{',
  '  "user_id": 1,',
  '  "city_code": "110000",',
  '  "car_type": 1,',
  '  "distance_m": 8000,',
  '  "duration_s": 1200,',
  '  "timestamp": 0',
  '}',
]));
children.push(para(''));
children.push(heading('计价公式', 3));
children.push(bullet('里程费 = max(0, 总里程 - 起步包含里程) × 每公里价（不足 1 公里按比例计）'));
children.push(bullet('时长费 = 总时长(秒) × 每分钟价 / 60（不足 1 分钟按比例计）'));
children.push(bullet('夜间费 = 夜间时段(如 23:00-05:00) 加收固定夜间附加费'));
children.push(bullet('动态调价 = 基础价(起步+里程+时长) × (factor-1)，高峰时段(早7-9/晚17-19) factor 自动上调至 1.3，且受 DynamicMaxFactor 上限约束'));
children.push(bullet('总价 = 起步价 + 里程费 + 时长费 + 夜间费 + 动态调价费'));
children.push(para(''));
children.push(para('示例算例：起步8元/2km、每公里1.8元、每分钟0.4元，里程8km、时长20min → 8 + (8-2)×1.8 + 20×0.4 = 8 + 10.8 + 8 = 26.8 元', { italic: true }));

// 3.2 CalculateDiscount
children.push(heading('3.2 CalculateDiscount — 优惠券抵扣计算', 2));
children.push(para('根据优惠券计算折扣金额与实付金额。优惠券本期作为入参传入，不落库（待对接优惠券表后改造）。'));
children.push(heading('请求参数', 3));
children.push(paramTable([
  ['order_id', 'int64', '是', '订单ID'],
  ['total_cents', 'int64', '是', '抵扣前订单总金额（分）'],
  ['coupon.coupon_id', 'int64', '否', '优惠券ID（本期透传）'],
  ['coupon.type', 'int32', '是', '优惠券类型：1固定金额 2折扣'],
  ['coupon.face_value_cents', 'int64', '否', '固定金额券面额（分），type=1 时必填'],
  ['coupon.discount', 'int32', '否', '折扣券折扣（80=8折），type=2 时必填，取值 1~99'],
  ['coupon.threshold_cents', 'int64', '否', '使用门槛（分），0 表示无门槛'],
  ['coupon.max_discount_cents', 'int64', '否', '折扣券最大优惠金额（分），0 表示不限制'],
]));
children.push(para(''));
children.push(heading('响应字段', 3));
children.push(paramTable([
  ['discount_amount_cents', 'int64', '-', '优惠券抵扣金额（分）'],
  ['platform_subsidy_cents', 'int64', '-', '平台补贴金额（分，本期固定 0）'],
  ['payable_amount_cents', 'int64', '-', '乘客实付金额（分）= total - discount - subsidy'],
]));
children.push(para(''));
children.push(heading('请求示例', 3));
children.push(...codeBlock([
  '{',
  '  "order_id": 1001,',
  '  "total_cents": 2950,',
  '  "coupon": {',
  '    "type": 2,',
  '    "discount": 80,',
  '    "threshold_cents": 1000,',
  '    "max_discount_cents": 500',
  '  }',
  '}',
]));
children.push(para(''));
children.push(heading('优惠规则', 3));
children.push(bullet('门槛校验：订单总额 ≥ 门槛才可用，否则返回错误'));
children.push(bullet('固定金额券：优惠 = min(面额, 订单总额)，不超过订单总额'));
children.push(bullet('折扣券：优惠 = min(订单总额×(1-折扣/100), 最大优惠金额)'));
children.push(bullet('优惠不为负、不超过订单总额；联动回写 order_price（若已存在）'));
children.push(para(''));
children.push(para('示例算例：订单 29.5 元，用 8 折券（门槛10元、上限5元）→ 优惠 min(29.5×20%, 5) = 5 元，实付 24.5 元'));

// 3.3 SaveActualOrderPrice
children.push(heading('3.3 SaveActualOrderPrice — 实际费用落库', 2));
children.push(para('行程结束时由订单模块调用，将实际费用快照写入 order_price（已存在则更新，不存在则新建），状态置「已确认」。幂等：按 order_id 唯一。'));
children.push(heading('请求参数', 3));
children.push(paramTable([
  ['order_id', 'int64', '是', '订单ID（唯一键）'],
  ['price_rule_id', 'int64', '否', '命中的计价规则ID（0 表示无）'],
  ['actual_price_cents', 'int64', '是', '实际总价（分）'],
  ['detail', 'PriceDetail', '否', '费用明细（分），可为空（仅落总价）'],
]));
children.push(para(''));
children.push(heading('响应字段', 3));
children.push(paramTable([
  ['success', 'bool', '-', '是否成功'],
  ['order_price_id', 'int64', '-', 'order_price 记录ID'],
]));

// ================= 四、paysvc =================
children.push(heading('四、paysvc 支付服务', 1));
children.push(para('服务地址：127.0.0.1:50054。支付链路：CreatePayment（预下单）→ 支付宝支付 → NotifyPayment（回调）→ 触发结算。'));

// 4.1 CreatePayment
children.push(heading('4.1 CreatePayment — 支付预下单', 2));
children.push(para('创建支付单并调第三方下单（支付宝渠道已接入真实手机网站支付），返回支付参数。'));
children.push(heading('请求参数', 3));
children.push(paramTable([
  ['order_id', 'int64', '是', '订单ID'],
  ['user_id', 'int64', '是', '支付用户ID'],
  ['amount_cents', 'int64', '是', '支付金额（分），通常为优惠后实付'],
  ['channel', 'int32', '是', '支付渠道：1微信 2支付宝 3余额'],
]));
children.push(para(''));
children.push(heading('响应字段', 3));
children.push(paramTable([
  ['payment_id', 'int64', '-', '支付单ID'],
  ['payment_no', 'string', '-', '平台支付单号'],
  ['transaction_id', 'string', '-', '第三方流水号（支付宝预下单阶段为空，回调后回填）'],
  ['pay_params', 'string', '-', '支付参数：支付宝渠道为跳转链接；mock 渠道为模拟串'],
  ['status', 'int32', '-', '状态：1待支付'],
]));
children.push(para(''));
children.push(heading('请求示例', 3));
children.push(...codeBlock([
  '{',
  '  "order_id": 1001,',
  '  "user_id": 10,',
  '  "amount_cents": 2450,',
  '  "channel": 2',
  '}',
]));

// 4.2 NotifyPayment
children.push(heading('4.2 NotifyPayment — 支付回调', 2));
children.push(para('处理第三方支付结果通知：验签 → 更新支付单 → 发 order.paid Kafka 事件 → 触发司机结算。'));
children.push(heading('请求参数', 3));
children.push(paramTable([
  ['payment_no', 'string', '是', '平台支付单号（商户订单号）'],
  ['trade_status', 'string', '是', '交易状态：TRADE_SUCCESS / TRADE_CLOSED 等'],
  ['transaction_id', 'string', '否', '第三方流水号'],
  ['sign', 'string', '否', '签名'],
  ['sign_type', 'string', '否', '签名类型 RSA2'],
  ['paid_at', 'int64', '否', '支付时间（Unix 秒）'],
  ['notify_raw', 'string', '是', '原始回调参数字符串（验签用）'],
]));
children.push(para(''));
children.push(heading('响应字段', 3));
children.push(paramTable([
  ['success', 'bool', '-', '是否处理成功'],
  ['message', 'string', '-', '处理说明'],
]));

// 4.3 GetPayment
children.push(heading('4.3 GetPayment — 支付查询', 2));
children.push(para('按支付单号（优先）或订单ID查询支付状态。'));
children.push(heading('请求参数', 3));
children.push(paramTable([
  ['payment_no', 'string', '否', '按支付单号查（优先）'],
  ['order_id', 'int64', '否', '或按订单ID查'],
]));
children.push(para(''));
children.push(heading('响应字段', 3));
children.push(paramTable([
  ['payment_id', 'int64', '-', '支付单ID'],
  ['payment_no', 'string', '-', '支付单号'],
  ['order_id', 'int64', '-', '订单ID'],
  ['amount_cents', 'int64', '-', '支付金额（分）'],
  ['channel', 'string', '-', '支付渠道（wechat/alipay/balance）'],
  ['status', 'int32', '-', '1待支付 2支付成功 3支付失败 4已退款'],
  ['transaction_id', 'string', '-', '第三方流水号'],
  ['refund_amount_cents', 'int64', '-', '已退款金额（分）'],
]));

// 4.4 RefundPayment
children.push(heading('4.4 RefundPayment — 退款', 2));
children.push(para('校验并执行退款，回写支付单。全额退款流转为「已退款」，部分退款保持「支付成功」。'));
children.push(heading('请求参数', 3));
children.push(paramTable([
  ['payment_no', 'string', '是', '支付单号'],
  ['refund_amount_cents', 'int64', '是', '退款金额（分），须 ≤ 可退金额'],
  ['reason', 'string', '否', '退款原因'],
]));
children.push(para(''));
children.push(heading('响应字段', 3));
children.push(paramTable([
  ['success', 'bool', '-', '是否成功'],
  ['refund_no', 'string', '-', '退款单号'],
  ['refunded_amount_cents', 'int64', '-', '累计已退款金额（分）'],
]));

// 4.5 SettleOrder
children.push(heading('4.5 SettleOrder — 司机结算', 2));
children.push(para('计算平台抽成与司机实收，写入结算单。'));
children.push(heading('请求参数', 3));
children.push(paramTable([
  ['order_id', 'int64', '是', '订单ID'],
  ['driver_id', 'int64', '是', '司机ID'],
  ['total_amount_cents', 'int64', '是', '订单实际总金额（分）'],
  ['commission_rate', 'double', '是', '平台抽成比例（%），如 20 表示 20%'],
]));
children.push(para(''));
children.push(heading('响应字段', 3));
children.push(paramTable([
  ['settlement_id', 'int64', '-', '结算单ID'],
  ['settlement_no', 'string', '-', '结算单号'],
  ['platform_commission_cents', 'int64', '-', '平台抽成（分）'],
  ['driver_income_cents', 'int64', '-', '司机实收（分）= total - commission'],
]));

// ================= 五、调用约定 =================
children.push(heading('五、调用方式示例', 1));
children.push(heading('5.1 gRPC 直连（Go）', 2));
children.push(...codeBlock([
  '// 以 pricesvc 为例',
  'client := price.NewPrice(zrpc.MustNewClient(zrpc.RpcClientConf{Target: "127.0.0.1:50053"}))',
  'resp, err := client.EstimatePrice(ctx, &price.EstimatePriceRequest{',
  '    UserId: 1, CityCode: "110000", CarType: 1,',
  '    DistanceM: 8000, DurationS: 1200,',
  '})',
]));
children.push(para(''));
children.push(heading('5.2 乘客端 HTTP 网关（模块一接入参考）', 2));
children.push(paramTable([
  ['路由', '方法', '调用 paysvc'],
  ['POST /api/passenger/v1/pay/create', '支付预下单', 'CreatePayment'],
  ['POST /api/passenger/v1/pay/query', '支付状态查询', 'GetPayment'],
], [45, 25, 30]));

// ================= 六、注意事项 =================
children.push(heading('六、注意事项', 1));
children.push(bullet('金额一律传「分」(int64)，勿在前端/网关做浮点换算，避免精度丢失'));
children.push(bullet('city_code/car_type 必须命中启用的计价规则，否则 EstimatePrice 返回「计价规则不存在」'));
children.push(bullet('支付宝渠道 pay_params 为跳转链接，需由客户端拉起浏览器/WebView 完成支付'));
children.push(bullet('支付回调 notifyUrl 部署后可访问后配置，当前沙箱环境密钥已就绪'));
children.push(bullet('CreatePayment 后订单状态流转由订单模块消费 order.paid 事件完成（非模块五职责）'));

const doc = new Document({
  numbering: { config: [{ reference: 'num', levels: [{ level: 0, format: 'decimal', text: '%1.', alignment: AlignmentType.LEFT }] }] },
  sections: [{
    properties: { page: { margin: { top: 1200, right: 1200, bottom: 1200, left: 1200 } } },
    children
  }]
});

const outputPath = 'c:\\Users\\hjy\\Desktop\\XiaoLong-Ridy\\docs\\module5\\模块五接口文档-全组联调版.docx';
Packer.toBuffer(doc).then(buffer => {
  fs.writeFileSync(outputPath, buffer);
  console.log('接口文档已生成：' + outputPath);
});
