# Driver RPC Naming Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Normalize the driver RPC contract and its generated artifacts so the service, client, and server symbols use `DriverService` instead of `Driversvc`, and the SMS login method no longer generates `s_m_s`-style files.

**Architecture:** Keep the `rpc/driversvc` directory and the existing API surface, but rename the protobuf service to `DriverService` and the SMS RPC method to `LoginBySms`. Regenerate the driver-side protobuf/go-zero outputs from the service-local command, then update the direct consumers in `api/driver` and `rpc/adminsvc` to use the new client/server symbols.

**Tech Stack:** Go, protobuf, gRPC, goctl/go-zero, Makefile, PowerShell.

## Global Constraints

- Do not change bottom-layer framework code.
- Do not change non-driver business flows.
- Keep the work scoped to the driver RPC service and its direct consumers.
- Regenerate driver RPC code from within `rpc/driversvc` using the service-local path.

---

### Task 1: Normalize the driver RPC contract

**Files:**
- Modify: `rpc/driversvc/proto/driversvc.proto`
- Modify: `rpc/driversvc/proto/driversvc_service_test.go`
- Modify: `Makefile`

**Interfaces:**
- Consumes: existing `CreateDriver`, `RegisterDriver`, and `LoginBySMS` RPC shape in `rpc/driversvc/proto/driversvc.proto`
- Produces: `DriverService` service symbols, `LoginBySms` RPC method name, and the updated `driver-goctl` Makefile target

- [ ] **Step 1: Rename the gRPC service and the SMS login RPC in the proto**

```proto
service DriverService {
  rpc CreateDriver(CreateDriverRequest) returns (CreateDriverResponse);
  rpc RegisterDriver(CreateDriverRequest) returns (CreateDriverResponse);
  rpc LoginBySms(LoginBySMSRequest) returns (LoginResponse);
}
```

- [ ] **Step 2: Update the driver RPC registration test to the new service descriptor name**

```go
func TestDriverExternalRPCMethodsAreUniqueAndRegistered(t *testing.T) {
	seen := make(map[string]struct{}, len(DriverService_ServiceDesc.Methods))
	for _, method := range DriverService_ServiceDesc.Methods {
		if _, ok := seen[method.MethodName]; ok {
			t.Fatalf("duplicate driver RPC method %q", method.MethodName)
		}
		seen[method.MethodName] = struct{}{}
	}
}
```

- [ ] **Step 3: Keep the generator command service-local and give it a canonical target name**

```make
driver-goctl:
	cd rpc/driversvc && goctl rpc protoc proto/driversvc.proto \
	--go_out=proto \
	--go-grpc_out=proto \
	--zrpc_out=. \
	--style=go_zero

driversvc-goctl: driver-goctl
```

### Task 2: Regenerate the driver RPC outputs

**Files:**
- Modify: `rpc/driversvc/proto/driversvc.pb.go`
- Modify: `rpc/driversvc/proto/driversvc_grpc.pb.go`
- Modify: `rpc/driversvc/internal/server/driversvc_server.go` and the regenerated `rpc/driversvc/internal/server/driver_server.go`
- Modify: `rpc/driversvc/driversvc/driversvc.go` and the regenerated driver wrapper under `rpc/driversvc/`
- Delete: stale generated `rpc/driversvc/driversvcclient/driversvc.go` and `rpc/driversvc/driversvc/driversvc.go` artifacts if goctl stops producing them

**Interfaces:**
- Consumes: the renamed `Driver` service and `LoginBySms` RPC method from Task 1
- Produces: `DriverServiceClient`, `NewDriverServiceClient`, `DriverServiceServer`, `RegisterDriverServiceServer`, and the cleaned driver-side generated tree

- [ ] **Step 1: Run the service-local goctl command from `rpc/driversvc`**

```powershell
Set-Location D:\gocode\src\XiaoLong-Ridy\rpc\driversvc

goctl rpc protoc proto/driversvc.proto --go_out=proto --go-grpc_out=proto --zrpc_out=. --style=go_zero
```

- [ ] **Step 2: Inspect the new generated tree and remove obsolete driver wrapper paths if they are no longer referenced**

```powershell
git status --short rpc/driversvc
```

Expected: the generated code should now be centered on `Driver` symbols, and any old `driversvc` wrapper directories that are no longer generated should be deleted.

- [ ] **Step 3: Verify the proto-generated service descriptor names match the new contract**

```go
const (
	Driver_CreateDriver_FullMethodName = "/driversvc.Driver/CreateDriver"
	Driver_LoginBySms_FullMethodName   = "/driversvc.Driver/LoginBySms"
)
```

### Task 3: Rewire driver-side logic and direct consumers

**Files:**
- Modify: `rpc/driversvc/driversvc.go`
- Modify: `rpc/driversvc/internal/server/driversvc_server.go` and the regenerated `rpc/driversvc/internal/server/driver_server.go`
- Modify: `rpc/driversvc/internal/logic/register_driver_logic.go`
- Modify: `rpc/driversvc/internal/logic/login_logic.go` and the regenerated `rpc/driversvc/internal/logic/login_by_sms_logic.go`
- Modify: `rpc/driversvc/client/local.go`
- Modify: `api/driver/internal/svc/service_context.go`
- Modify: `api/driver/internal/logic/auth_logic.go`
- Modify: `api/driver/internal/logic/driver_client_test.go`
- Modify: `api/driver/internal/logic/auth_logic_test.go`
- Modify: `rpc/adminsvc/internal/svc/servicecontext.go`
- Modify: `rpc/adminsvc/internal/logic/adminservice/driver_helpers.go`
- Modify: `rpc/adminsvc/internal/logic/adminservice/driver_audit_logic_test.go`

**Interfaces:**
- Consumes: `DriverServiceClient`, `NewDriverServiceClient`, `RegisterDriverServiceServer`, `LoginBySms`, and the existing `CreateDriverLogic` / `LoginLogic` implementations
- Produces: working driver RPC entrypoint, driver API client calls, and admin driver-audit calls that compile against the renamed symbols

- [ ] **Step 1: Update the driver RPC entrypoint and server registration to the new server symbol names**

```go
proto.RegisterDriverServer(grpcServer, server.NewDriverServer(ctx))
```

- [ ] **Step 2: Make `RegisterDriver` reuse the existing create-driver path instead of returning an empty response**

```go
func (l *RegisterDriverLogic) RegisterDriver(in *proto.CreateDriverRequest) (*proto.CreateDriverResponse, error) {
	return NewCreateDriverLogic(l.ctx, l.svcCtx).CreateDriver(in)
}
```

- [ ] **Step 3: Make the SMS login RPC call into the existing login flow instead of leaving a generated stub behind**

```go
func (l *LoginBySmsLogic) LoginBySms(in *proto.LoginBySMSRequest) (*proto.LoginResponse, error) {
	return NewLoginLogic(l.ctx, l.svcCtx).LoginBySms(in)
}
```

- [ ] **Step 4: Rename driver API and admin driver-client call sites from `DriversvcClient` / `NewDriversvcClient` to `DriverServiceClient` / `NewDriverServiceClient`, and from `LoginBySMS` to `LoginBySms` where the generated RPC interface changed**

```go
client := driversproto.NewDriverClient(driverConn)
resp, err := client.LoginBySms(l.ctx, &driversproto.LoginBySMSRequest{Phone: strings.TrimSpace(req.Phone)})
```

- [ ] **Step 5: Update the local driver test doubles and client adapters so the renamed interface still satisfies the same API behavior**

```go
type fakeDriverClient struct {
	driversproto.DriverClient
}
```

### Task 4: Run focused verification for the renamed driver RPC

**Files:**
- Test: `rpc/driversvc/...`
- Test: `api/driver/...`
- Test: `rpc/adminsvc/...`

**Interfaces:**
- Consumes: the renamed driver RPC symbols from Tasks 1-3
- Produces: passing compile and behavior checks for the driver RPC, driver API, and admin driver audit flow

- [ ] **Step 1: Run the driver RPC package tests**

```bash
go test ./rpc/driversvc/...
```

- [ ] **Step 2: Run the driver API tests that compile against the driver RPC client**

```bash
go test ./api/driver/...
```

- [ ] **Step 3: Run the admin RPC tests that depend on the driver RPC client**

```bash
go test ./rpc/adminsvc/...
```

- [ ] **Step 4: Fix any remaining symbol mismatches until the three focused test runs pass cleanly**

```text
Expected: no references remain to `DriversvcClient`, `NewDriversvcClient`, `RegisterDriversvcServer`, or `LoginBySms` where the renamed RPC interface is now in use.
```
