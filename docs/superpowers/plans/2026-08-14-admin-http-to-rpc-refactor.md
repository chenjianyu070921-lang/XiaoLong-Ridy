# admin HTTP-to-RPC Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move admin backend business actions from `api/admin` into `rpc/adminsvc`, while keeping HTTP routes and Postman-facing paths stable.

**Architecture:** `api/admin` becomes a thin adapter that performs auth, request parsing, path dispatch, and response formatting. `rpc/adminsvc` owns the admin business logic for users, drivers, orders, coupons, and operation logs, with proto-defined contracts between layers. Docs are updated to reflect the new responsibility split and available endpoints.

**Tech Stack:** Go, go-zero style RPC, protobuf, net/http, MySQL, Redis.

---

### Task 1: Extend adminsvc RPC contract for admin actions

**Files:**
- Modify: `rpc/adminsvc/admin.proto`
- Modify: `rpc/adminsvc/adminsvc/admin.pb.go`
- Modify: `rpc/adminsvc/adminsvc/admin_grpc.pb.go`
- Modify: `rpc/adminsvc/client/adminservice/adminservice.go`

- [ ] **Step 1: Add RPC request/response messages for user freeze/unfreeze, coupon create/update/status, abnormal order query, and operation log query.**

```proto
message ChangeUserStatusRequest {
  int64 id = 1;
  int32 status = 2;
  string reason = 3;
  string remark = 4;
  int64 admin_id = 5;
  string ip = 6;
}

message CouponRequest {
  int64 id = 1;
  string name = 2;
  int32 type = 3;
  string face_value = 4;
  string discount = 5;
  string threshold_amount = 6;
  int64 total_count = 7;
  int32 per_user_limit = 8;
  string valid_start_at = 9;
  string valid_end_at = 10;
  int32 status = 11;
  int64 admin_id = 12;
  string ip = 13;
}

message AbnormalOrderListRequest {
  int32 page = 1;
  int32 page_size = 2;
  string keyword = 3;
  string abnormal_type = 4;
  int64 user_id = 5;
  int64 driver_id = 6;
  string start_time = 7;
  string end_time = 8;
}
```

- [ ] **Step 2: Add RPC methods for the new admin actions.**

```proto
rpc FreezeUser(ChangeUserStatusRequest) returns (CommonResponse);
rpc UnfreezeUser(ChangeUserStatusRequest) returns (CommonResponse);
rpc CreateCoupon(CouponRequest) returns (CommonResponse);
rpc UpdateCoupon(CouponRequest) returns (CommonResponse);
rpc DisableCoupon(CouponRequest) returns (CommonResponse);
rpc ListAbnormalOrders(AbnormalOrderListRequest) returns (OrderListResponse);
```

- [ ] **Step 3: Regenerate proto outputs with the repo's existing protobuf/goctl flow.**

Run: `go generate ./...` or the repo's documented proto generation command if present.
Expected: `admin.pb.go`, `admin_grpc.pb.go`, and client bindings include the new methods.

- [ ] **Step 4: Verify the generated service interface contains the new methods.**

Run: `rg -n "FreezeUser|UnfreezeUser|CreateCoupon|UpdateCoupon|DisableCoupon|ListAbnormalOrders" rpc/adminsvc`
Expected: each method appears in proto, generated interface, and client.

### Task 2: Implement core rpc logic in adminsvc

**Files:**
- Modify: `rpc/adminsvc/internal/logic/adminservice/*.go`
- Modify: `rpc/adminsvc/internal/server/adminservice/adminserviceserver.go`
- Modify: `rpc/adminsvc/internal/svc/servicecontext.go`

- [ ] **Step 1: Wire required repositories/models into the adminsvc service context.**

```go
// ServiceContext holds repos and clients required by adminsvc logic.
type ServiceContext struct {
    config.Config
    // add repositories/clients here for user, driver, orderclient, coupon, and log access
}
```

- [ ] **Step 2: Implement user freeze/unfreeze logic with status persistence and audit context.**

```go
func (l *FreezeUserLogic) FreezeUser(in *adminsvc.ChangeUserStatusRequest) (*adminsvc.CommonResponse, error) {
    // validate status, update user record, write operation log
    return &adminsvc.CommonResponse{Message: "ok"}, nil
}
```

- [ ] **Step 3: Implement driver certification approve/reject logic so approve updates driver/vehicle status.**

```go
func (l *ApproveDriverCertificationLogic) ApproveDriverCertification(in *adminsvc.AuditDriverCertificationRequest) (*adminsvc.CommonResponse, error) {
    // update certification, driver, vehicle, and write log
    return &adminsvc.CommonResponse{Message: "ok"}, nil
}
```

- [ ] **Step 4: Implement coupon create/update/disable logic using the existing coupon model.**

```go
func (l *CreateCouponLogic) CreateCoupon(in *adminsvc.CouponRequest) (*adminsvc.CommonResponse, error) {
    // persist coupon template and audit log
    return &adminsvc.CommonResponse{Message: "ok"}, nil
}
```

- [ ] **Step 5: Implement normal order list/detail and abnormal order list query.**

```go
func (l *ListAbnormalOrdersLogic) ListAbnormalOrders(in *adminsvc.AbnormalOrderListRequest) (*adminsvc.OrderListResponse, error) {
    // query abnormal orders by abnormal type and paging
    return &adminsvc.OrderListResponse{}, nil
}
```

- [ ] **Step 6: Run package tests for rpc/adminsvc.**

Run: `go test ./rpc/adminsvc/...`
Expected: build passes and the new logic compiles.

### Task 3: Replace api/admin handler internals with rpc adminsvc calls

**Files:**
- Modify: `api/admin/internal/handler/router.go`
- Modify: `api/admin/internal/logic/*.go`
- Modify: `api/admin/internal/svc/servicecontext.go`
- Modify: `api/admin/internal/types/types.go`

- [ ] **Step 1: Add an adminsvc client to the HTTP service context.**

```go
type ServiceContext struct {
    Config config.Config
    AdminSvc adminservice.AdminService
    SessionRepository *repository.SessionRepository
}
```

- [ ] **Step 2: Convert user/driver/order/coupon/log handlers to call rpc methods instead of repositories.**

```go
resp, err := r.ctx.AdminSvc.ListUsers(req.Context(), &adminsvc.UserListRequest{...})
```

- [ ] **Step 3: Keep auth/login/logout/me locally wired only where necessary, but preserve current route paths and auth middleware behavior.**

```go
r.mux.HandleFunc("/admin/v1/users", r.authRequired(r.handleUsers))
```

- [ ] **Step 4: Ensure error mapping from rpc errors back to HTTP codes stays consistent.**

```go
switch {
case errors.Is(err, logic.ErrUnauthorized):
    writeError(w, http.StatusUnauthorized, 40004, "unauthorized")
}
```

- [ ] **Step 5: Run admin package tests and a focused build.**

Run: `go test ./api/admin/...`
Expected: handler layer compiles against rpc client types.

### Task 4: Update admin docs to reflect the rpc boundary

**Files:**
- Modify: `docs/api/管理后台接口文档.md`
- Modify: `docs/admin/管理后台全模块RPC依赖调用清单.md`

- [ ] **Step 1: Rewrite the intro section so api/admin is documented as HTTP/auth/parameter translation only.**

```md
> api/admin 仅负责 HTTP 接入层、鉴权、参数转换和响应包装。
> 所有用户、司机、订单、优惠券和操作日志核心动作均下沉至 rpc/adminsvc。
```

- [ ] **Step 2: Update each endpoint section to list the rpc method it calls.**

```md
- HTTP: `GET /admin/v1/users`
- RPC: `AdminService.ListUsers`
```

- [ ] **Step 3: Update the dependency table so it matches the actual call chain after the refactor.**

```md
| HTTP handler | rpc/adminsvc method | purpose |
```

- [ ] **Step 4: Review for stale direct-DB wording and remove any claims that admin HTTP directly owns business persistence.**

Run: `rg -n "直接|repository|MySQL|Redis" docs/api/管理后台接口文档.md docs/admin/管理后台全模块RPC依赖调用清单.md`
Expected: no stale ownership statements remain.

### Task 5: End-to-end verification

**Files:**
- No new files

- [ ] **Step 1: Start the required local dependencies if they are not already running.**

Run: `docker ps`
Expected: MySQL and Redis are available on local ports.

- [ ] **Step 2: Build or run the admin HTTP and rpc services.**

Run: `go test ./api/admin/... ./rpc/adminsvc/...`
Expected: both modules compile.

- [ ] **Step 3: Exercise the P0/P1 admin endpoints with Postman or curl.**

Run: `GET /admin/v1/users`, `GET /admin/v1/orders`, `GET /admin/v1/driver-certifications`, `GET /admin/v1/coupons`
Expected: responses flow through rpc/adminsvc.

- [ ] **Step 4: Commit the refactor after tests pass.**

```bash
git add api/admin rpc/adminsvc docs/api/管理后台接口文档.md docs/admin/管理后台全模块RPC依赖调用清单.md
git commit -m "feat(admin): move core admin actions to rpc adminsvc"
```
