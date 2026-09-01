# Driver Heatmap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a driver-side heatmap endpoint that reads nearby wait-accept orders from `ride_order`, aggregates them into grid cells, and caches short-lived results in Redis.

**Architecture:** The HTTP handler will authenticate as a driver, validate position + radius, and delegate to a small logic helper in `api/driver/internal/logic`. That helper will query the shared MySQL connection already present in `ServiceContext`, filter `ride_order` rows with `status = 1`, aggregate by grid cell, and store/retrieve a JSON response from Redis for a few seconds.

**Tech Stack:** Go, `gorm.io/gorm`, `github.com/redis/go-redis/v9`, existing `api/driver` HTTP stack, `miniredis` for tests.

## Global Constraints

- Keep all user-facing changes inside the driver API surface.
- Reuse the existing `api/driver` MySQL configuration and Redis client wiring.
- Do not add cross-service RPCs for this endpoint.
- Cache TTL stays short, on the order of seconds.
- The query targets `ride_order` rows where `status = 1`.

---

### Task 1: Add request/response types and route

**Files:**
- Modify: `api/driver/driver.api`
- Modify: `api/driver/internal/types/types.go`
- Modify: `api/driver/main.go`
- Modify: `api/driver/internal/handler/driver_handler.go`
- Test: `api/driver/main_test.go`

**Interfaces:**
- Consumes: `longitude`, `latitude`, `radiusMeters`
- Produces: `HeatmapRequest`, `HeatmapResponse`, `HeatmapPoint`

- [ ] **Step 1: Write the failing test**

```go
func TestDriverHeatmapRouteIsRegistered(t *testing.T) {
    handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "heatmap-route-test-key"})
    req := httptest.NewRequest(http.MethodPost, "/api/driver/v1/orders/heatmap", bytes.NewBufferString(`{"longitude":116.397,"latitude":39.908,"radiusMeters":2000}`))
    resp := httptest.NewRecorder()
    handler.ServeHTTP(resp, req)
    if resp.Code == http.StatusNotFound {
        t.Fatal("heatmap route is missing")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./api/driver -run TestDriverHeatmapRouteIsRegistered -v`
Expected: FAIL because the route and types do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Add the route, request/response structs, and a handler stub that forwards into logic.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./api/driver -run TestDriverHeatmapRouteIsRegistered -v`
Expected: PASS.

### Task 2: Implement heatmap query, grid aggregation, and Redis cache

**Files:**
- Add: `api/driver/internal/logic/heatmap_logic.go`
- Modify: `api/driver/internal/svc/storage.go`
- Modify: `api/driver/internal/svc/service_context.go`
- Test: `api/driver/internal/logic/heatmap_logic_test.go`

**Interfaces:**
- Consumes: `HeatmapRequest`
- Produces: `HeatmapResponse` with `Points []HeatmapPoint`

- [ ] **Step 1: Write the failing test**

```go
func TestGetOrderHeatmapAggregatesAndCaches(t *testing.T) {
    // seed miniredis and a temp MySQL database with ride_order rows
    // call the new logic twice
    // first call should hit DB and return grouped points
    // second call should read the cached JSON and avoid a second DB query
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./api/driver/internal/logic -run TestGetOrderHeatmapAggregatesAndCaches -v`
Expected: FAIL because the logic is missing.

- [ ] **Step 3: Write minimal implementation**

Query `ride_order` with `status = 1`, filter by radius, bucket points into grid cells, and cache the serialized response in Redis for a few seconds.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./api/driver/internal/logic -run TestGetOrderHeatmapAggregatesAndCaches -v`
Expected: PASS.

### Task 3: Wire handler validation and full driver API check

**Files:**
- Modify: `api/driver/internal/handler/driver_handler.go`
- Modify: `api/driver/main_test.go`

**Interfaces:**
- Consumes: authenticated driver JWT
- Produces: JSON heatmap response

- [ ] **Step 1: Write the failing integration test**

```go
func TestDriverHeatmapEndpointReturnsPoints(t *testing.T) {
    // use httptest against newHTTPHandler with MySQL + Redis test setup
    // assert status 200 and a non-empty points list for seeded data
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./api/driver -run TestDriverHeatmapEndpointReturnsPoints -v`
Expected: FAIL until the handler is wired.

- [ ] **Step 3: Write minimal implementation**

Connect the handler to the logic and ensure invalid inputs return 400.

- [ ] **Step 4: Run the full driver API test suite**

Run: `go test ./api/driver ./api/driver/internal/logic`
Expected: PASS.
