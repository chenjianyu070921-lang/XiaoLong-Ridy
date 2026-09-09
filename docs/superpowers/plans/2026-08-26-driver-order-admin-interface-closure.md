# Driver Order/Admin Interface Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the driver/admin order and driver-record interfaces usable for frontend and external integration.

**Architecture:** Keep all changes at API/RPC business boundaries. Admin driver list/detail reads driver tables through adminsvc. Driver order detail still reads ordersvc and applies token-driver authorization in api/driver.

**Tech Stack:** Go, net/http handlers, existing gRPC clients, Vue/Vant admin and driver web modules.

## Global Constraints

- Do not modify bottom-layer framework code.
- Do not touch non-driver business logic.
- Do not run the broken root-level goctl rpc protoc command.
- Use existing proto files and generated packages unless a missing RPC contract forces regeneration.

---

### Task 1: Driver Order Detail Authorization

**Files:**
- Modify: `api/driver/internal/logic/order_logic_test.go`
- Modify: `api/driver/internal/logic/order_logic.go`

**Interfaces:**
- Consumes: `ordersvc.GetOrder(GetOrderRequest{OrderId})`
- Produces: `OrderLogic.GetMyOrderDetail(driverID int64, orderID int64) (*types.GetMyOrderDetailResponse, error)`

- [ ] **Step 1: Write the failing test**

```go
func TestGetMyOrderDetailAllowsUnassignedWaitAcceptOrder(t *testing.T) {
	client := &fakeOrderClient{getOrderResponseDriverID: -1, getOrderResponseStatus: orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT}
	logic := NewOrderLogic(context.Background(), &svc.ServiceContext{OrderClient: client})

	resp, err := logic.GetMyOrderDetail(25, 1001)
	if err != nil {
		t.Fatalf("GetMyOrderDetail() error = %v", err)
	}
	if resp.Order.DriverID != 0 || resp.Order.Status != int32(orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT) {
		t.Fatalf("GetMyOrderDetail() response = %+v", resp)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./api/driver/internal/logic -run TestGetMyOrderDetailAllowsUnassignedWaitAcceptOrder -count=1`
Expected: FAIL with forbidden access.

- [ ] **Step 3: Implement minimal authorization change**

Allow detail when `order.driver_id == driverID`, or when `order.driver_id == 0 && order.status == ORDER_STATUS_WAIT_ACCEPT`; otherwise return `ErrForbiddenDriverResource`.

- [ ] **Step 4: Run tests**

Run: `go test ./api/driver/internal/logic -run "TestGetMyOrderDetail|TestListMyDispatches|TestListMyOrders|TestListAvailableOrders" -count=1`

### Task 2: Admin Driver Base Interfaces

**Files:**
- Modify: `rpc/adminsvc/admin.proto`
- Modify generated adminsvc files only if required by existing project conventions.
- Modify: `rpc/adminsvc/internal/server/adminservice/adminserviceserver.go`
- Create: `rpc/adminsvc/internal/logic/adminservice/listdriverslogic.go`
- Create: `rpc/adminsvc/internal/logic/adminservice/getdriverlogic.go`
- Modify: `api/admin/internal/types/types.go`
- Modify: `api/admin/internal/logic/driver_logic.go`
- Modify: `api/admin/internal/handler/router.go`
- Modify: `api/admin/internal/handler/router_test.go`
- Modify: `web/admin/src/api/modules.js`
- Modify: `web/admin/src/views/console/index.vue`

**Interfaces:**
- Produces: `GET /admin/v1/drivers`
- Produces: `GET /admin/v1/drivers/{id}`

- [ ] **Step 1: Add failing HTTP route tests**

Add test cases for `/admin/v1/drivers?page=1&page_size=20` and `/admin/v1/drivers/10` expecting successful responses from fake admin service.

- [ ] **Step 2: Run test to verify failure**

Run: `go test ./api/admin/internal/handler -run TestAdminRoutes -count=1`
Expected: FAIL or empty placeholder behavior for drivers.

- [ ] **Step 3: Add adminsvc driver RPC and logic**

Add `ListDrivers` and `GetDriver` RPC methods with list/detail messages. Implement SQL query against `driver` and left join current vehicle/certification summary.

- [ ] **Step 4: Wire admin HTTP and frontend**

Route `/admin/v1/drivers` and `/admin/v1/drivers/{id}` through `DriverLogic`. Replace `driversApi.unavailable` with real list/detail calls and remove the alert text.

- [ ] **Step 5: Run tests**

Run: `go test ./api/admin/internal/handler ./api/admin/internal/logic ./rpc/adminsvc/internal/logic/adminservice -count=1`

### Task 3: Driver Web Order/Dispatch Links

**Files:**
- Modify: `web/user/src/views/driver/DriverHome.vue`
- Modify: `web/user/src/api/driver.js` if path aliases are needed.

**Interfaces:**
- Consumes: `getDriverOrderDetail(orderId)`
- Consumes: `listDriverDispatches(payload)`
- Consumes: `listAvailableOrders(payload)`

- [ ] **Step 1: Normalize order item IDs**

Ensure dispatch, available, and my-order records all expose `orderId`.

- [ ] **Step 2: Fix action visibility**

Allow accept/reject for available orders and dispatch records only when dispatch status is pending or order status is wait-accept.

- [ ] **Step 3: Fix detail link behavior**

Keep details on the driver API and show errors with existing `safeApiCall` path.

- [ ] **Step 4: Verify frontend compiles**

Run the admin/user web build or lint command available in the repo.
