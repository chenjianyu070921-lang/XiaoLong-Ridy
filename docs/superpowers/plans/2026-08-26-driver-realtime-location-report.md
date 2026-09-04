# Driver Realtime Location Report Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the existing driver realtime location report path with reusable driversvc logic that updates MySQL and Redis for dispatch, trip navigation, and future dispatch hall reads.

**Architecture:** `api/driver` keeps the existing HTTP endpoint and forwards validated requests to `driversvc.ReportLocation`. `driversvc` owns the reusable heartbeat, latest location persistence, and Redis sync helpers. Redis stores both GEO candidate data and direct latest-position snapshots while preserving listen preference sets.

**Tech Stack:** Go, go-zero style RPC/API code, GORM repositories, go-redis, miniredis tests, MySQL migrations/docs.

## Global Constraints

- Do not change bottom-layer/framework code.
- Do not change non-driver business behavior.
- Keep `POST /api/driver/v1/drivers/location/report` as the driver-facing endpoint.
- Keep driver identity from JWT; clients must not pass `driverId`.
- Keep `driversvc.ReportLocation` reusable from other driver-side RPC/API paths.
- Location heartbeat must not reset existing listen preferences.

---

### Task 1: Redis Position Snapshot

**Files:**
- Modify: `rpc/driversvc/internal/logic/dispatch_availability_test.go`
- Modify: `rpc/driversvc/internal/logic/dispatch_availability.go`

**Interfaces:**
- Consumes: `constants.RedisDriverPos`, `constants.RedisDriverOnline`, `constants.RedisDriverGeo`.
- Produces: `syncDispatchDriverOnline(ctx, svcCtx, driverID, longitude, latitude)` writes hash `driver:pos:<driver_id>` with `driver_id`, `longitude`, `latitude`, `report_time`; `syncDispatchDriverOffline` deletes the hash.

- [ ] Add failing test `TestSyncDispatchDriverOnlineWritesPositionSnapshot` that calls `syncDispatchDriverOnline` and asserts Redis hash fields under `fmt.Sprintf(constants.RedisDriverPos, 25)`.
- [ ] Run `go test ./rpc/driversvc/internal/logic -run TestSyncDispatchDriverOnlineWritesPositionSnapshot -count=1` and confirm it fails because the hash is missing.
- [ ] Add `HSet` and `Expire`/`Del` behavior to the existing Redis pipeline.
- [ ] Re-run the focused test and confirm it passes.

### Task 3: Contract And Docs

**Files:**
- Modify: `docs/driver/13-driver-location-report.md`

**Interfaces:**
- Produces: documented behavior for MySQL latest location, Redis GEO/online/latest snapshot, optional locationsvc dual write, and trajectory write when `orderId` is present.

- [ ] Update docs with request fields `deviceId`, `longitude`, `latitude`, `heading`, `speedKmh`, `orderId`.
- [ ] Document that `driversvc.ReportLocation` is the reusable core for other driver-side interfaces.
- [ ] Document Redis keys used by dispatch hall/navigation: `driver:geo:<city>`, `driver:online`, `driver:pos:<driver_id>`.

### Task 4: Verification

**Files:**
- Verify only.

**Interfaces:**
- Produces: evidence that driver location reporting still compiles and dispatch consumers remain compatible.

- [ ] Run `gofmt` on touched Go files.
- [ ] Run `go test ./rpc/driversvc/internal/logic -run "TestSyncDispatchDriverOnlineWritesPositionSnapshot" -count=1`.
- [ ] Run `go test ./api/driver/... ./rpc/driversvc/... ./rpc/dispatchsvc/... -count=1`.
- [ ] Run `git diff --check`.