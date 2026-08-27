# Driver Reject Reason Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the driver reject-order API require a reject reason and persist it in a dedicated dispatch_record.reject_reason field while keeping the existing front-end endpoint.

**Architecture:** Keep the public HTTP route `POST /api/driver/v1/orders/reject`. Validate and trim `reason` at the driver API boundary, pass it through dispatchsvc, store it on `dispatch_record.reject_reason`, and also keep `remark` populated for compatibility with existing list/detail displays.

**Tech Stack:** Go, net/http handlers, go-zero style RPC services, protobuf, GORM repositories, in-memory repositories, MySQL migration SQL, Go tests.

## Global Constraints

- Do not modify framework or bottom-layer code.
- Keep scope to the driver reject-order path and its direct dispatchsvc persistence dependency.
- Do not run driver proto generation from the repository root.
- Preserve the existing front-end route: `POST /api/driver/v1/orders/reject`.
- Use TDD: write failing tests before production code changes.

---

### Task 1: Driver API reason validation

**Files:**
- Modify: `api/driver/internal/logic/order_logic_test.go`
- Modify: `api/driver/internal/logic/order_logic.go`
- Modify: `docs/driver/17-order-reject.md`

**Interfaces:**
- Consumes: `types.RejectOrderRequest{OrderID int64, Reason string}`
- Produces: `OrderLogic.RejectOrder(driverID int64, req *types.RejectOrderRequest)` returns `ErrInvalidParam` when trimmed reason is empty, and forwards trimmed reason to `dispatchproto.RejectDispatchRequest.Reason`.

- [ ] **Step 1: Write failing tests**

```go
func TestRejectOrderRequiresReason(t *testing.T) {
    client := &fakeDispatchClient{}
    logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: client})

    _, err := logic.RejectOrder(25, &types.RejectOrderRequest{OrderID: 1001, Reason: "   "})

    if err != ErrInvalidParam {
        t.Fatalf("RejectOrder() error = %v, want %v", err, ErrInvalidParam)
    }
    if client.rejectRequest != nil {
        t.Fatalf("RejectOrder() should not call dispatchsvc for empty reason, got %+v", client.rejectRequest)
    }
}
```

```go
func TestRejectOrderTrimsReasonBeforeDispatch(t *testing.T) {
    client := &fakeDispatchClient{}
    logic := NewOrderLogic(context.Background(), &svc.ServiceContext{DispatchClient: client})

    _, err := logic.RejectOrder(25, &types.RejectOrderRequest{OrderID: 1001, Reason: "  too far  "})

    if err != nil {
        t.Fatalf("RejectOrder() error = %v", err)
    }
    if client.rejectRequest.GetReason() != "too far" {
        t.Fatalf("RejectOrder() reason = %q, want too far", client.rejectRequest.GetReason())
    }
}
```

- [ ] **Step 2: Verify RED**

Run: `go test ./api/driver/internal/logic -run "TestRejectOrderRequiresReason|TestRejectOrderTrimsReasonBeforeDispatch" -count=1`

Expected: at least one test fails because `RejectOrder` currently forwards the raw reason and allows empty reason.

- [ ] **Step 3: Implement validation**

In `OrderLogic.RejectOrder`, compute `reason := strings.TrimSpace(req.Reason)`, return `ErrInvalidParam` when empty, and send `Reason: reason` to dispatchsvc.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./api/driver/internal/logic -run "TestRejectOrderRequiresReason|TestRejectOrderTrimsReasonBeforeDispatch" -count=1`

Expected: PASS.

### Task 2: Dedicated dispatch reject reason persistence

**Files:**
- Modify: `rpc/dispatchsvc/internal/model/dispatch_record.go`
- Modify: `rpc/dispatchsvc/internal/repository/gorm_dispatch_repository.go`
- Modify: `rpc/dispatchsvc/internal/repository/memory_dispatch_repository.go`
- Modify: `rpc/dispatchsvc/internal/logic/reject_dispatch_logic.go`
- Modify: `rpc/dispatchsvc/internal/logic/dispatch_status_flow_test.go`
- Modify: `scripts/sql/migrate/04_order_dispatch_module.sql`

**Interfaces:**
- Consumes: `DispatchRepository.RejectByOrderAndDriver(ctx, orderID, driverID uint64, reason string)`.
- Produces: `model.DispatchRecord.RejectReason string` mapped to `dispatch_record.reject_reason`; rejecting writes both `RejectReason` and `Remark` to the reason.

- [ ] **Step 1: Write failing dispatchsvc test**

```go
if record.DriverId == 9001 {
    if record.RejectReason != "too far" {
        t.Fatalf("reject reason = %q, want too far", record.RejectReason)
    }
    if record.Remark != "too far" {
        t.Fatalf("remark = %q, want too far", record.Remark)
    }
}
```

- [ ] **Step 2: Verify RED**

Run: `go test ./rpc/dispatchsvc/internal/logic -run TestRejectDispatchFlow -count=1`

Expected: build fails because `DispatchRecord.RejectReason` does not exist, or the assertion fails.

- [ ] **Step 3: Implement persistence**

Add `RejectReason string `gorm:"column:reject_reason;size:255;default:''" json:"rejectReason"`` to the model. Update GORM and memory repositories to set both `reject_reason` and `remark` during reject. Trim reason in `RejectDispatch`; empty RPC reason can still default to `driver rejected order` for direct RPC compatibility.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./rpc/dispatchsvc/internal/logic ./rpc/dispatchsvc/internal/repository -count=1`

Expected: PASS.

### Task 3: Expose rejectReason to front-end dispatch list

**Files:**
- Modify: `rpc/dispatchsvc/proto/dispatchsvc.proto`
- Modify generated: `rpc/dispatchsvc/proto/dispatchsvc.pb.go`
- Modify: `rpc/dispatchsvc/internal/logic/list_dispatch_records_logic.go`
- Modify: `api/driver/internal/types/types.go`
- Modify: `api/driver/driver.api`
- Modify: `api/driver/internal/logic/order_logic.go`
- Modify: `api/driver/internal/logic/order_logic_test.go`
- Modify: `docs/driver/17-order-reject.md`
- Modify relevant dispatch listing docs if present.

**Interfaces:**
- Produces: `dispatchsvc.DispatchRecord.reject_reason` with Go getter `GetRejectReason()`.
- Produces: driver API dispatch list item JSON field `rejectReason`.

- [ ] **Step 1: Write failing API list test**

Update the fake dispatch list response with `RejectReason: "too far"` and assert `resp.List[0].Dispatch.RejectReason == "too far"`.

- [ ] **Step 2: Verify RED**

Run: `go test ./api/driver/internal/logic -run TestListMyDispatchesCombinesDispatchAndOrder -count=1`

Expected: build fails because the API `DispatchRecord` type lacks `RejectReason`, or assertion fails.

- [ ] **Step 3: Implement exposure**

Add `reject_reason = 10` to dispatch proto. Regenerate or manually update generated dispatch proto consistently. Map model `RejectReason` into RPC `DispatchRecord.RejectReason`, and map RPC `GetRejectReason()` into driver API `types.DispatchRecord.RejectReason`.

- [ ] **Step 4: Verify GREEN**

Run: `go test ./api/driver/internal/logic ./rpc/dispatchsvc/... -count=1`

Expected: PASS.

### Task 4: Final verification

**Files:**
- Verify: all modified files.

- [ ] **Step 1: Format**

Run: `gofmt -w api/driver/internal/logic/order_logic.go api/driver/internal/logic/order_logic_test.go api/driver/internal/types/types.go rpc/dispatchsvc/internal/model/dispatch_record.go rpc/dispatchsvc/internal/repository/gorm_dispatch_repository.go rpc/dispatchsvc/internal/repository/memory_dispatch_repository.go rpc/dispatchsvc/internal/logic/reject_dispatch_logic.go rpc/dispatchsvc/internal/logic/dispatch_status_flow_test.go rpc/dispatchsvc/internal/logic/list_dispatch_records_logic.go rpc/dispatchsvc/proto/dispatchsvc.pb.go`

- [ ] **Step 2: Test focused packages**

Run: `go test ./api/driver/... ./rpc/dispatchsvc/... -count=1`

Expected: PASS.

- [ ] **Step 3: Check diff**

Run: `git diff --check`

Expected: no whitespace errors or conflict markers.