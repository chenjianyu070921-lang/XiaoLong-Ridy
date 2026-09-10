## [10:09] - [Bug修复/编码治理]: 司机端全量乱码扫描，修复 1 处 UTF-8 BOM

- **文件**: web/driver/scripts/check-driver-mine.mjs（移除 BOM，备份 .bak_bom）
- **决策**: 司机端源码统一 UTF-8 无 BOM；运行日志（dev*.log）不属于源码，仅检测不修改
- **验证**: Python 扫描脚本全量重扫 229 文件 0 问题；node --check 语法通过
