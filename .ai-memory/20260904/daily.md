# 2026-09-04 日志

## [09:17] - 重构: 司机端报错提示统一收口到拦截器，消除三层重复提示

- **文件**: `web/driver/src/utils/safe-request.js`（新增）、`api/request.js`（保持为唯一提示点）、`api/driver.js`、`composables/useDriverAssets.js`、`views/DriverHome.vue`、`views/DriverLogin.vue`、`views/DriverProfileEdit.vue`、`components/driver-home/DriverReviewsPanel.vue`、`stores/driver.js`
- **决策**: 采用「拦截器为唯一提示点」方案（备选：safeApiCall 为唯一提示点 / 双层标记位协调）。
  选择理由是失败模式更安全——漏改一处时前者仅文案通用化，后者会导致错误静默丢失（用户点了没反应）。
  静默开关统一为 axios config 的 `silentError`，废弃 `safeApiCall` 自带的 `options.silent` 第二套开关；
  15 个不接受 config 的 api 函数补齐 config 参数，使「想静默」真正生效。
  接受语境文案（如「车辆查询失败」）退化为后端 message 兜底（已与用户确认）。
- **根因**: 提示职责分散在拦截器 / safeApiCall / 业务层 catch 三层，且静默有两套互不相通的开关。
  附带修掉：高频轮询（心跳、定位上报）断网时 Toast 刷屏；loadIncome 非静默时 5 个并发请求连弹 5 次再叠 Dialog。
- **验证**: `vite build` 通过（426 modules）；lint 0 错误；grep 全量复核 24 处 catch 块 +
  71 处 api 调用点，确认零遗漏双弹点、旧开关 `options.silent` 零残留。
- **新增约定（后续改动须遵守）**: 要么拦截器弹（默认），要么业务层传 `silentError: true` 自己弹，二者互斥。

## [本日早前，时间未记录] - 重构: 提现金额单位在 API 边界归一化为 amountCents

- **文件**: `web/driver/src/utils/driver-format.js`、`composables/useDriverAssets.js`、`views/mine/DriverWalletPage.vue`
- **决策**: 核对后端确认 `driver_withdraw.amount` 单位为「元」（DECIMAL(10,2) 全链路原样透传，
  proto 为 double），而收入为「分」。前端**不存在** 100 倍显示 bug（原 `round(amount*100)` 正确），
  但隐式转换藏在视图层私有函数 `withdrawAmountCents` 且无注释，极易被误删。
  改为在数据入口归一化出 `amountCents`，视图层与收入字段共用 `formatPrice`。
  后端单位契约未改（涉及 DB 迁移 + admin 端，另行排期）。
- **验证**: `vite build` 通过；node 运行时等价性实证 10 例全部 `legacyEqual=true`。

## [09:40] - 修复: 司机端 JWT 签名密钥改明文，clone 后可直接启动

- **文件**: `rpc/driversvc/etc/driversvc.yaml`、`api/driver/internal/svc/service_context.go`、`api/driver/internal/svc/service_context_test.go`
- **背景**: driversvc 启动校验要求 signingKey 非空且非默认弱值，yaml 原故意留空依赖环境变量注入，致他人 clone 后 `make run-driversvc` 直接 panic（"signing key is empty"）。driver-api 的 `resolveSigningKey()` 只读环境变量无回退，同样起不来；且 driver-api 有跨服务一致性校验（DRIVER_SIGNING_KEY 必须与 DRIVERSVC_SIGNING_KEY 一致）。
- **决策**: 两处共填同一固定开发明文 `driversvc-local-dev-key`（≠被校验拒绝的默认弱值 `local-development-signing-key`）；driver-api 的 `resolveSigningKey()` 增加回退该常量。保留环境变量覆盖（生产注入强密钥）。
- **验证**: `go build ./...` 通过；`go test ./internal/svc/...` 通过（更新 `TestResolveSigningKeyFallsBackToLocalKeyWhenEnvMissing`）。
- **风险**: 明文密钥进入仓库；生产须用环境变量覆盖。usersvc 仍用 `local-development-signing-key`（另一套），本次未动。

## [09:40] - 核查: "订单分类被删除" 不成立

- **结论**: 本次会话（含此前资质改造轮次）未删除任何订单分类功能。
- **证据**: `DriverOrdersPanel.vue:56-61` 当前仍有状态分类筛选下拉（orderMode + orderStatus 两个 van-dropdown-item）；git diff 该文件仅 `+8/-1`（新增轨迹按钮），零分类删除；`DriverHome.vue:384-392` 的 `orderStatusOptions` 7 项完整；后端 order_logic.go(+34)、types.go(+29/-6) 均为资质改造 + 新增轨迹类型。
- **提示**: 前端"按车型/服务类型（快车/专车/拼车）分类"在 model 与组件内均未找到，可能从未实现。待用户指认具体所指。

## [10:20] - 修复: 前端审查问题 5–9（7/8/9 实修，5/6 已处理）

- **问题5 (DriverHome 内联车辆/资质死代码)**: 搜索 `DriverHome.vue` 已无任何 `loadVehicle/submitVehicle/submitVehicleUpdate/removeVehicle/loadCertification/submitCertification` 残留，前序轮次已清理，无需改。
- **问题6 (我的页工具按钮无反应)**: `DriverMinePanel.vue:60-66` 的 `openTool` 已对无 `action` 按钮 `showToast('XX暂未开放')`（发票中心/银行卡/邀请有奖），无需改。
- **问题7 (订单记录页加载失败静默)**: `DriverOrderRecordsPage.vue` 导入 `showToast`；`loadRecords` 两调用加 `{ silentError: true }` 由业务层统一提示一次，catch 内 `showToast(error?.message || '订单记录加载失败')`，消除静默且不与拦截器重复弹窗。
- **问题8 (成功响应 data:null 返回 null 致崩溃)**: `api/request.js` 拦截器成功分支改为 `payload ?? {}`，data 为 null 时降级空对象，阻止 null 传播至 `persistSession`/`refreshProfile`（`stores/driver.js:35/64`）的 `null.driver`/`null.token` 崩溃；一处改动覆盖登录密码/短信/刷新资料三处调用。
- **问题9 (抢单池分支不可达)**: `DriverHome.vue` 移除 `loadOrders` 中 `orderMode==='available'` 永不可达分支（orderModeOptions 仅 orders/dispatches），并删除因此而成的死代码 `loadAvailableOrders` 函数（原 1364-1370）。
- **验证**: `read_lints` 三文件 0 错误；`npx vite build` 通过。

## [10:30] - 自检: 5–9 改动交付前回归检查

- **结果**: 通过，无遗留错误。
- `read_lints` 三文件 0 错误；`search_content` 确认 `loadAvailableOrders` 在 `DriverHome.vue` 内 0 残留引用（函数已删净）；`request.js` 拦截器 `payload ?? {}` 改动落地正确。
- 逻辑回归: 问题8 的 null→{} 仅作用于后端异常返回 data:null 的边界场景，persistSession/refreshProfile 经 {} 安全；正常接口（data 为对象/数组/undefined）行为不变。`npx vite build` 为全部改动完成后所跑，已通过。

## [10:15] - 功能实现: 接通司机端订单轨迹入口和查询接口

- **文件**: `api/driver/driver.api`, `api/driver/internal/types/types.go`, `api/driver/internal/logic/order_logic.go`, `api/driver/internal/logic/order_logic_test.go`, `api/driver/internal/handler/order_handler.go`, `api/driver/main.go`, `api/driver/main_test.go`, `web/driver/src/api/driver.js`, `web/driver/src/components/driver-home/DriverOrdersPanel.vue`, `web/driver/src/components/driver-home/DriverTrajectoryPanel.vue`, `web/driver/src/views/DriverHome.vue`, `web/driver/scripts/check-driver-web.mjs`
- **决策**: 在订单列表和首页行程中入口打开轨迹底部弹层；后端只按当前司机 JWT + orderId 查询 `TrajectoryRepository.ListByOrder`。
- **验证**: `go test ./api/driver/... -count=1` 通过；`npm run test:driver-web` 通过；`npm run build` 通过；`git diff --check` 退出码 0。

## [10:24] - 代码审查修复: 防止订单轨迹刷新请求乱序覆盖当前订单

- **文件**: `web/driver/src/views/DriverHome.vue`, `web/driver/src/components/driver-home/DriverTrajectoryPanel.vue`, `web/driver/scripts/check-driver-web.mjs`
- **决策**: 采用请求序号 + 当前 orderId 复核，允许新订单轨迹请求自然覆盖旧请求；轨迹列表 key 增加坐标和 index，避免同秒轨迹点重复 key。
- **验证**: `npm run test:driver-web` 通过；`npm run build` 通过。

## [10:31] - 重构: 删除 DriverHome 车辆/资质死状态并固化检查

- **文件**: `web/driver/src/views/DriverHome.vue`, `web/driver/scripts/check-driver-web.mjs`
- **决策**: DriverHome 不再维护车辆/资质表单同步副本；车辆、资质、钱包操作统一保留在独立页面和 `useDriverAssets`，首页检查脚本禁止旧 mine 事件和内联资产函数回流。
- **验证**: `rg` 扫描 DriverHome 无旧函数/旧事件/旧状态命中；`npm run test:driver-web` 通过；`npm run build` 通过；`git diff --check` 退出码 0。

## [10:46] - Bug修复: 修复司机端“我的”订单概览统计并固化提现金额单位约束

- **文件**: `web/driver/src/views/DriverHome.vue`, `web/driver/scripts/check-driver-web.mjs`
- **决策**: “我的”页订单概览用现有 `listDriverOrders` 分状态查询 `total` 后透传 `orderStats`；提现金额后端契约确认为“元”，前端继续在 `useDriverAssets` 数据入口归一化为 `amountCents` 再展示。
- **验证**: `npm run test:driver-web` 通过；`npm run build` 通过；`go test ./api/driver/internal/logic -run "TestCreateWithdraw|TestGetIncomeSummary" -count=1` 通过；`go test ./rpc/driversvc/internal/logic -run "TestCreateWithdraw|TestListWithdraws" -count=1` 通过；`git diff --check` 退出码 0。

## [11:00] - Bug修复: 修复司机首页行程态缺少开始/结束主流程入口

- **文件**: `web/driver/src/views/DriverHome.vue`, `web/driver/scripts/check-driver-web.mjs`
- **决策**: 首页底部主行程面板进入 `driving` 模板后必须按订单状态暴露“开始行程/结束行程”；结束提交统一复用 `resolveOrderId`，兼容 `orderId/orderID/id` 等订单形态。
- **验证**: 先新增回归断言并确认 `npm run test:driver-web` Red 失败；修复后 `npm run test:driver-web` 通过；`npm run build` 通过；`go test ./api/driver/internal/logic -run "TestConfirmArrive|TestStartTrip|TestFinishTrip" -count=1` 通过；`go test ./rpc/ordersvc/internal/logic -run "TestConfirmArrive|TestStartTrip|TestFinishTrip" -count=1` 因既有 `fakePayClient` 缺少 `GetSettlement` 编译失败，未纳入本次前端修复；`git diff --check` 退出码 0。
