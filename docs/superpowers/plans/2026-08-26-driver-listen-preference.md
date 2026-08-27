 # Driver Listen Preference Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add driver listen preferences for realtime and reservation orders, persist them, sync them when drivers go online, and filter dispatch candidates by order type.

**Architecture:** `driversvc` owns preference persistence and Redis synchronization. `api/driver` exposes driver-facing preference read/write and online request fields. `dispatchsvc` receives `order_type` and filters Redis GEO candidates by preference membership.

**Tech Stack:** Go, go-zero RPC/API style, gRPC/protobuf, GORM repositories, Redis sets/GEO, PowerShell proto generation scripts, MySQL migrations.

## Global Constraints

- Do not change bottom-layer/framework code.
- Keep scope to driver-side business plus dispatch candidate filtering required by driver preferences.
- Use explicit `order_type`; do not overload `car_type`.
- Default missing preference to accept both realtime and reservation orders.
- Validate preferences so at least one of realtime/reservation is accepted.
- Use `rpc/driversvc/scripts/regenerate_proto.ps1` for driversvc generation; do not use the old root-relative command.

---

### Task 1: Proto And Contract Surface

**Files:**
- Modify: `rpc/driversvc/proto/driversvc.proto`
- Modify: `rpc/dispatchsvc/proto/dispatchsvc.proto`
- Modify generated files under `rpc/driversvc` and `rpc/dispatchsvc/proto`

**Interfaces:**
- Produces: `SetDriverListenPreference`, `GetDriverListenPreference`, `DriverListenPreference`, optional online preference fields, and `DispatchOrderRequest.order_type`.

- [ ] Add failing compile-oriented tests that reference the new proto fields/methods.
- [ ] Run focused tests and confirm compile failure before generation.
- [ ] Edit proto definitions.
- [ ] Regenerate driversvc via `powershell -ExecutionPolicy Bypass -File rpc/driversvc/scripts/regenerate_proto.ps1`.
- [ ] Regenerate dispatchsvc proto from `rpc/dispatchsvc/proto` with `protoc --go_out=. --go-grpc_out=. dispatchsvc.proto`.
- [ ] Run focused compile tests.

### Task 2: driversvc Preference Persistence And Redis Sync

**Files:**
- Create: `rpc/driversvc/internal/model/driver_listen_preference.go`
- Create: `rpc/driversvc/internal/repository/driver_listen_preference_repository.go`
- Create: `rpc/driversvc/internal/repository/gorm_driver_listen_preference_repository.go`
- Modify: `rpc/driversvc/internal/svc/service_context.go`
- Modify: `rpc/driversvc/internal/logic/set_driver_online_logic.go`
- Modify: `rpc/driversvc/internal/logic/set_driver_offline_logic.go`
- Modify: `rpc/driversvc/internal/logic/dispatch_availability.go`
- Create: `rpc/driversvc/internal/logic/set_driver_listen_preference_logic.go`
- Create: `rpc/driversvc/internal/logic/get_driver_listen_preference_logic.go`

**Interfaces:**
- Consumes: proto types from Task 1.
- Produces: repository methods `GetByDriverID`, `Upsert`, default preference helper, and Redis keys `driver:pref:realtime`, `driver:pref:reservation`.

- [ ] Write failing driversvc logic tests for default both, invalid none, upsert, online sync, offline cleanup.
- [ ] Run tests and confirm failures.
- [ ] Implement model, repository, service context wiring, logic handlers, and Redis sync.
- [ ] Run driversvc focused tests.

### Task 3: Driver API Preference Endpoints

**Files:**
- Modify: `api/driver/driver.api`
- Modify: `api/driver/internal/types/types.go`
- Modify: `api/driver/internal/svc/service_context.go`
- Create or modify: `api/driver/internal/logic/listen_preference_logic.go`
- Create or modify: `api/driver/internal/handler/listen_preference_handler.go`
- Modify: `api/driver/internal/logic/online_logic.go`
- Modify: `api/driver/internal/handler/online_handler.go`
- Modify: `api/driver/main.go`

**Interfaces:**
- Consumes: driversvc preference RPC from Task 1.
- Produces: HTTP `POST /api/driver/v1/drivers/listen-preference`, `GET /api/driver/v1/drivers/listen-preference`, and optional online preference forwarding.

- [ ] Write failing API logic/route tests for preference update/get and online forwarding.
- [ ] Run tests and confirm failures.
- [ ] Implement types, logic, handlers, and route registration following existing driver API patterns.
- [ ] Run API focused tests.

### Task 4: dispatchsvc Preference Filtering

**Files:**
- Modify: `common/constants/constants.go`
- Modify: `rpc/dispatchsvc/internal/engine/geo_dispatch_engine.go`
- Modify: `rpc/dispatchsvc/internal/engine/mock_dispatch.go`
- Modify: `rpc/dispatchsvc/internal/logic/dispatch_order_logic.go`
- Modify: `rpc/dispatchsvc/internal/svc/service_context.go`

**Interfaces:**
- Consumes: `DispatchOrderRequest.order_type`.
- Produces: candidate filtering by `driver:pref:realtime` or `driver:pref:reservation`.

- [ ] Write failing dispatch engine tests proving realtime and reservation filtering.
- [ ] Run tests and confirm failures.
- [ ] Implement preference checker and order type propagation.
- [ ] Run dispatchsvc focused tests.

### Task 5: SQL, Docs, And Verification

**Files:**
- Modify: `scripts/sql/migrate/02_driver_module.sql`
- Create: `scripts/sql/migrate/12_driver_listen_preference.sql`
- Modify driver docs under `docs/driver/`

**Interfaces:**
- Produces: schema for `driver_listen_preference` and documented API behavior.

- [ ] Add SQL table to base migration and incremental migration.
- [ ] Add driver API docs.
- [ ] Run `gofmt` on touched Go files.
- [ ] Run focused tests.
- [ ] Run `go test ./api/driver/... ./rpc/driversvc/... ./rpc/dispatchsvc/... -count=1`.
- [ ] Run `git diff --check` and targeted `rg` searches.