const ExcelJS = require('C:\\Users\\hjy\\AppData\\Roaming\\npm\\node_modules\\exceljs');
const path = 'c:\\Users\\hjy\\Desktop\\XiaoLong-Ridy\\docs\\module5\\日check-乔宇翔-2026-08-21.xlsx';

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
ws.getCell('A1').value = '2026年8月21日（日check）';
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
    ['设计', '新增 pricesvc 真实 e2e 联调客户端：EstimatePrice→CalculateDiscount→SaveActualOrderPrice 三接口串联', '是', '进度100%。直连 pricesvc 跑通，参数含 city_code/car_type/distance/duration', '补夜间/高峰动态调价分支联调'],
    ['研发', '更新 pay_e2e_client.go 适配真实 AlipayChannel 联调链路（mock→alipay）', '是', '进度100%。下单/退款/结算全流程打通', '接入真实沙箱密钥跑一次端到端'],
    ['研发', '新增 diag_sign.go 签名诊断工具：用 yaml 私钥生成签名 + 公钥验签', '是', '进度100%。快速定位签名规则/编码问题', '沉淀为 docs/superpowers/排障 SOP'],
    ['测试', 'price_e2e_client.go + pay_e2e_client.go 联合联调：预估→优惠→实付→下单→退款→结算全链路', '是', '进度100%。各接口打印明细校验通过', '加并发场景联调（幂等/重复下单）'],
    ['优化', '计价支付接口文档.md 增补真实联调示例 + AlipayChannel 配置说明', '是', '进度100%。文档与代码同步', '接入 docs/api 模块五接口文档统一入口'],
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
ws.getCell('A9').value = '备注：内容基于今日 8-21 改动整理，覆盖 pricesvc/paysvc 真实 e2e 联调客户端与支付宝签名诊断工具。';
ws.getCell('A9').font = { name: '微软雅黑', size: 10, italic: true, color: { argb: 'FF808080' } };
ws.getCell('A9').alignment = { horizontal: 'left', vertical: 'middle', wrapText: true };

wb.xlsx.writeFile(path).then(() => {
    console.log('日check 已生成：' + path);
});