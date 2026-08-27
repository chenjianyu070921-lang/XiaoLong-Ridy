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

func TestLoadDriverConfigReadsYamlAndEnvCanOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "driver.yaml")
	if err := os.WriteFile(path, []byte(`
httpAddr: ":18082"
driverGrpcAddr: "driversvc:5055"
orderGrpcAddr: "ordersvc:50051"
dispatchGrpcAddr: "dispatchsvc:8083"
locationGrpcAddr: "locationsvc:5056"
redisAddr: "redis:6379"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadDriverConfig(path)
	if err != nil {
		t.Fatalf("loadDriverConfig() error = %v", err)
	}
	if cfg.HTTPAddr != ":18082" || cfg.DriverGRPCAddr != "driversvc:5055" ||
		cfg.OrderGRPCAddr != "ordersvc:50051" || cfg.DispatchGRPCAddr != "dispatchsvc:8083" ||
		cfg.LocationGRPCAddr != "locationsvc:5056" || cfg.RedisAddr != "redis:6379" {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	t.Setenv("ORDER_GRPC_ADDR", "ordersvc-prod:50051")
	if got := envOr("ORDER_GRPC_ADDR", cfg.OrderGRPCAddr); got != "ordersvc-prod:50051" {
		t.Fatalf("env override = %q", got)
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
