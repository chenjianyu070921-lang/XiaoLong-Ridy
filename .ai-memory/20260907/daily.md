## [15:30] - [重构]: driver 路由登记对齐仓库主流 mux 约定

- **文件**: `api/driver/internal/handler/routes.go`(删) → `router.go`(新)；`api/driver/main.go`；`api/driver/main_test.go`
- **决策**: 调研 `admin`/`passenger`/`driver` 三服务 `main.go`，确认仓库约定为标准库 `&http.Server{Handler: NewRouter(svcCtx)}` + 手写 mux 路由文件，**无人使用 go-zero `rest.Server` / goctl `RegisterHandlers(server *rest.Server,...)`**。故选项 B（`rest.Server` 迁移，让 goctl 拥有 routes.go）被否决——会与全仓背离且非本仓库实际做法。选项 A（轻量 mux）即仓库真实约定，driver 已符合，仅做命名/形状对齐。
- **改动**: `RegisterHandlers(mux *http.ServeMux, svcCtx)` → `NewRouter(svcCtx *svc.ServiceContext) http.Handler`（mux 创建、静态资质服务 `/certification-files/`、全部路由登记收进 `NewRouter`）；`main.go` 启动改为 `middleware.InternalServiceAuth(svcCtx)(handler.NewRouter(svcCtx))` 由 `recoverMiddleware(withCORS(...))` 包裹；`main_test.go` 以同义 `newHTTPHandler` 辅助保留 21 处调用点行为等价。
- **验证**: `go vet ./api/driver/...` 无输出；`go build ./...` EXIT=0；`go test ./api/driver/...` 全部 `ok`（handler/logic/middleware/svc/types）。
- **约定沉淀**: 新增路由时，仿 `passenger`/`admin` 在 `handler/router.go` 的 `NewRouter` 里加一行 `mux.Handle(...)`（镜像 goctl 生成的 `XxxHandler(svcCtx)` 名），**不要**对线上源码跑 `goctl api go`（goctl 每次重写 routes.go 为 rest.Server 版，会破坏本仓库 mux 约定）。

## [19:21] - [修复]: api/driver 的 gRPC 兜底默认值从远端改为 localhost，消除“目录不对就偷连远端”的坑

- **文件**: `api/driver/main.go`（6 个 `defaultXxxGRPCAddr` 常量）
- **根因**: `api/driver/etc/driver.yaml` 早已把 grpc 地址写成 `127.0.0.1`，rpc 服务间也用 `127.0.0.1` 互调（见 ordersvc/paysvc yaml）；但 `main.go` 的兜底默认值仍写死 `115.191.16.159`。从错误目录跑 `go run ./api/driver`（找不到 `etc/driver.yaml`）会 silent 回落到远端兜底值，导致连不上本地 driversvc。用户“连不上、别人能连上”即此因。
- **改动**: 6 个常量 `115.191.16.159:5005x` → `127.0.0.1:5005x`，与 yaml/rpc 约定对齐；补注释说明可用 env（`DRIVER_GRPC_ADDR` 等）或 yaml 覆盖为远端。
- **验证**: `go vet`/`go build ./api/driver/...` EXIT=0；本地全链路冒烟：`driversvc` 监听 `0.0.0.0:50055` + `driver-api` 加载 `etc/driver.yaml` 启动日志确认 gRPC 目标为 `127.0.0.1:50055/50051/50054/50053/50056/50057`，二者本地互通。
- **结论**: rpc 本就能本地启动；仓库设计是 rpc 全本地跑、共享远端 MySQL/Redis/Kafka/MinIO。本地联调用 `make run-driversvc` + `make run-driver-api` 即通（或任意目录跑时确保能找到 `etc/driver.yaml`）。
