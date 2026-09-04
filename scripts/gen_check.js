const ExcelJS = require('C:\\Users\\hjy\\AppData\\Roaming\\npm\\node_modules\\exceljs');
const path = 'c:\\Users\\hjy\\Desktop\\XiaoLong-Ridy\\docs\\module5\\日check-乔宇翔-2026-09-03.xlsx';

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
ws.getCell('A1').value = '2026年9月3日（日check）';
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
    ['设计', '梳理 PriceRuleRepo.FindActive 规则命中：status=1 + car_type 匹配 + (city_code 精确 OR 空全局)，Order("city_code DESC") 让城市规则优先', '是', '进度100%。能讲清「非空 city_code 排前」的排序技巧', '补讲规则版本生效时间 effective_at/expire_at'],
    ['研发', '梳理 PaymentRepo.FindUnsettledPaidPayments：LEFT JOIN settlement ON s.order_id=p.order_id + WHERE s.id IS NULL 找未结算单', '是', '进度100%。能讲清 LEFT JOIN + 空值判断替代 NOT EXISTS', '补讲同订单多支付单时的关联准确性'],
    ['研发', '梳理 UpdateSelective 防覆盖：Updates(map) 只更指定列，避免 Save 全字段覆盖丢字段/created_at 清零/decimal 空值污染', '是', '进度100%。能讲清条件更新与全量 Save 的区别及使用场景', '补讲事务内再套条件更新的并发安全'],
    ['设计', '梳理 PriceRule model 字段：价格 decimal(10,2)/时段 time 指针/动态因子 decimal(3,2)，与 OrderPrice/Payment/Settlement 金额字段设计对齐', '是', '进度100%。能讲清 decimal 精度与小数字段类型规范', '补讲 index 索引与唯一键设计'],
    ['研发', '梳理 PaymentRepo 查询族：FindByPaymentNo/FindByOrderId(id DESC 取最新)/FindUnsentPaidPayments(status+event_sent 双条件)', '是', '进度100%。能讲清每个查询对应对账/回调/结算哪个业务方', '补讲分页与 limit 参数化'],
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
ws.getCell('A9').value = '备注：今日围绕 repository 层 SQL 技巧（城市优先/LEFT JOIN/条件更新）与 model 字段设计做代码业务梳理。';
ws.getCell('A9').font = { name: '微软雅黑', size: 10, italic: true, color: { argb: 'FF808080' } };
ws.getCell('A9').alignment = { horizontal: 'left', vertical: 'middle', wrapText: true };

wb.xlsx.writeFile(path).then(() => {
    console.log('日check 已生成：' + path);
});