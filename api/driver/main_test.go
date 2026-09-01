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
	getOrderResponse  *orderproto.GetOrderResponse
}

func (p *pollingOrderClient) GetOrder(context.Context, *orderproto.GetOrderRequest) (*orderproto.GetOrderResponse, error) {
	if p.getOrderResponse != nil {
		return p.getOrderResponse, nil
	}
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

type pollingDispatchClient struct {
	listRequest  *dispatchproto.ListDispatchRecordsRequest
	listResponse *dispatchproto.ListDispatchRecordsResponse
}

func (p *pollingDispatchClient) RejectDispatch(context.Context, *dispatchproto.RejectDispatchRequest) (*dispatchproto.RejectDispatchResponse, error) {
	return nil, nil
}

func (p *pollingDispatchClient) ListDispatchRecords(_ context.Context, req *dispatchproto.ListDispatchRecordsRequest) (*dispatchproto.ListDispatchRecordsResponse, error) {
	p.listRequest = req
	if p.listResponse != nil {
		return p.listResponse, nil
	}
	return &dispatchproto.ListDispatchRecordsResponse{List: []*dispatchproto.DispatchRecord{}, Total: 0, Page: req.GetPage(), PageSize: req.GetPageSize()}, nil
}

func TestLoadDriverConfigReadsYamlAndEnvCanOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "driver.yaml")
	if err := os.WriteFile(path, []byte(`
httpAddr: ":18082"
driverGrpcAddr: "driversvc:5055"
orderGrpcAddr: "ordersvc:50051"
dispatchGrpcAddr: "dispatchsvc:50056"
locationGrpcAddr: "locationsvc:9001"
redisAddr: "redis:6379"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadDriverConfig(path)
	if err != nil {
		t.Fatalf("loadDriverConfig() error = %v", err)
	}
	if cfg.HTTPAddr != ":18082" || cfg.DriverGRPCAddr != "driversvc:5055" ||
		cfg.OrderGRPCAddr != "ordersvc:50051" || cfg.DispatchGRPCAddr != "dispatchsvc:50056" ||
		cfg.LocationGRPCAddr != "locationsvc:9001" || cfg.RedisAddr != "redis:6379" {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	t.Setenv("ORDER_GRPC_ADDR", "ordersvc-prod:50051")
	if got := envOr("ORDER_GRPC_ADDR", cfg.OrderGRPCAddr); got != "ordersvc-prod:50051" {
		t.Fatalf("env override = %q", got)
	}
}

func TestLoadDriverConfigDefaultsUseSharedBackendServer(t *testing.T) {
	cfg, err := loadDriverConfig(filepath.Join(t.TempDir(), "missing-driver.yaml"))
	if err != nil {
		t.Fatalf("loadDriverConfig() error = %v", err)
	}
	if cfg.DriverGRPCAddr != "115.191.16.159:50055" ||
		cfg.OrderGRPCAddr != "115.191.16.159:50051" ||
		cfg.DispatchGRPCAddr != "115.191.16.159:50056" ||
		cfg.LocationGRPCAddr != "115.191.16.159:9001" {
		t.Fatalf("unexpected default backend addresses: %+v", cfg)
	}
}

func TestCertificationFilesRouteServesLocalStoredImages(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DRIVER_CERT_LOCAL_DIR", dir)
	imagePath := filepath.Join(dir, "drivers", "25")
	if err := os.MkdirAll(imagePath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(imagePath, "id_card_front-1.png"), []byte("png-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "cert-file-test-key"})
	request := httptest.NewRequest(http.MethodGet, "/api/driver/v1/certification-files/drivers/25/id_card_front-1.png", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if response.Body.String() != "png-data" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestDriverAvatarUploadRequiresDriverToken(t *testing.T) {
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "avatar-route-auth-test-key"})
	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/drivers/avatar/upload", bytes.NewBufferString(`{
		"avatar": "data:image/png;base64,aGVsbG8="
	}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("avatar upload status = %d, want %d: %s", response.Code, http.StatusUnauthorized, response.Body.String())
	}
}

func TestDriverAvatarUploadStoresLocalImageAndReturnsURL(t *testing.T) {
	const signingKey = "avatar-upload-test-key"
	dir := t.TempDir()
	t.Setenv("DRIVER_AVATAR_LOCAL_DIR", dir)

	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: signingKey})
	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, signingKey)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/drivers/avatar/upload", bytes.NewBufferString(`{
		"avatar": "data:image/png;base64,aGVsbG8="
	}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("avatar upload status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			AvatarURL string `json:"avatarUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.AvatarURL == "" || !strings.HasPrefix(body.Data.AvatarURL, "/api/driver/v1/avatar-files/drivers/25/avatar-") || !strings.HasSuffix(body.Data.AvatarURL, ".png") {
		t.Fatalf("avatarUrl = %q, want local avatar file URL", body.Data.AvatarURL)
	}

	storedPath := filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(body.Data.AvatarURL, "/api/driver/v1/avatar-files/")))
	data, err := os.ReadFile(storedPath)
	if err != nil {
		t.Fatalf("avatar file was not stored: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("stored avatar data = %q, want %q", string(data), "hello")
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
		{http.MethodPost, "/api/driver/v1/orders/heatmap"},
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

func TestGetDriverRouteRejectsNonGetMethod(t *testing.T) {
	const signingKey = "driver-get-method-test-key"
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: signingKey})
	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, signingKey)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/drivers/get", bytes.NewBufferString(`{}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /drivers/get status = %d, want %d: %s", response.Code, http.StatusMethodNotAllowed, response.Body.String())
	}
}

func TestDriverHeatmapRouteRequiresDriverToken(t *testing.T) {
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "heatmap-route-test-key"})
	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/orders/heatmap", bytes.NewBufferString(`{
		"longitude": 116.397,
		"latitude": 39.908,
		"radiusMeters": 2000
	}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("heatmap status = %d, want %d: %s", response.Code, http.StatusUnauthorized, response.Body.String())
	}
}

func TestWalletSummaryRouteIsNotRegisteredWithoutWalletHandler(t *testing.T) {
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "wallet-route-test-key"})
	request := httptest.NewRequest(http.MethodGet, "/api/driver/v1/wallet/summary", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("wallet summary status = %d, want %d: %s", response.Code, http.StatusNotFound, response.Body.String())
	}
}

// TestDriverManagementRoutesAreNotExposedToDrivers 防止司机端管理接口泄漏：
// 创建/删除/按手机号查询司机属于 admin 端能力，不应在司机端 API 暴露，否则任意登录司机可越权操作其他司机。
func TestDriverManagementRoutesAreNotExposedToDrivers(t *testing.T) {
	handler := newHTTPHandler(&svc.ServiceContext{SigningKey: "mgmt-route-test-key"})
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/driver/v1/drivers"},
		{http.MethodGet, "/api/driver/v1/drivers/delete?id=25"},
		{http.MethodPost, "/api/driver/v1/drivers/by-phone"},
	}
	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			request := httptest.NewRequest(route.method, route.path, bytes.NewBufferString(`{}`))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != http.StatusNotFound {
				t.Fatalf("driver management route must not be exposed to driver app, status = %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestUpdateDriverRouteForcesSelfIdentityAndIgnoresStatus(t *testing.T) {
	client := &recordingUpdateDriverClient{}
	const signingKey = "update-driver-route-test-key"
	handler := newHTTPHandler(&svc.ServiceContext{
		SigningKey:   signingKey,
		DriverClient: client,
	})

	token, err := jwtx.SignAccountToken(jwtx.AccountTokenPayload{
		AccountID:     25,
		AccountType:   "driver",
		AccountStatus: 2,
		TTL:           time.Minute,
	}, signingKey)
	if err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/driver/v1/drivers/update", bytes.NewBufferString(`{
		"id": 99,
		"phone": "13800000002",
		"status": "DRIVER_STATUS_FROZEN"
	}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if client.updateDriverRequest == nil {
		t.Fatal("UpdateDriver() was not called")
	}
	if got := client.updateDriverRequest.GetId(); got != 25 {
		t.Fatalf("UpdateDriver() id = %d, want 25", got)
	}
	if got := client.updateDriverRequest.GetStatus(); got != driversproto.DriverStatus_DRIVER_STATUS_UNSPECIFIED {
		t.Fatalf("UpdateDriver() status = %v, want unspecified", got)
	}
	if got := client.updateDriverRequest.GetPhone(); got != "13800000002" {
		t.Fatalf("UpdateDriver() phone = %q, want 13800000002", got)
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

func TestDriverPushWebSocketSkipsRedisOrderWithoutPendingDispatch(t *testing.T) {
	const signingKey = "driver-ws-pending-dispatch-test-key"
	redisServer := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer rdb.Close()

	if err := rdb.SAdd(context.Background(), fmt.Sprintf(constants.RedisDriverAvailable, 25), "1001").Err(); err != nil {
		t.Fatal(err)
	}

	orderClient := &pollingOrderClient{getOrderResponse: &orderproto.GetOrderResponse{
		OrderId:             1001,
		OrderNo:             "NO-1001",
		FromAddress:         "from",
		ToAddress:           "to",
		Status:              orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
		EstimatedPriceCents: 29900,
		CreatedAt:           123,
	}}
	dispatchClient := &pollingDispatchClient{listResponse: &dispatchproto.ListDispatchRecordsResponse{
		List: []*dispatchproto.DispatchRecord{},
	}}
	server := httptest.NewServer(newHTTPHandler(&svc.ServiceContext{
		SigningKey:       signingKey,
		RedisClient:      rdb,
		OrderClient:      orderClient,
		DispatchClient:   dispatchClient,
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
	if ack.Type != "connected" || ack.Degraded {
		t.Fatalf("unexpected ws ack: %+v", ack)
	}

	_ = conn.SetReadDeadline(time.Now().Add(80 * time.Millisecond))
	var msg struct {
		Type    string `json:"type"`
		OrderID int64  `json:"orderId"`
	}
	if err := websocket.JSON.Receive(conn, &msg); err == nil {
		t.Fatalf("unexpected ws message without pending dispatch: %+v", msg)
	}

	if dispatchClient.listRequest == nil {
		t.Fatal("ListDispatchRecords() was not called")
	}
	if dispatchClient.listRequest.GetDriverId() != 25 || dispatchClient.listRequest.GetStatus() != constants.DispatchStatusPending {
		t.Fatalf("ListDispatchRecords() request = %+v, want current driver's pending dispatch records", dispatchClient.listRequest)
	}
}

type recordingUpdateDriverClient struct {
	updateDriverRequest *driversproto.UpdateDriverRequest
}

func (r *recordingUpdateDriverClient) CreateDriver(context.Context, *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) RegisterDriver(context.Context, *driversproto.CreateDriverRequest) (*driversproto.CreateDriverResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) UpdateDriver(_ context.Context, req *driversproto.UpdateDriverRequest) (*driversproto.UpdateDriverResponse, error) {
	r.updateDriverRequest = req
	return &driversproto.UpdateDriverResponse{
		Id:        req.GetId(),
		Status:    driversproto.DriverStatus_DRIVER_STATUS_NORMAL,
		UpdatedAt: 123,
	}, nil
}

func (r *recordingUpdateDriverClient) GetDriver(context.Context, *driversproto.GetDriverRequest) (*driversproto.GetDriverResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) GetDriverByPhone(context.Context, *driversproto.GetDriverByPhoneRequest) (*driversproto.GetDriverByPhoneResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) SetDriverOnline(context.Context, *driversproto.SetDriverOnlineRequest) (*driversproto.SetDriverOnlineResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) SetDriverOffline(context.Context, *driversproto.SetDriverOfflineRequest) (*driversproto.SetDriverOfflineResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) ReportLocation(context.Context, *driversproto.ReportLocationRequest) (*driversproto.ReportLocationResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) SetDriverServiceStatus(context.Context, *driversproto.SetDriverServiceStatusRequest) (*driversproto.SetDriverServiceStatusResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) Heartbeat(context.Context, *driversproto.HeartbeatRequest) (*driversproto.HeartbeatResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) DeleteDriver(context.Context, *driversproto.DeleteDriverRequest) (*driversproto.DeleteDriverResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) Login(context.Context, *driversproto.LoginRequest) (*driversproto.LoginResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) LoginBySMS(context.Context, *driversproto.LoginBySMSRequest) (*driversproto.LoginResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) CreateVehicle(context.Context, *driversproto.CreateVehicleRequest) (*driversproto.CreateVehicleResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) UpdateVehicle(context.Context, *driversproto.UpdateVehicleRequest) (*driversproto.UpdateVehicleResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) DeleteVehicle(context.Context, *driversproto.DeleteVehicleRequest) (*driversproto.DeleteVehicleResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) GetVehicle(context.Context, *driversproto.GetVehicleRequest) (*driversproto.GetVehicleResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) ListNearbyDrivers(context.Context, *driversproto.ListNearbyDriversRequest) (*driversproto.ListNearbyDriversResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) GetDriverAiScore(context.Context, *driversproto.GetDriverAiScoreRequest) (*driversproto.GetDriverAiScoreResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) UploadCertification(context.Context, *driversproto.UploadCertificationRequest) (*driversproto.UploadCertificationResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) GetCertification(context.Context, *driversproto.GetCertificationRequest) (*driversproto.GetCertificationResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) CreateWithdraw(context.Context, *driversproto.CreateWithdrawRequest) (*driversproto.CreateWithdrawResponse, error) {
	return nil, nil
}

func (r *recordingUpdateDriverClient) ListWithdraws(context.Context, *driversproto.ListWithdrawsRequest) (*driversproto.ListWithdrawsResponse, error) {
	return nil, nil
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

func TestAuthRateLimitAppliedToPublicEndpoints(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		SigningKey: "rate-limit-test-key",
		CodeCache:  svc.NewLocalCodeCache(time.Minute),
	}
	handler := newHTTPHandler(svcCtx)

	// send-sms-code：5 次/分钟/IP，第 6 次应 429。
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/driver/v1/auth/send-sms-code", bytes.NewBufferString(`{"phone":"13800138000"}`))
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code == http.StatusTooManyRequests {
			t.Fatalf("send-sms-code request %d unexpectedly limited", i+1)
		}
	}
	limited := httptest.NewRecorder()
	handler.ServeHTTP(limited, httptest.NewRequest(http.MethodPost, "/api/driver/v1/auth/send-sms-code", bytes.NewBufferString(`{"phone":"13800138000"}`)))
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("6th send-sms-code status = %d, want 429", limited.Code)
	}

	// login-by-password：10 次/分钟/IP，第 11 次应 429。
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/driver/v1/auth/login-by-password", bytes.NewBufferString(`{"phone":"13800138000","password":"wrong"}`))
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		if resp.Code == http.StatusTooManyRequests {
			t.Fatalf("login request %d unexpectedly limited", i+1)
		}
	}
	limited = httptest.NewRecorder()
	handler.ServeHTTP(limited, httptest.NewRequest(http.MethodPost, "/api/driver/v1/auth/login-by-password", bytes.NewBufferString(`{"phone":"13800138000","password":"wrong"}`)))
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("11th login status = %d, want 429", limited.Code)
	}
}
