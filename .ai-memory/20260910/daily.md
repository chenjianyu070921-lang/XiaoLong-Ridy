## [10:09] - [Bug修复/编码治理]: 司机端全量乱码扫描，修复 1 处 UTF-8 BOM

- **文件**: web/driver/scripts/check-driver-mine.mjs（移除 BOM，备份 .bak_bom）
- **决策**: 司机端源码统一 UTF-8 无 BOM；运行日志（dev*.log）不属于源码，仅检测不修改
- **验证**: Python 扫描脚本全量重扫 229 文件 0 问题；node --check 语法通过


## 今日日 check（2026-09-10，基于 git 提交 bb8289b「聊天大修」19:39:32 + 会话工作；修订版：所有格子不填空）

| 类型 | 任务 | 是否提交 | 进度评估及问题建议(抽查) | 追踪(只追踪建议修改) |
| --- | --- | --- | --- | --- |
| 设计 | 无独立设计文档产出；聊天推送方案确定：司机侧复用既有 driver:push 通道实时推送，乘客侧实时推送列为 Phase 2 | 是（方案随 bb8289b 落地） | 方案与既有架构一致，未新增组件；司机侧推送已实现，乘客侧未动 | 乘客侧实时推送方案待设计（Phase 2） |
| 研发 | 1. 聊天大修：get_conversation_logic.go 订单终态回写会话状态（SendMessage 据此禁发消息）+ readonly 不再被丢弃 + ordersvc 调用补 3s 超时 + 对端名字兜底 2. send_message_logic.go 新增 publishToDriver：乘客消息经 Redis driver:push 实时推送到司机 WS 3. api/chat：chat_handler.go 更新 + 新增 chat_local/chat_test/chat_test2.yaml 多环境配置 4. 前端聊天联调：DriverChatPanel.vue / DriverOrderChatPage.vue / chat.js 5. 编码治理：司机端 229 文件乱码扫描，修复 check-driver-mine.mjs UTF-8 BOM | 是（bb8289b） | 聊天核心逻辑已落地并提交（+363/-886）；BOM 修复后全量重扫 0 乱码；helper.go 抽取 + helper_test.go 109 行单测入库 | 前端 msgType 裸数字未在本次提交内确认是否统一，待核对 |
| 测试 | helper_test.go 新增 109 行单测；前端构建通过；node --check 语法校验通过；乱码全量重扫 0 问题 | 是（bb8289b） | 单测已入库；聊天联调依赖 chat 后端服务就绪 | 待 chat 服务启动后验证会话接口与实时推送链路 |
| 优化 | 清理 8 个 tmp_* 调试文件 + chatdiag/findtables 工具 + 孤儿代码（get_or_create_conversation_logic.go / chat_model.go）+ 空文件 dev.err；多环境配置拆分；会话接口防挂起 3s 超时 | 是（bb8289b） | 调试垃圾清零，仓库净减 886 行 | 保持单 Vite 实例；临时锁文件/日志属运行产物不再入库 |
| 部署 | 无部署动作；本地环境状态：前端 dev:5175 监听中、driversvc:50055 监听中、api/driver:18082 未启动 | 否 | 聊天大修后 api/driver 未重启，聊天接口联调需先启动该服务 | 启动 api/driver（18082）后验证聊天接口；上线前需过 PRR 检查 |
