# 工作区缓存清理设计

## 目标

清理可再生的本地构建缓存、测试产物和前端依赖，减少未跟踪文件数量与磁盘占用；不修改任何业务源码、数据库文件、接口定义或已有 Git 暂存内容。

## 清理范围

- `.gotmp/`：本地测试日志、临时 YAML、二进制与 Go 缓存。
- `.tmp-gocache/`、`.tmp-gocache-review/`：本地 Go 构建与测试缓存。
- `api/admin/.codex-gocache/`、`rpc/adminsvc/.codex-gocache/`：工具运行产生的 Go 缓存。
- `web-admin/node_modules/`、`web-admin/.vite/`：可由包管理器与 Vite 重新生成的依赖和缓存。

## 保护措施

- 不执行 `git reset`、`git checkout`、`git clean` 或任何暂存区修改操作。
- 不删除 `package.json`、锁文件、Go 模块文件、业务源码、SQL 迁移或文档。
- 更新 `.gitignore`，防止同类生成物再次成为未跟踪文件。

## 验证

清理后核对目标目录不存在，并检查 Git 暂存区中的既有修改仍保持不变。不运行会重新生成缓存的构建或测试命令。
