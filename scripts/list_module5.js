const fs = require('fs');
const d = 'c:/Users/hjy/Desktop/XiaoLong-Ridy/docs/module5/';
for (const f of fs.readdirSync(d)) {
  console.log(f);
}
