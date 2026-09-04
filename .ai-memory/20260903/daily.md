# 工作日志 - 2026-09-03
session-id: 20260903-0935

## [09:35] - Bug fix: 拒绝司机 JWT 默认签名密钥并改为运行时注入

- **文件**: api/driver/internal/svc/service_context.go, api/driver/internal/svc/service_context_test.go, rpc/driversvc/driversvc.go, rpc/driversvc/etc/driversvc.yaml, rpc/driversvc/internal/config/config.go, rpc/driversvc/internal/config/config_test.go
- **决策**: driver API 在 DRIVER_SIGNING_KEY 缺失时返回空值并由启动校验失败；driver RPC 使用 DRIVERSVC_SIGNING_KEY 覆盖 yaml 后校验，拒绝空值和公开开发默认值；driversvc.yaml 只保留空占位。
- **验证**: go test ./api/driver/internal/svc, go test ./rpc/driversvc/internal/config, go test ./api/driver/internal/... ./rpc/driversvc/... 均通过；go test ./api/driver/... ./rpc/driversvc/... 因既有未注册路由断言失败。

## [09:51] - Bug fix: 司机端限流默认忽略可伪造 X-Forwarded-For

- **文件**: api/driver/internal/middleware/ratelimit.go, api/driver/internal/middleware/ratelimit_test.go
- **决策**: DRIVER_TRUSTED_PROXY_HOPS 未配置时使用 RemoteAddr 作为司机端公开接口限流 key；仅在显式声明可信代理跳数时从 X-Forwarded-For 右侧选择对应客户端 IP，链路异常或 IP 非法则回退 RemoteAddr。
- **验证**: go test ./api/driver/internal/middleware 通过；go test ./api/driver/internal/... 通过；git diff --check 通过；rg 确认司机端限流入口均走 clientKey。

## [10:05] - 安全加固: 司机端配置明文凭据改环境变量占位符 + 启动 fail-fast

- **文件**: api/driver/etc/driver.yaml, api/driver/main.go, api/driver/main_test.go, rpc/driversvc/etc/driversvc.yaml, rpc/driversvc/driversvc.go, rpc/driversvc/internal/config/config.go, rpc/driversvc/internal/config/config_test.go
- **决策**: 两个司机端 yaml 只保留占位符；占位符展开与缺失校验分别在 api/driver(main) 与 rpc/driversvc(internal/config) 各实现一份 ExpandEnv（用户限定"只能修改司机端"，故不下沉到 common/config）；driversvc 额外拒绝 MinIO 出厂账号 minioadmin。
- **约束**: 仅限司机端，未触碰其余 20+ 个仍含明文的服务配置；已入库凭据判定泄露，需全量轮换（同一密码在 MySQL/Redis 横向复用）。
- **验证**: go build + go vet 通过；新增 6 个用例全过；无 env 启动 driversvc/driver-api 均 panic 并列出缺失变量名；注入合规凭据后越过配置校验进入 DB 连接阶段；注入 minioadmin 被拒。
- **遗留**: api/driver 的 4 个路由断言失败为改动前既有问题（avatar/upload、reviews/list 未注册）。
- **已回退**: 用户选择方案 B（小组开发，其他服务仍为明文硬编码，要求司机端保持一致以便组员直连测试）。api/driver/etc/driver.yaml、api/driver/main.go、api/driver/main_test.go 已完全恢复基线；driversvc 三处文件仅保留 09:35 的 JWT 签名密钥改动。ExpandEnv、ValidateMinioCredentials、minioadmin 守卫与相关用例全部移除。

## [10:11] - Bug fix: 清理司机端限流器过期桶防止 map 无界保留

- **文件**: api/driver/internal/middleware/ratelimit.go, api/driver/internal/middleware/ratelimit_test.go
- **决策**: 在 loginLimiter.allow 持锁后按当前时间机会性删除已过期桶，再处理当前 key；保持内存限流器现有固定窗口语义，不引入新依赖，不改非司机端模块。
- **验证**: 新增 TestLimiterRemovesExpiredBuckets，红灯为 len(buckets)=3；修复后 go test ./api/driver/internal/middleware -run TestLimiterRemovesExpiredBuckets -count=1 通过，go test ./api/driver/internal/middleware -count=1 通过，go test ./api/driver/internal/... -count=1 通过，git diff --check 通过。

## [10:32] - Bug fix: 移除司机端未实现资产 tab 挂载入口

- **文件**: web/driver/src/views/DriverHome.vue
- **决策**: 按用户要求删除未计划实现的资产 tab 路径，不补充 DriverMinePanel 的缺失 props；当前 tabbar 仅保留首页、订单、我的，DriverMinePanel 只在我的 tab 挂载并传入 required 的 formatDriverStatus。
- **验证**: npm run build 通过；git diff --check -- web/driver/src/views/DriverHome.vue web/driver/src/components/driver-home/DriverMinePanel.vue 通过；rg 确认 DriverAssetsPanel、资产 tab、activeTab=3 残留入口不存在，formatDriverStatus 仍传入 DriverMinePanel。
## [10:40] - Bug fix: expose driver income summary withdrawable cents

- **Files**: api/driver/driver.api, api/driver/internal/types/types.go, api/driver/internal/logic/income_logic.go, api/driver/internal/logic/income_logic_test.go
- **Decision**: Add `withdrawableCents` to the driver income summary contract and populate it from the existing settled paysvc income aggregate so DriverMinePanel no longer renders a permanent zero from a missing field.
- **Verification**: Added a failing regression assertion first; `go test ./api/driver/internal/logic -run TestGetIncomeSummaryUsesPaysvcSettlements -count=1`, `go test ./api/driver/internal/logic -count=1`, `go test ./api/driver/internal/... -count=1`, `npm run build` in web/driver, and `git diff --check -- api/driver/driver.api api/driver/internal/types/types.go api/driver/internal/logic/income_logic.go api/driver/internal/logic/income_logic_test.go` passed. `go test ./api/driver/... -count=1` still fails on pre-existing 404 route assertions for avatar upload and reviews/list.
## [10:47] - Feature fix: show driver withdraw records on wallet page

- **Files**: web/driver/src/api/driver.js, web/driver/src/composables/useDriverAssets.js, web/driver/src/views/mine/DriverWalletPage.vue, web/driver/src/styles/driver-mine-pages.css, web/driver/scripts/check-driver-web.mjs
- **Decision**: Reuse the existing driver `/withdraws/list` backend endpoint from the wallet page instead of adding a new route or changing backend contracts.
- **Verification**: Added RED assertions to `npm run test:driver-web` and observed failure on missing `/withdraws/list`; after implementation `npm run test:driver-web`, `npm run build`, and `git diff --check -- web/driver/scripts/check-driver-web.mjs web/driver/src/api/driver.js web/driver/src/composables/useDriverAssets.js web/driver/src/views/mine/DriverWalletPage.vue web/driver/src/styles/driver-mine-pages.css` passed.
## [10:51] - Bug fix: stop driver help center from opening orders

- **Files**: web/driver/src/components/driver-home/DriverMinePanel.vue, web/driver/src/views/DriverHome.vue, web/driver/scripts/check-driver-web.mjs
- **Decision**: Give the mine-page help center its own `open-help` event and handle it with an unavailable toast instead of reusing or depending on order navigation.
- **Verification**: Added RED assertions to `npm run test:driver-web` and observed failure on missing dedicated help event; after implementation `npm run test:driver-web`, `npm run build`, and `git diff --check -- web/driver/src/components/driver-home/DriverMinePanel.vue web/driver/src/views/DriverHome.vue web/driver/scripts/check-driver-web.mjs` passed.
## [10:58] - Bug fix: remove dead driver web API wrappers

- **Files**: web/driver/src/api/driver.js, web/driver/src/components/driver-home/DriverMinePanel.vue, web/driver/src/views/DriverHome.vue, web/driver/scripts/check-driver-web.mjs
- **Decision**: Remove the unused `listNearbyDrivers`, `getOrderTrajectory`, and `listPassengerReviews` frontend wrappers; keep passenger reviews as a semantic unavailable UI event instead of calling the missing `/reviews/list` route, and remove stale trajectory wiring from the active driver home path.
- **Verification**: Added RED assertions to `npm run test:driver-web` and observed failure on `listNearbyDrivers`; after implementation `npm run test:driver-web`, `npm run build`, `git diff --check -- web/driver/src/api/driver.js web/driver/src/components/driver-home/DriverMinePanel.vue web/driver/src/views/DriverHome.vue web/driver/scripts/check-driver-web.mjs`, and targeted `rg` dead-route scan passed.
## [11:17] - Bug fix: correct driver order status labels

- **Files**: web/driver/src/views/DriverHome.vue, web/driver/scripts/check-driver-web.mjs
- **Decision**: Align the active driver home order status formatter and filter labels with the ordersvc enum: 4 is 待支付, 5 is 已完成, and 6 is 已取消.
- **Verification**: Added RED assertions to `npm run test:driver-web` and observed failure on status 4; after implementation `npm run test:driver-web`, `npm run build`, `git diff --check -- web/driver/src/views/DriverHome.vue web/driver/scripts/check-driver-web.mjs`, and targeted status-label `rg` scan passed.
## [11:23] - Bug fix: replace unavailable driver withdrawable income display

- **Files**: web/driver/src/components/driver-home/DriverMinePanel.vue, web/driver/scripts/check-driver-web.mjs
- **Decision**: Stop the driver mine wallet card from reading the unavailable `incomeSummary.withdrawableCents` field; render backend-owned `incomeSummary.completedOrders` as the secondary business metric instead.
- **Verification**: Added RED assertions to `npm run test:driver-web` and observed failure on `withdrawableCents`; after implementation `npm run test:driver-web` and `npm run build` in `web/driver` passed.
## [11:33] - Bug fix: remove unused driver HTTP routes

- **Files**: api/driver/driver.api, api/driver/main.go, api/driver/main_test.go, api/driver/etc/driver.yaml, api/driver/internal/handler/driver_handler.go, api/driver/internal/handler/order_handler.go, api/driver/internal/handler/trajectory_handler.go, api/driver/internal/logic/driver_logic.go, api/driver/internal/logic/order_logic.go, api/driver/internal/logic/trajectory_logic.go, api/driver/internal/logic/validate.go, api/driver/internal/logic/driver_client_test.go, api/driver/internal/logic/order_logic_test.go, api/driver/internal/svc/service_context.go, api/driver/internal/types/types.go
- **Decision**: Remove the zero-frontend-call driver HTTP routes `/orders/trajectory`, `/drivers/nearby`, and `/orders/cancel` from api/driver's public contract and mux; keep lower-level RPC capabilities and trajectory storage intact for non-HTTP/internal use.
- **Verification**: Added RED test `TestUnusedDriverHTTPRoutesAreNotExposed` and observed 401/driver.api declaration failures; after implementation `go test ./api/driver -count=1`, `go test ./api/driver/internal/... -count=1`, `go test ./api/driver/... -count=1`, `npm run test:driver-web`, targeted route scans, and `git diff --check -- api/driver web/driver/scripts/check-driver-web.mjs .ai-memory/20260903/daily.md` passed.

## [11:40] - Refactor: reuse shared driver order status formatter

- **Files**: web/driver/src/views/DriverHome.vue, web/driver/scripts/check-driver-web.mjs
- **Decision**: Remove the local duplicate `formatOrderStatus` from DriverHome and import the shared formatter from `@/utils/driver-format` so driver order status labels stay single-sourced.
- **Verification**: Added RED assertions to `npm run test:driver-web` and observed failure on missing shared import; after implementation `npm run test:driver-web`, `npm run build`, targeted formatter scan, and `git diff --check -- web/driver/src/views/DriverHome.vue web/driver/scripts/check-driver-web.mjs` passed.

## [12:03] - Bug fix: add configurable driver API CORS

- **Files**: api/driver/main.go, api/driver/main_test.go, api/driver/etc/driver.yaml
- **Decision**: Add an explicit driver H5 origin whitelist via `corsAllowedOrigins` or `DRIVER_CORS_ALLOWED_ORIGINS`; only configured origins receive CORS headers, and preflight `OPTIONS` is handled before auth/rate-limit middleware.
- **Verification**: Added RED CORS tests and observed build failure for missing config/function; after implementation `go test ./api/driver -run "TestDriverCORS|TestLoadDriverConfigReadsYamlAndEnvCanOverride" -count=1`, `go test ./api/driver -count=1`, `go test ./api/driver/... -count=1`, `npm run test:driver-web`, and `git diff --check -- api/driver/main.go api/driver/main_test.go api/driver/etc/driver.yaml` passed.

## [14:22] - Bug fix: make driver mine back buttons return previous page

- **Files**: web/driver/src/views/mine/DriverWalletPage.vue, web/driver/src/views/mine/DriverIncomePage.vue, web/driver/src/views/mine/DriverOrderRecordsPage.vue, web/driver/src/views/mine/DriverCertificationPage.vue, web/driver/src/views/mine/DriverVehiclePage.vue, web/driver/src/views/DriverProfileEdit.vue
- **Decision**: Remove the homepage fallback from driver mine/profile back buttons and use browser history back directly.
- **Verification**: `npm run test:driver-web` passed; targeted scan confirmed the mine/profile back handlers no longer call `router.replace('/home')` or gate on `window.history.length` in these views.

## [14:44] - Bug fix: separate driver service score from certification upload

- **Files**: web/driver/src/components/driver-home/DriverMinePanel.vue, web/driver/src/views/DriverHome.vue, web/driver/scripts/check-driver-web.mjs
- **Decision**: Show service score as a read-only dialog backed by the loaded dashboard score, and guard AMap heatmap updates so empty datasets hide the layer instead of calling SDK `setDataSet`.
- **Verification**: `npm run test:driver-web` and `npm run build` passed; targeted scans found no project source defining `reportAllChanges` and confirmed service score no longer directly navigates to certification upload.

## [14:58] - Bug fix: preserve driver home tab when returning from mine pages

- **Files**: web/driver/src/views/DriverHome.vue, web/driver/src/components/driver-home/DriverMinePanel.vue, web/driver/scripts/check-driver-web.mjs
- **Decision**: Store the driver home tab in the `/home?tab=` query and replace the previous history entry with `/home?tab=mine` before opening standalone mine pages, so their back buttons still use browser history but return to the mine tab.
- **Verification**: `npm run test:driver-web`, `npm run build`, targeted route-state scan, and `git diff --check -- web/driver/src/views/DriverHome.vue web/driver/src/components/driver-home/DriverMinePanel.vue web/driver/scripts/check-driver-web.mjs .ai-memory/20260903/daily.md` passed.

## [15:35] - Feature implementation: add driver grab-order list endpoint alias

- **Files**: api/driver/driver.api, api/driver/main.go, api/driver/main_test.go, web/driver/src/api/driver.js, web/driver/scripts/check-driver-web.mjs
- **Decision**: Keep the existing `/orders/available` implementation as the source of truth and add `/orders/grab-list` as the driver-facing business-named grab-order list endpoint, with the frontend wrapper delegating compatibility calls to the new path.
- **Verification**: `go test ./api/driver -count=1`, `go test ./api/driver/internal/... -count=1`, `npm run test:driver-web`, `npm run build`, targeted route scan, and `git diff --check -- api/driver/driver.api api/driver/main.go api/driver/main_test.go web/driver/src/api/driver.js web/driver/scripts/check-driver-web.mjs .ai-memory/20260903/daily.md` passed.

## [15:30] - 安全清理: 删除司机端无鉴权死代码链路

- **文件**: api/driver/internal/handler/driver_handler.go, api/driver/internal/logic/driver_logic.go, api/driver/internal/logic/validate.go, api/driver/internal/types/types.go, api/driver/driver.api
- **决策**: 删除整条从未在 main.go 注册的死代码链——3 个 handler（CreateDriverHandler、GetDriverByPhoneHandler、DeleteDriverHandler；其中 GetDriverByPhoneHandler 仅做登录态校验而无归属校验、DeleteDriverHandler 完全无鉴权）、对应的 3 个 logic 方法（CreateDriver、GetDriverByPhone、DeleteDriver）、validate.go 的 normalizeCreateDriverRequest（仅被 CreateDriver 使用），以及 types.go 与 driver.api 中孤立的 6 个类型（CreateDriverRequest/Response、GetDriverByPhoneRequest/Response、DeleteDriverRequest、DeleteResponse）。保留 RPC 客户端层（service_context.go 的 DriverClient），并移除 driver_logic.go 仅被 GetDriverByPhone 使用的 `strings` import。
- **附带修复**: driver_logic.go 中 maskIDCard 的注释原本错位在 GetDriverByPhone 之前，删除后者后该注释自然归位到 maskIDCard 函数前。
- **验证**: go build ./api/driver/... 通过（BUILD_EXIT_0）；go test ./api/driver/... 全过；gofmt -l 对改动文件无输出（格式干净）。

## [16:30] - 完善双向评价系统（重点：司机接收乘客端评价）

- **范围（用户确认）**: 完整双向——司机查看乘客对自己的评价 + 司机评价乘客；新增 `driver_review` 表，前端双向页面，让 `api/driver/internal/logic/review_logic_test.go` 蓝图通过。
- **后端 api/driver**:
  - 新增 `types/review.go`、`svc/review_repository.go`（PassengerReview 映射 order_review 只读 / DriverOrderReview 映射 driver_review / ReviewRepository 接口 + GORM 实现 + ErrReviewAlreadyExists）、`logic/review_logic.go`（ListReceivedReviews / SubmitDriverReview / ListGivenReviews + ErrReviewRepositoryNotConfigured + ErrReviewAlreadyExists）、`handler/review_handler.go`（3 个 handler）。
  - `svc/service_context.go`：ServiceContext 增 `ReviewRepository` 字段；MySQL 初始化分支注入 `NewGormDriverReviewRepository(db)` 并 `AutoMigrate(&DriverOrderReview{})`（无 DSN 时保持 nil，逻辑层返回 50001）。
  - `main.go` 注册 `GET /reviews/received`、`POST /reviews/submit`、`POST /reviews/given`（均走 protected 鉴权）。`driver.api` 补类型与路由。`handler/error.go` 为 `ErrReviewRepositoryNotConfigured` 增加 50001 映射。
  - **隐私保护（对齐蓝图 test 期望 `UserID==0`）**: 评价列表不暴露对端用户身份 ID。`PassengerReview.CreatedAt` 用 int64（UNIX 秒，查询以 `UNIX_TIMESTAMP(created_at)` 投影）以匹配蓝图 `CreatedAt: 123`。
- **前端 web/driver**:
  - `src/api/driver.js` 新增 `listReceivedReviews` / `listGivenReviews` / `submitDriverReview`。
  - 新增 `src/components/driver-home/DriverReviewsPanel.vue`（Vant 弹层：收到的评价列表 + 我给出的评价 + 评价乘客表单），`DriverHome.vue` 的 `openPassengerReviews` 由弹 Toast 占位改为打开该面板（`reviewsPanelVisible`/`reviewsPanelMode`）。
  - `scripts/check-driver-web.mjs` 断言更新：由"固化暂未开放"改为"断言打开真实面板"。
- **迁移**: 新增 `scripts/sql/migrate/10_driver_review.sql`（建 driver_review 表，order_id 唯一键防重复评价，与 09_passenger_review.sql 范式一致）。
- **验证**: go build ./api/driver/... 通过；go test ./api/driver/... 全过（含蓝图 review_logic_test.go）；gofmt 干净；`npm run build`（web/driver）通过；`node scripts/check-driver-web.mjs` 通过（CHECK_EXIT_0）。

## [17:11] - Feature implementation: add driver in-trip realtime fare query

- **Files**: api/driver/driver.api, api/driver/etc/driver.yaml, api/driver/internal/handler/error.go, api/driver/internal/handler/order_handler.go, api/driver/internal/logic/order_logic.go, api/driver/internal/logic/order_logic_test.go, api/driver/internal/logic/validate.go, api/driver/internal/svc/service_context.go, api/driver/internal/types/types.go, api/driver/main.go, api/driver/main_test.go, web/driver/scripts/check-driver-web.mjs, web/driver/src/api/driver.js, web/driver/src/views/DriverHome.vue
- **Decision**: Add `POST /api/driver/v1/orders/realtime-fare` as a driver-owned, on-trip-only read query that calls pricesvc `EstimatePrice`; the H5 driving panel polls it every 15 seconds and displays a compact realtime fare strip with estimated-fare fallback.
- **Verification**: `go test ./api/driver/internal/logic -run "TestGetRealtimeFare|TestOrderLogicRejectsInvalidOrderParameters" -count=1`, `go test ./api/driver -run "TestRealtimeFareRouteReturnsPriceDetail|TestLoadConfig|TestDefaultBackend" -count=1`, `go test ./api/driver/internal/... -count=1`, `go test ./api/driver -count=1`, `npm run test:driver-web`, `npm run build`, `Invoke-WebRequest http://localhost:5176/`, and `git diff --check -- <touched realtime fare files>` passed.
