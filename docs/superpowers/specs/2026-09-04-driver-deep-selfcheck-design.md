# 司机端深度自检设计（2026-09-04）

## 背景
前序验证已确认：司机端前端（web/driver）可编译可启动（Vite :5175）；后端 rpc/driversvc（:50055）与 api/driver（:18082）可编译可启动，且远程 MySQL/Redis 依赖可达。`make run-drversvc` 拼写错误（正确为 `run-driversvc`），且它只启动后端 RPC，不启动前端。

本次目标：在"能编译、能启动"之上，做**更深一层**的验证——找出真实的潜在问题（不只是能否跑起来），覆盖全栈静态复查 + 端到端实测，最终产出《司机端深度自检报告》。

## 范围
- 前端：web/driver（Vue 组件、composables、api 封装、store、router、config）
- 后端：rpc/driversvc + api/driver（Handler/Logic/配置/接口契约）
- 端到端：后端链路实测（启动 driversvc + driver-api，用项目自带 check 脚本 / curl 跑核心流程）；前端交互逻辑以静态审查为主（本环境无法真实浏览器点击）

## 自检方法
### A. 全栈静态复查
前端维度：运行时崩溃（空值/异步未判空/DOM 未挂载）、状态机反向、坐标/金额单位/分页边界、竞态（地图重复初始化、重复请求、tab 切换容器丢失）、错误处理吞错、接口字段与 api/driver 契约一致性（重点复查本次改动涉及的订单/认证/校验）。
后端维度：启动/配置/签名校验、gRPC 地址一致性、Handler/Logic 错误处理（含本次改动 error.go/order_handler/certification_logic/validate.go）、前后端接口字段契约一致性。

### B. 端到端实测
1. 启动 rpc/driversvc（make run-driversvc, :50055）
2. 启动 api/driver（make run-driver-api, :18082）
3. 用 scripts/check-driver-*.mjs 与 curl 执行核心流程：登录、上线/下线、上报位置、接单、结束行程、收入查询
4. 记录哪些流程跑通、哪些报错，定位根因

## 交付物
《司机端深度自检报告》含：
1. 启动验证结论（前端+后端编译/启动/端口/依赖可达）
2. 全栈静态问题清单（文件:行号 / 现象 / 根因 / 严重程度 / 建议）
3. 端到端实测结论（核心流程跑通/报错情况）
4. 总评与优先级建议

## 约束
- 只读，不改动任何业务代码
- 本地实测用完即关停进程、释放端口
- 如实标注无法验证的部分
