const ExcelJS = require('C:\\Users\\hjy\\AppData\\Roaming\\npm\\node_modules\\exceljs');
const path = 'c:\\Users\\hjy\\Desktop\\XiaoLong-Ridy\\docs\\module5\\日check-乔宇翔-2026-08-26.xlsx';

const wb = new ExcelJS.Workbook();
const ws = wb.addWorksheet('日check', { views: [{ showGridLines: false }] });

const thin = { style: 'thin', color: { argb: 'FF7F7F7F' } };
const border = { top: thin, left: thin, bottom: thin, right: thin };
const headerFill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFD9E1F2' } };
const titleFill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFDDEBF7' } };
const grayFill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFF2F2F2' } };
const greenFill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFE2EFDA' } };
const yellowFill = { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFF2CC' } };

const titleFont = { name: '微软雅黑', size: 16, bold: true, color: { argb: 'FF1F4E79' } };
const headerFont = { name: '微软雅黑', size: 11, bold: true, color: { argb: 'FF1F4E79' } };
const cellFont = { name: '微软雅黑', size: 10.5 };
const labelFont = { name: '微软雅黑', size: 10.5, bold: true };

ws.columns = [
    { width: 10 },
    { width: 46 },
    { width: 12 },
    { width: 50 },
    { width: 40 },
];

function setBorder(r, c) { ws.getCell(r, c).border = border; }

ws.mergeCells('A1:E1');
ws.getCell('A1').value = '2026年8月26日（日check）';
ws.getCell('A1').font = titleFont;
ws.getCell('A1').alignment = { horizontal: 'center', vertical: 'middle' };
ws.getCell('A1').fill = titleFill;
ws.getRow(1).height = 34;
for (let c = 1; c <= 5; c++) setBorder(1, c);

ws.mergeCells('A2:B2');
ws.getCell('A2').value = '编号：3';
ws.getCell('A2').font = labelFont;
ws.getCell('A2').alignment = { horizontal: 'center', vertical: 'middle' };
ws.getCell('A2').fill = grayFill;
ws.mergeCells('C2:E2');
ws.getCell('C2').value = '姓名：乔宇翔    Git仓库：https://github.com/chenjianyu070921-lang/XiaoLong-Ridy.git';
ws.getCell('C2').font = cellFont;
ws.getCell('C2').alignment = { vertical: 'middle', wrapText: true };
ws.getRow(2).height = 26;
for (let c = 1; c <= 5; c++) setBorder(2, c);

const headers = ['类型', '任务', '是否提交', '进度评估及问题建议（抽查）', '追踪（只追踪建议修改）'];
headers.forEach((h, i) => {
    const col = i + 1;
    ws.getCell(3, col).value = h;
    ws.getCell(3, col).font = headerFont;
    ws.getCell(3, col).fill = headerFill;
    ws.getCell(3, col).alignment = { horizontal: 'center', vertical: 'middle', wrapText: true };
    setBorder(3, col);
});
ws.getRow(3).height = 30;

const data = [
    ['计价', '梳理 pricesvc 三接口的入参/出参契约与 proto 字段语义（EstimatePriceRequest/Response、Coupon 子消息）', '是', '进度100%。能口述每个字段含义、cents 单位约定', '补讲 EstimatePriceResponse.Detail 6 个分项字段'],
    ['计价', '梳理 pricesvc repository 层：PriceRuleRepo.FindActive 按 city_code+car_type 查规则，OrderPriceRepo.FindByOrderId 拉明细', '是', '进度100%。能讲清仓库层与 model.TableName 的对应关系', '补讲 GORM decimal 字段读写的精度处理'],
    ['支付', '梳理 paysvc 退款业务规则：仅 status=2 已支付可退，refunded_cents 累计，全额流转 status=3 已退款', '是', '进度100%。能讲清 rule.ValidateRefund 的边界校验', '补讲重复退款请求幂等'],
    ['支付', '梳理结算业务：CalcSettlement(commission_rate) 拆平台抽成与司机实收，settlement_no 单号规范 SET+时间戳', '是', '进度100%。能讲清 settlement 表字段与 model.SettlementStatus 常量', '补讲司机提现链路差异'],
    ['联调', '梳理跨模块联调链路：ordersvc.FinishTrip → pricesvc.SaveActualOrderPrice → paysvc.CreatePayment → NotifyPayment → SettleOrder', '是', '进度100%。能口述各模块在订单/支付事件中的角色', '补讲 event_reconcile_job 对账补偿触发时机'],
];

data.forEach((row, idx) => {
    const r = idx + 4;
    ws.getCell(r, 1).value = row[0];
    ws.getCell(r, 1).font = labelFont;
    ws.getCell(r, 1).alignment = { horizontal: 'center', vertical: 'middle' };
    ws.getCell(r, 1).fill = grayFill;

    ws.getCell(r, 2).value = row[1];
    ws.getCell(r, 2).font = cellFont;
    ws.getCell(r, 2).alignment = { vertical: 'middle', wrapText: true };

    ws.getCell(r, 3).value = row[2];
    ws.getCell(r, 3).font = cellFont;
    ws.getCell(r, 3).alignment = { horizontal: 'center', vertical: 'middle' };
    ws.getCell(r, 3).fill = row[2] === '是' ? greenFill : yellowFill;

    ws.getCell(r, 4).value = row[3];
    ws.getCell(r, 4).font = cellFont;
    ws.getCell(r, 4).alignment = { vertical: 'middle', wrapText: true };

    ws.getCell(r, 5).value = row[4];
    ws.getCell(r, 5).font = cellFont;
    ws.getCell(r, 5).alignment = { vertical: 'middle', wrapText: true };

    for (let c = 1; c <= 5; c++) setBorder(r, c);
    ws.getRow(r).height = 52;
});

ws.mergeCells('A9:E9');
ws.getCell('A9').value = '备注：今日围绕模块代码业务做接口级提问清单与跨模块联调演练，准备答辩代码演示。';
ws.getCell('A9').font = { name: '微软雅黑', size: 10, italic: true, color: { argb: 'FF808080' } };
ws.getCell('A9').alignment = { horizontal: 'left', vertical: 'middle', wrapText: true };

wb.xlsx.writeFile(path).then(() => {
    console.log('日check 已生成：' + path);
});