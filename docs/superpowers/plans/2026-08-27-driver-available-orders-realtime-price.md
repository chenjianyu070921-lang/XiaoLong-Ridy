# Driver Available Orders And Realtime Price Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a temporary location-based driver dispatch hall and a durable realtime trip price query.

**Architecture:** `api/driver` owns the temporary dispatch hall aggregation because it can combine JWT driver identity, Redis position snapshots, and existing ordersvc reads. Realtime pricing is added to ordersvc so order ownership and status checks stay with the order state machine before delegating to pricesvc.

**Tech Stack:** Go, net/http, go-zero style RPC code, protobuf/gRPC, GORM/in-memory repositories, go-redis, miniredis, pricesvc EstimatePrice.

## Global Constraints

- Driver identity comes from JWT; clients must not pass `driverId`.
- Available orders depend on `driver:pos:<driver_id>` and `driver:online`.
- The available orders implementation is temporary technical debt for training/demo data volume.
- Realtime price must route through ordersvc and reuse pricesvc EstimatePrice.
- Realtime price must not persist pricing, finish the trip, or create payment.

---

### Task 1: Temporary Available Orders Logic

**Files:**
- Modify: `api/driver/internal/logic/order_logic_test.go`
- Modify: `api/driver/internal/logic/order_logic.go`
- Modify: `api/driver/internal/types/types.go`
- Modify: `api/driver/driver.api`

**Interfaces:**
- Consumes: `constants.RedisDriverPos`, `constants.RedisDriverOnline`, `ordersvc.ListOrders`, `ordersvc.GetOrder`.
- Produces: `ListAvailableOrders(driverID int64, page, pageSize int32) (*types.ListMyOrdersResponse, error)` filters and sorts wait-accept orders by distance from the current driver's Redis position.

- [ ] Write failing tests for offline/no-position empty result and online distance-sorted result.
- [ ] Run `go test ./api/driver/internal/logic -run "TestListAvailableOrders" -count=1` and confirm the new tests fail.
- [ ] Implement Redis online/position read, in-memory distance filtering, sorting, and pagination.
- [ ] Re-run the focused tests and confirm they pass.

### Task 2: Realtime Price In Ordersvc

**Files:**
- Modify: `rpc/ordersvc/proto/ordersvc.proto`
- Modify: `rpc/ordersvc/proto/ordersvc.pb.go`
- Modify: `rpc/ordersvc/proto/ordersvc_grpc.pb.go`
- Modify: `rpc/ordersvc/order/order.go`
- Modify: `rpc/ordersvc/internal/server/order_server.go`
- Create: `rpc/ordersvc/internal/logic/realtime_price_logic.go`
- Create/modify: `rpc/ordersvc/internal/logic/realtime_price_logic_test.go`

**Interfaces:**
- Consumes: `OrderRepository.GetByID`, `PriceClient.EstimatePrice`.
- Produces: `RealtimePrice(RealtimePriceRequest) returns (RealtimePriceResponse)` with `price_rule_id`, `total_cents`, and fee detail fields.

- [ ] Write failing tests for success, driver mismatch, non-ON_TRIP status, and forwarded pricesvc request fields.
- [ ] Run `go test ./rpc/ordersvc/internal/logic -run "TestRealtimePrice" -count=1` and confirm the tests fail because the RPC/logic does not exist.
- [ ] Add proto messages and service method, regenerate or consistently update generated files.
- [ ] Implement `RealtimePriceLogic` using the same city fallback behavior as FinishTrip.
- [ ] Re-run focused tests and confirm they pass.

### Task 3: Driver API Realtime Price Endpoint

**Files:**
- Modify: `api/driver/driver.api`
- Modify: `api/driver/internal/types/types.go`
- Modify: `api/driver/internal/svc/service_context.go`
- Modify: `api/driver/internal/logic/order_logic.go`
- Modify: `api/driver/internal/logic/order_logic_test.go`
- Modify: `api/driver/internal/handler/order_handler.go`
- Modify: `api/driver/main.go`
- Modify: `web/user/src/api/driver.js`

**Interfaces:**
- Consumes: `ordersvc.RealtimePrice`.
- Produces: `POST /api/driver/v1/orders/realtime-price` with request `{ "orderId": 1001, "distanceM": 1200, "durationS": 600 }`.

- [ ] Write failing api/driver logic test proving the authenticated driver id is forwarded to ordersvc.
- [ ] Run `go test ./api/driver/internal/logic -run "TestRealtimePrice" -count=1` and confirm it fails.
- [ ] Add OrderClient method, gRPC adapter method, logic method, handler, route, and frontend API helper.
- [ ] Re-run focused api/driver tests and confirm they pass.

### Task 4: Docs And Verification

**Files:**
- Modify: `docs/driver/00-index.md`
- Create/modify: `docs/driver/28-order-available.md`
- Create: `docs/driver/29-order-realtime-price.md`

**Interfaces:**
- Produces: documented temporary debt for available orders and stable contract for realtime price.

- [ ] Document `POST /api/driver/v1/orders/available` as a temporary Redis-position-based dispatch hall.
- [ ] Document `POST /api/driver/v1/orders/realtime-price`.
- [ ] Run `gofmt` on touched Go files.
- [ ] Run `go test ./api/driver/... ./rpc/ordersvc/... -count=1`.
- [ ] Run `git diff --check`.
