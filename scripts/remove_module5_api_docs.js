const fs = require('fs');

const files = [
  'c:/Users/hjy/Desktop/XiaoLong-Ridy/docs/module5/模块五接口文档-全组联调版.md',
  'c:/Users/hjy/Desktop/XiaoLong-Ridy/docs/module5/模块五接口文档-全组联调版.docx',
];

for (const f of files) {
  if (fs.existsSync(f)) {
    fs.unlinkSync(f);
    console.log('deleted:', f);
  } else {
    console.log('not found (skip):', f);
  }
}
