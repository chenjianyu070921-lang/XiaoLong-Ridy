package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/jwtx"
	dispatchproto "XiaoLong-Ridy/rpc/dispatchsvc/proto"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

func TestAgentChatEndpointRequiresDriverTokenAndRunsAgent(t *testing.T) {
	const signingKey = "agent-chat-test-key"
	const serviceToken = "agent-chat-service-token"
	t.Setenv("DRIVER_AGENT_SERVICE_TOKEN", serviceToken)
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: signingKey})

	unauthorizedRequest := httptest.NewRequest(http.MethodPost, "/api/driver/v1/agent/chat", bytes.NewBufferString(`{"question":"price for product 1001"}`))
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorizedRequest)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorizedResponse.Code, http.StatusUnauthorized)
	}

	serviceRequest := httptest.NewRequest(http.MethodPost, "/api/driver/v1/agent/chat", bytes.NewBufferString(`{"question":"price for product 1001"}`))
	serviceRequest.Header.Set("X-Internal-Service-Token", serviceToken)
	serviceResponse := httptest.NewRecorder()
	handler.ServeHTTP(serviceResponse, serviceRequest)
	if serviceResponse.Code != http.StatusOK {
		t.Fatalf("service status = %d, want %d: %s", serviceResponse.Code, http.StatusOK, serviceResponse.Body.String())
	}

	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     1,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, signingKey)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/agent/chat", bytes.NewBufferString(`{"question":"price for product 1001"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var body struct {
		Code int `json:"code"`
		Data struct {
			Answer       string `json:"answer"`
			LoopCount    int    `json:"loopCount"`
			Mode         string `json:"mode"`
			Observations []struct {
				ToolName string `json:"toolName"`
			} `json:"observations"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.Answer == "" || body.Data.LoopCount != 2 || body.Data.Mode != "scripted" || len(body.Data.Observations) != 1 || body.Data.Observations[0].ToolName != "get_product_price" {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

type pollingOrderClient struct {
	listOrdersRequest *orderproto.ListOrdersRequest
}

func (p *pollingOrderClient) GetOrder(context.Context, *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	return nil, nil
}

func (p *pollingOrderClient) ListOrders(_ context.Context, req *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error) {
	p.listOrdersRequest = req
	return &orderproto.ListOrdersResponse{
		List: []*orderproto.OrderSummary{{
			OrderId:             1001,
			OrderNo:             "NO-1001",
			FromAddress:         "from",
			ToAddress:           "to",
			Status:              orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
			EstimatedPriceCents: 29900,
			CreatedAt:           123,
		}},
		Total:    1,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

func (p *pollingOrderClient) AcceptOrder(context.Context, *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error) {
	return nil, nil
}

func (p *pollingOrderClient) CancelOrder(context.Context, *orderproto.CancelOrderRequest) (*orderproto.CancelOrderResponse, error) {
	return nil, nil
}

func (p *pollingOrderClient) StartTrip(context.Context, *orderproto.StartTripRequest) (*orderproto.StartTripResponse, error) {
	return nil, nil
}

func (p *pollingOrderClient) ConfirmArrive(context.Context, *orderproto.ConfirmArriveRequest) (*orderproto.ConfirmArriveResponse, error) {
	return nil, nil
}

func (p *pollingOrderClient) FinishTrip(context.Context, *orderproto.FinishTripRequest) (*orderproto.FinishTripResponse, error) {
	return nil, nil
}

func internalHTTPTestServiceContext(token string, driverClient svc.DriverClient) *svc.ServiceContext {
	return &svc.ServiceContext{
		SigningKey:   "internal-http-test-key",
		InternalAuth: internalHTTPTestAuth(token),
		DriverClient: driverClient,
	}
}

func internalHTTPTestAuth(token string) svc.InternalAuthConfig {
	return svc.InternalAuthConfig{
		ServiceToken: token,
		AllowedRoutes: []svc.InternalRouteConfig{
			{Method: http.MethodGet, Path: "/api/driver/v1/drivers/get"},
			{Method: http.MethodGet, Path: "/api/driver/v1/vehicles/get"},
			{Method: http.MethodPost, Path: "/api/driver/v1/drivers/nearby"},
			{Method: http.MethodPost, Path: "/api/driver/v1/orders/dispatches"},
		},
		RateLimit: svc.InternalRateLimitConfig{Limit: 60, WindowSeconds: 60},
	}
}

type httpTestDriverClient struct {
	getDriverRequest    *driversproto.GetDriverRequest
	getVehicleRequest   *driversproto.GetVehicleRequest
	nearbyDriverRequest *driversproto.ListNearbyDriversRequest
}

func (c *httpTestDriverClient) CreateDriver(context.Context, *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) RegisterDriver(context.Context, *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) UpdateDriver(context.Context, *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) GetDriver(_ context.Context, req *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error) {
	c.getDriverRequest = req
	return &driversproto.GetDriverResponse{Driver: &driversproto.Driver{
		Id:           req.GetId(),
		Phone:        "13800000000",
		RealName:     "driver",
		IdCardNo:     "110101199001010011",
		Status:       driversproto.DriverStatus_DRIVER_STATUS_NORMAL,
		CreatedAt:    100,
		UpdatedAt:    200,
		OnlineStatus: 1,
	}}, nil
}

func (c *httpTestDriverClient) GetDriverByPhone(context.Context, *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) SetDriverOnline(context.Context, *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) SetDriverOffline(context.Context, *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) SetDriverListenPreference(context.Context, *driversproto.SetDriverListenPreferenceRequest) (*driversproto.DriverListenPreferenceResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) GetDriverListenPreference(context.Context, *driversproto.GetDriverListenPreferenceRequest) (*driversproto.DriverListenPreferenceResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) ReportLocation(context.Context, *driversproto.ReportLocationRequest) (*driversproto.ReportLocationResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) SetDriverServiceStatus(context.Context, *driversproto.SetDriverServiceStatusRequest) (*driversproto.SetDriverServiceStatusResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) Heartbeat(context.Context, *driversproto.HeartbeatRequest) (*driversproto.HeartbeatResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) DeleteDriver(context.Context, *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) Login(context.Context, *driversproto.LoginRequest) (*driversproto.LoginResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) LoginBySMS(context.Context, *driversproto.LoginBySMSRequest) (*driversproto.LoginResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) CreateVehicle(context.Context, *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) UpdateVehicle(context.Context, *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) DeleteVehicle(context.Context, *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) GetVehicle(_ context.Context, req *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error) {
	c.getVehicleRequest = req
	return &driversproto.GetVehicleResponse{Vehicle: &driversproto.Vehicle{
		Id:       req.GetId(),
		DriverId: 25,
		PlateNo:  "京A12345",
		Status:   driversproto.VehicleStatus_VEHICLE_STATUS_NORMAL,
	}}, nil
}

func (c *httpTestDriverClient) ListNearbyDrivers(_ context.Context, req *driversproto.ListNearbyDriversRequest) (*driversproto.ListNearbyDriversResponse, error) {
	c.nearbyDriverRequest = req
	return &driversproto.ListNearbyDriversResponse{Drivers: []*driversproto.NearbyDriver{{
		DriverId:       25,
		Longitude:      req.GetLongitude(),
		Latitude:       req.GetLatitude(),
		DistanceMeters: 100,
	}}}, nil
}

func (c *httpTestDriverClient) GetDriverAiScore(context.Context, *driversproto.GetDriverAiScoreRequest) (*driversproto.GetDriverAiScoreResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) UploadCertification(context.Context, *driversproto.UploadCertificationRequest) (*driversproto.UploadCertificationResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) GetCertification(context.Context, *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) CreateWithdraw(context.Context, *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error) {
	return nil, nil
}

func (c *httpTestDriverClient) ListWithdraws(context.Context, *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error) {
	return nil, nil
}

type httpTestDispatchClient struct {
	listRequest *dispatchproto.ListDispatchRecordsRequest
}

func (c *httpTestDispatchClient) RejectDispatch(context.Context, *dispatchproto.RejectDispatchRequest) (*dispatchproto.RejectDispatchResponse, error) {
	return nil, nil
}

func (c *httpTestDispatchClient) ListDispatchRecords(_ context.Context, req *dispatchproto.ListDispatchRecordsRequest) (*dispatchproto.ListDispatchRecordsResponse, error) {
	c.listRequest = req
	return &dispatchproto.ListDispatchRecordsResponse{List: []*dispatchproto.DispatchRecord{{
		Id:       1,
		OrderId:  1001,
		DriverId: req.GetDriverId(),
		Status:   1,
		Remark:   "admin query",
	}}, Total: 1, Page: req.GetPage(), PageSize: req.GetPageSize()}, nil
}

type httpTestOrderClient struct{}

func (c *httpTestOrderClient) GetOrder(_ context.Context, req *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	return &orderproto.GetOrderResponse{
		OrderId:             req.GetOrderId(),
		OrderNo:             "NO-1001",
		FromAddress:         "pickup",
		ToAddress:           "destination",
		Status:              orderproto.OrderStatus_ORDER_STATUS_ACCEPTED,
		EstimatedPriceCents: 1000,
		CreatedAt:           100,
	}, nil
}

func (c *httpTestOrderClient) ListOrders(context.Context, *orderproto.ListOrdersRequest) (*orderproto.ListOrdersResponse, error) {
	return nil, nil
}

func (c *httpTestOrderClient) AcceptOrder(context.Context, *orderproto.AcceptOrderRequest) (*orderproto.AcceptOrderResponse, error) {
	return nil, nil
}

func (c *httpTestOrderClient) CancelOrder(context.Context, *orderproto.CancelOrderRequest) (*orderproto.CancelOrderResponse, error) {
	return nil, nil
}

func (c *httpTestOrderClient) StartTrip(context.Context, *orderproto.StartTripRequest) (*orderproto.StartTripResponse, error) {
	return nil, nil
}

func (c *httpTestOrderClient) ConfirmArrive(context.Context, *orderproto.ConfirmArriveRequest) (*orderproto.ConfirmArriveResponse, error) {
	return nil, nil
}

func (c *httpTestOrderClient) FinishTrip(context.Context, *orderproto.FinishTripRequest) (*orderproto.FinishTripResponse, error) {
	return nil, nil
}

func TestLoadDriverConfigReadsYamlAndEnvCanOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "driver.yaml")
	if err := os.WriteFile(path, []byte(`
httpAddr: ":18082"
driverGrpcAddr: "driversvc:50055"
orderGrpcAddr: "ordersvc:50051"
dispatchGrpcAddr: "dispatchsvc:8083"
locationGrpcAddr: "locationsvc:5056"
redisAddr: "redis:6379"
internalAuth:
  serviceToken: "admin-shared-token"
  allowedRoutes:
    - method: GET
      path: /api/driver/v1/drivers/get
    - method: POST
      path: /api/driver/v1/orders/dispatches
  rateLimit:
    limit: 7
    windowSeconds: 30
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadDriverConfig(path)
	if err != nil {
		t.Fatalf("loadDriverConfig() error = %v", err)
	}
	if cfg.HTTPAddr != ":18082" || cfg.DriverGRPCAddr != "driversvc:50055" ||
		cfg.OrderGRPCAddr != "ordersvc:50051" || cfg.DispatchGRPCAddr != "dispatchsvc:8083" ||
		cfg.LocationGRPCAddr != "locationsvc:5056" || cfg.RedisAddr != "redis:6379" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.InternalAuth.ServiceToken != "admin-shared-token" ||
		len(cfg.InternalAuth.AllowedRoutes) != 2 ||
		cfg.InternalAuth.AllowedRoutes[0].Method != http.MethodGet ||
		cfg.InternalAuth.AllowedRoutes[0].Path != "/api/driver/v1/drivers/get" ||
		cfg.InternalAuth.RateLimit.Limit != 7 ||
		cfg.InternalAuth.RateLimit.WindowSeconds != 30 {
		t.Fatalf("unexpected internal auth config: %+v", cfg.InternalAuth)
	}

	t.Setenv("ORDER_GRPC_ADDR", "ordersvc-prod:50051")
	if got := envOr("ORDER_GRPC_ADDR", cfg.OrderGRPCAddr); got != "ordersvc-prod:50051" {
		t.Fatalf("env override = %q", got)
	}
}

func TestInternalServiceTokenAllowsWhitelistedGetDriverWithoutJWT(t *testing.T) {
	driverClient := &httpTestDriverClient{}
	handler := newHTTPHandler(internalHTTPTestServiceContext("internal-token", driverClient))

	request := httptest.NewRequest(http.MethodGet, "/api/driver/v1/drivers/get?id=42", nil)
	request.Header.Set("X-Internal-Service-Token", "internal-token")
	request.Header.Set("X-Trace-Id", "trace-admin-42")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if driverClient.getDriverRequest.GetId() != 42 {
		t.Fatalf("GetDriver() id = %d, want 42", driverClient.getDriverRequest.GetId())
	}
	var body struct {
		TraceID string `json:"traceId"`
		Data    struct {
			Driver struct {
				ID int64 `json:"id"`
			} `json:"driver"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.TraceID != "trace-admin-42" || body.Data.Driver.ID != 42 {
		t.Fatalf("unexpected response: %s", response.Body.String())
	}
}

func TestInternalServiceTokenRejectsNonWhitelistedRouteEvenWithJWT(t *testing.T) {
	handler := newHTTPHandler(internalHTTPTestServiceContext("internal-token", &httpTestDriverClient{}))
	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, "internal-http-test-key")
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/vehicles/delete?id=77", bytes.NewBufferString(`{}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Internal-Service-Token", "internal-token")
	request.Header.Set("X-Trace-Id", "trace-admin-denied")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
	}
}

func TestInternalServiceTokenRejectsBadToken(t *testing.T) {
	handler := newHTTPHandler(internalHTTPTestServiceContext("internal-token", &httpTestDriverClient{}))

	request := httptest.NewRequest(http.MethodGet, "/api/driver/v1/drivers/get?id=42", nil)
	request.Header.Set("X-Internal-Service-Token", "wrong-token")
	request.Header.Set("X-Trace-Id", "trace-admin-bad-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusForbidden, response.Body.String())
	}
}

func TestInternalServiceTokenRequiresTraceID(t *testing.T) {
	handler := newHTTPHandler(internalHTTPTestServiceContext("internal-token", &httpTestDriverClient{}))

	request := httptest.NewRequest(http.MethodGet, "/api/driver/v1/drivers/get?id=42", nil)
	request.Header.Set("X-Internal-Service-Token", "internal-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestDriverJWTStillWorksWithoutInternalToken(t *testing.T) {
	driverClient := &httpTestDriverClient{}
	handler := newHTTPHandler(internalHTTPTestServiceContext("internal-token", driverClient))
	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, "internal-http-test-key")
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/driver/v1/drivers/get?id=42", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if driverClient.getDriverRequest.GetId() != 25 {
		t.Fatalf("GetDriver() id = %d, want JWT driver 25", driverClient.getDriverRequest.GetId())
	}
}

func TestInternalListNearbyDriversReturnsListAndDrivers(t *testing.T) {
	handler := newHTTPHandler(internalHTTPTestServiceContext("internal-token", &httpTestDriverClient{}))

	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/drivers/nearby", bytes.NewBufferString(`{"longitude":116.397,"latitude":39.908,"radiusMeters":3000,"limit":10}`))
	request.Header.Set("X-Internal-Service-Token", "internal-token")
	request.Header.Set("X-Trace-Id", "trace-admin-nearby")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		Data struct {
			List    []any `json:"list"`
			Drivers []any `json:"drivers"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data.List) != 1 || len(body.Data.Drivers) != 1 {
		t.Fatalf("nearby response should include both list and drivers: %s", response.Body.String())
	}
}

func TestInternalListMyDispatchesUsesBodyDriverID(t *testing.T) {
	dispatchClient := &httpTestDispatchClient{}
	handler := newHTTPHandler(&svc.ServiceContext{
		SigningKey:     "internal-http-test-key",
		InternalAuth:   internalHTTPTestAuth("internal-token"),
		DispatchClient: dispatchClient,
		OrderClient:    &httpTestOrderClient{},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/orders/dispatches", bytes.NewBufferString(`{"driverId":77,"page":1,"pageSize":20,"status":1}`))
	request.Header.Set("X-Internal-Service-Token", "internal-token")
	request.Header.Set("X-Trace-Id", "trace-admin-dispatches")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if dispatchClient.listRequest.GetDriverId() != 77 {
		t.Fatalf("ListDispatchRecords() driverId = %d, want 77", dispatchClient.listRequest.GetDriverId())
	}
	var body struct {
		Data struct {
			OrderQueryOk bool `json:"orderQueryOk"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Data.OrderQueryOk {
		t.Fatalf("orderQueryOk = false, want true: %s", response.Body.String())
	}
}

func TestInternalRateLimitDoesNotConsumeDriverJWTTraffic(t *testing.T) {
	driverClient := &httpTestDriverClient{}
	svcCtx := internalHTTPTestServiceContext("internal-token", driverClient)
	svcCtx.InternalAuth.RateLimit = svc.InternalRateLimitConfig{Limit: 1, WindowSeconds: 60}
	handler := newHTTPHandler(svcCtx)

	internal1 := httptest.NewRequest(http.MethodGet, "/api/driver/v1/drivers/get?id=42", nil)
	internal1.Header.Set("X-Internal-Service-Token", "internal-token")
	internal1.Header.Set("X-Trace-Id", "trace-admin-rate-1")
	internalResp1 := httptest.NewRecorder()
	handler.ServeHTTP(internalResp1, internal1)
	if internalResp1.Code != http.StatusOK {
		t.Fatalf("first internal status = %d: %s", internalResp1.Code, internalResp1.Body.String())
	}

	internal2 := httptest.NewRequest(http.MethodGet, "/api/driver/v1/drivers/get?id=42", nil)
	internal2.Header.Set("X-Internal-Service-Token", "internal-token")
	internal2.Header.Set("X-Trace-Id", "trace-admin-rate-2")
	internalResp2 := httptest.NewRecorder()
	handler.ServeHTTP(internalResp2, internal2)
	if internalResp2.Code != http.StatusTooManyRequests {
		t.Fatalf("second internal status = %d, want %d: %s", internalResp2.Code, http.StatusTooManyRequests, internalResp2.Body.String())
	}

	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, "internal-http-test-key")
	if err != nil {
		t.Fatal(err)
	}
	driverRequest := httptest.NewRequest(http.MethodGet, "/api/driver/v1/drivers/get?id=42", nil)
	driverRequest.Header.Set("Authorization", "Bearer "+token)
	driverResponse := httptest.NewRecorder()
	handler.ServeHTTP(driverResponse, driverRequest)
	if driverResponse.Code != http.StatusOK {
		t.Fatalf("driver JWT status = %d, want %d: %s", driverResponse.Code, http.StatusOK, driverResponse.Body.String())
	}
}

func TestOrderEndpointsRequireDriverToken(t *testing.T) {
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "order-route-test-key"})
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/driver/v1/vehicles"},
		{http.MethodGet, "/api/driver/v1/vehicles/get?id=77"},
		{http.MethodPost, "/api/driver/v1/vehicles/update"},
		{http.MethodPost, "/api/driver/v1/vehicles/delete?id=77"},
		{http.MethodPost, "/api/driver/v1/drivers/by-phone?phone=13800000000"},
		{http.MethodPost, "/api/driver/v1/drivers/nearby"},
		{http.MethodPost, "/api/driver/v1/orders/cancel"},
		{http.MethodPost, "/api/driver/v1/orders/reject"},
		{http.MethodPost, "/api/driver/v1/orders/dispatches"},
		{http.MethodPost, "/api/driver/v1/orders/list"},
		{http.MethodGet, "/api/driver/v1/income/summary"},
		{http.MethodGet, "/api/driver/v1/income/today"},
		{http.MethodGet, "/api/driver/v1/income/week"},
		{http.MethodPost, "/api/driver/v1/income/bills"},
		{http.MethodPost, "/api/driver/v1/reviews/list"},
		{http.MethodPost, "/api/driver/v1/orders/trajectory"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			request := httptest.NewRequest(route.method, route.path, bytes.NewBufferString(`{}`))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusUnauthorized, response.Body.String())
			}
		})
	}
}

func TestDriverHTTPExternalRoutesAreRegisteredWithoutConflicts(t *testing.T) {
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "route-conflict-test-key"})
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/driver/v1/img-captcha?phone=13800000000"},
		{http.MethodPost, "/api/driver/v1/img-captcha/verify"},
		{http.MethodPost, "/api/driver/v1/img-captcha/invalidate"},
		{http.MethodPost, "/api/driver/v1/auth/send-sms-code"},
		{http.MethodPost, "/api/driver/v1/auth/login-by-password"},
		{http.MethodPost, "/api/driver/v1/auth/login-by-sms"},
		{http.MethodPost, "/api/driver/v1/drivers/register"},
		{http.MethodGet, "/api/driver/v1/drivers/get"},
		{http.MethodPost, "/api/driver/v1/drivers/update"},
		{http.MethodPost, "/api/driver/v1/drivers/online"},
		{http.MethodPost, "/api/driver/v1/drivers/offline"},
		{http.MethodPost, "/api/driver/v1/drivers/location/report"},
		{http.MethodPost, "/api/driver/v1/vehicles"},
		{http.MethodGet, "/api/driver/v1/vehicles/get?id=1"},
		{http.MethodPost, "/api/driver/v1/orders/available"},
		{http.MethodPost, "/api/driver/v1/orders/detail"},
		{http.MethodPost, "/api/driver/v1/orders/accept"},
		{http.MethodPost, "/api/driver/v1/orders/reject"},
		{http.MethodPost, "/api/driver/v1/orders/start-trip"},
		{http.MethodPost, "/api/driver/v1/orders/confirm-arrive"},
		{http.MethodPost, "/api/driver/v1/orders/finish-trip"},
		{http.MethodPost, "/api/driver/v1/orders/trajectory"},
		{http.MethodPost, "/api/driver/v1/reviews/list"},
		{http.MethodGet, "/api/driver/v1/income/summary"},
		{http.MethodGet, "/api/driver/v1/income/today"},
		{http.MethodGet, "/api/driver/v1/income/week"},
		{http.MethodPost, "/api/driver/v1/income/bills"},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			request := httptest.NewRequest(route.method, route.path, bytes.NewBufferString(`{}`))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code == http.StatusNotFound || response.Code == http.StatusMethodNotAllowed {
				t.Fatalf("route is not externally callable, status = %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestDriverPushWebSocketPollsAssignedOrdersWithoutRedisPublisher(t *testing.T) {
	const signingKey = "driver-ws-poll-test-key"
	orderClient := &pollingOrderClient{}
	server := httptest.NewServer(newHTTPHandler(&svc.ServiceContext{
		SigningKey:       signingKey,
		OrderClient:      orderClient,
		PushPollInterval: 10 * time.Millisecond,
		PushPollPageSize: 20,
	}))
	defer server.Close()

	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, signingKey)
	if err != nil {
		t.Fatal(err)
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/driver/v1/ws?token=" + token
	conn, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var ack struct {
		Type     string `json:"type"`
		Degraded bool   `json:"degraded"`
	}
	if err := websocket.JSON.Receive(conn, &ack); err != nil {
		t.Fatal(err)
	}
	if ack.Type != "connected" || !ack.Degraded {
		t.Fatalf("unexpected ws ack: %+v", ack)
	}

	var msg struct {
		Type    string `json:"type"`
		OrderID int64  `json:"orderId"`
	}
	if err := websocket.JSON.Receive(conn, &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != "dispatch_order" || msg.OrderID != 1001 {
		t.Fatalf("unexpected ws polling message: %+v", msg)
	}
	if orderClient.listOrdersRequest.GetDriverId() != 25 {
		t.Fatalf("ListOrders() driver id = %d, want 25", orderClient.listOrdersRequest.GetDriverId())
	}
}

func TestDriverPushWebSocketAuth(t *testing.T) {
	const signingKey = "driver-ws-test-key"
	server := httptest.NewServer(newHTTPHandler(&svc.ServiceContext{SigningKey: signingKey}))
	defer server.Close()

	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, signingKey)
	if err != nil {
		t.Fatal(err)
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/driver/v1/ws"
	conn, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := websocket.JSON.Send(conn, map[string]string{"type": "auth", "token": token}); err != nil {
		t.Fatal(err)
	}
	var ack struct {
		Type     string `json:"type"`
		DriverID int64  `json:"driverId"`
		Degraded bool   `json:"degraded"`
	}
	if err := websocket.JSON.Receive(conn, &ack); err != nil {
		t.Fatal(err)
	}
	if ack.Type != "connected" || ack.DriverID != 25 || !ack.Degraded {
		t.Fatalf("unexpected ws ack: %+v", ack)
	}
}

func TestDriverPushWebSocketForwardsRedisMessages(t *testing.T) {
	const signingKey = "driver-ws-redis-test-key"
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer rdb.Close()

	server := httptest.NewServer(newHTTPHandler(&svc.ServiceContext{SigningKey: signingKey, RedisClient: rdb}))
	defer server.Close()

	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, signingKey)
	if err != nil {
		t.Fatal(err)
	}

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/driver/v1/ws"
	conn, err := websocket.Dial(wsURL, "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := websocket.JSON.Send(conn, map[string]string{"type": "auth", "token": token}); err != nil {
		t.Fatal(err)
	}
	var ack struct {
		Type     string `json:"type"`
		Degraded bool   `json:"degraded"`
	}
	if err := websocket.JSON.Receive(conn, &ack); err != nil {
		t.Fatal(err)
	}
	if ack.Type != "connected" || ack.Degraded {
		t.Fatalf("unexpected ws ack: %+v", ack)
	}

	payload := `{"type":"dispatch","orderId":1001}`
	if err := rdb.Publish(
		context.Background(),
		fmt.Sprintf(constants.RedisDriverPush, 25),
		payload,
	).Err(); err != nil {
		t.Fatal(err)
	}

	var got string
	if err := websocket.Message.Receive(conn, &got); err != nil {
		t.Fatal(err)
	}
	if got != payload {
		t.Fatalf("ws payload = %s, want %s", got, payload)
	}
}

func TestImgCaptchaGenerateRefreshAndInvalidate(t *testing.T) {
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "captcha-test-key"})

	request := httptest.NewRequest(http.MethodGet, "/api/driver/v1/img-captcha?phone=13800000000", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}

	var body struct {
		Code int `json:"code"`
		Data struct {
			UUID      string `json:"uuid"`
			ImgBase64 string `json:"imgBase64"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.UUID == "" || body.Data.ImgBase64 == "" {
		t.Fatalf("unexpected captcha response: %s", response.Body.String())
	}

	refreshRequest := httptest.NewRequest(http.MethodGet, "/api/driver/v1/img-captcha?phone=13800000000", nil)
	refreshResponse := httptest.NewRecorder()
	handler.ServeHTTP(refreshResponse, refreshRequest)
	if refreshResponse.Code != http.StatusOK {
		t.Fatalf("refresh status = %d, body = %s", refreshResponse.Code, refreshResponse.Body.String())
	}

	verifyOld := httptest.NewRequest(http.MethodPost, "/api/driver/v1/img-captcha/verify", bytes.NewBufferString(fmt.Sprintf(`{"phone":"13800000000","uuid":%q,"userInputCode":"0000"}`, body.Data.UUID)))
	verifyOldResponse := httptest.NewRecorder()
	handler.ServeHTTP(verifyOldResponse, verifyOld)
	if verifyOldResponse.Code != http.StatusBadRequest {
		t.Fatalf("old captcha status = %d, want 400: %s", verifyOldResponse.Code, verifyOldResponse.Body.String())
	}
}

func TestVerifyImgCaptchaHandlerConsumesCanonicalAndLegacyCodeFields(t *testing.T) {
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "captcha-field-test-key"})
	tests := []struct {
		name string
		body string
	}{
		{
			name: "canonical code",
			body: `{"phone":"13800000000","uuid":"missing-captcha","code":"0000"}`,
		},
		{
			name: "legacy userInputCode",
			body: `{"phone":"13800000000","uuid":"missing-captcha","userInputCode":"0000"}`,
		},
		{
			name: "backend captchaCode alias",
			body: `{"phone":"13800000000","captchaId":"missing-captcha","captchaCode":"0000"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/img-captcha/verify", bytes.NewBufferString(tc.body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", response.Code, response.Body.String())
			}
			var body struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != 41002 {
				t.Fatalf("business code = %d, want captcha invalid code 41002: %s", body.Code, response.Body.String())
			}
		})
	}
}
