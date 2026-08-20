const fs = require('fs');

const src = 'c:/Users/hjy/Desktop/XiaoLong-Ridy/docs/module5/模块五接口文档-全组联调版.md';
const dst = 'c:/Users/hjy/Desktop/XiaoLong-Ridy/docs/api/模块五-计价与支付接口文档.md';

const content = fs.readFileSync(src, 'utf8');
fs.writeFileSync(dst, content, 'utf8');
console.log('copied to docs/api:', fs.existsSync(dst), fs.statSync(dst).size);
