package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"XiaoLong-Ridy/api/driver/internal/svc"
	"XiaoLong-Ridy/common/constants"
	"XiaoLong-Ridy/common/jwtx"
	driversproto "XiaoLong-Ridy/rpc/driversvc/proto"
	orderproto "XiaoLong-Ridy/rpc/ordersvc/proto"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

const (
	pushAuthTimeout  = 10 * time.Second
	pushPingEvery    = 25 * time.Second
	pushPollEvery    = 3 * time.Second
	pushPollPageSize = 20
)

type wsAuthMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type wsEnvelope struct {
	Type       string `json:"type"`
	DriverID   int64  `json:"driverId,omitempty"`
	Channel    string `json:"channel,omitempty"`
	Degraded   bool   `json:"degraded,omitempty"`
	Message    string `json:"message,omitempty"`
	ServerTime int64  `json:"serverTime,omitempty"`
}

type wsDispatchOrder struct {
	Type                string `json:"type"`
	OrderID             int64  `json:"orderId"`
	OrderNo             string `json:"orderNo"`
	FromAddress         string `json:"fromAddress"`
	ToAddress           string `json:"toAddress"`
	Status              int32  `json:"status"`
	EstimatedPriceCents int64  `json:"estimatedPriceCents"`
	CreatedAt           int64  `json:"createdAt"`
	ServerTime          int64  `json:"serverTime"`
}

func DriverPushWSHandler(svcCtx *svc.ServiceContext) http.Handler {
	server := websocket.Server{
		Handshake: func(*websocket.Config, *http.Request) error {
			return nil
		},
		Handler: func(conn *websocket.Conn) {
			defer conn.Close()
			serveDriverPushWS(conn, svcCtx)
		},
	}
	return server
}

func serveDriverPushWS(conn *websocket.Conn, svcCtx *svc.ServiceContext) {
	claims, err := authenticatePushConn(conn, svcCtx)
	if err != nil {
		_ = sendWSJSON(conn, wsEnvelope{Type: "auth_failed", Message: err.Error(), ServerTime: time.Now().Unix()})
		return
	}

	channel := fmt.Sprintf(constants.RedisDriverPush, claims.AccountID)
	degraded := svcCtx == nil || svcCtx.RedisClient == nil
	if degraded {
		if err := sendWSJSON(conn, wsEnvelope{
			Type:       "connected",
			DriverID:   int64(claims.AccountID),
			Channel:    channel,
			Degraded:   true,
			ServerTime: time.Now().Unix(),
		}); err != nil {
			return
		}
		serveDriverPushLoop(conn, svcCtx, int64(claims.AccountID), nil, true)
		return
	}

	pubsub := svcCtx.RedisClient.Subscribe(conn.Request().Context(), channel)
	defer pubsub.Close()
	if _, err := pubsub.ReceiveTimeout(conn.Request().Context(), 3*time.Second); err != nil {
		_ = sendWSJSON(conn, wsEnvelope{Type: "push_unavailable", Degraded: true, Message: err.Error(), ServerTime: time.Now().Unix()})
		keepWSAlive(conn)
		return
	}
	if err := sendWSJSON(conn, wsEnvelope{
		Type:       "connected",
		DriverID:   int64(claims.AccountID),
		Channel:    channel,
		Degraded:   false,
		ServerTime: time.Now().Unix(),
	}); err != nil {
		return
	}

	serveDriverPushLoop(conn, svcCtx, int64(claims.AccountID), pubsub.Channel(), false)
}

func serveDriverPushLoop(conn *websocket.Conn, svcCtx *svc.ServiceContext, driverID int64, redisCh <-chan *redis.Message, degraded bool) {
	done := make(chan struct{})
	go drainWSReads(conn, done)

	pingTicker := time.NewTicker(pushPingEvery)
	defer pingTicker.Stop()
	pollTicker := time.NewTicker(resolvePushPollInterval(svcCtx))
	defer pollTicker.Stop()
	seenOrders := make(map[int64]struct{})

	if !sendPolledDispatchOrders(conn, svcCtx, driverID, seenOrders) {
		return
	}
	for {
		select {
		case <-done:
			return
		case <-conn.Request().Context().Done():
			return
		case msg, ok := <-redisCh:
			if !ok {
				return
			}
			if err := websocket.Message.Send(conn, msg.Payload); err != nil {
				return
			}
		case <-pollTicker.C:
			if !sendPolledDispatchOrders(conn, svcCtx, driverID, seenOrders) {
				return
			}
		case <-pingTicker.C:
			if err := sendWSJSON(conn, wsEnvelope{Type: "ping", Degraded: degraded, ServerTime: time.Now().Unix()}); err != nil {
				return
			}
		}
	}
}

func sendPolledDispatchOrders(conn *websocket.Conn, svcCtx *svc.ServiceContext, driverID int64, seen map[int64]struct{}) bool {
	if svcCtx == nil || svcCtx.OrderClient == nil || driverID <= 0 {
		return true
	}
	pageSize := svcCtx.PushPollPageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = pushPollPageSize
	}
	resp, err := svcCtx.OrderClient.ListOrders(conn.Request().Context(), &orderproto.ListOrdersRequest{
		DriverId: driverID,
		Status:   orderproto.OrderStatus_ORDER_STATUS_WAIT_ACCEPT,
		Page:     1,
		PageSize: pageSize,
	})
	if err != nil {
		return sendWSJSON(conn, wsEnvelope{Type: "push_poll_error", Degraded: true, Message: err.Error(), ServerTime: time.Now().Unix()}) == nil
	}
	for _, order := range resp.GetList() {
		orderID := order.GetOrderId()
		if orderID <= 0 {
			continue
		}
		if _, ok := seen[orderID]; ok {
			continue
		}
		seen[orderID] = struct{}{}
		if err := sendWSJSON(conn, wsDispatchOrder{
			Type:                "dispatch_order",
			OrderID:             orderID,
			OrderNo:             order.GetOrderNo(),
			FromAddress:         order.GetFromAddress(),
			ToAddress:           order.GetToAddress(),
			Status:              int32(order.GetStatus()),
			EstimatedPriceCents: order.GetEstimatedPriceCents(),
			CreatedAt:           order.GetCreatedAt(),
			ServerTime:          time.Now().Unix(),
		}); err != nil {
			return false
		}
	}
	return true
}

func resolvePushPollInterval(svcCtx *svc.ServiceContext) time.Duration {
	if svcCtx != nil && svcCtx.PushPollInterval > 0 {
		return svcCtx.PushPollInterval
	}
	return pushPollEvery
}

func authenticatePushConn(conn *websocket.Conn, svcCtx *svc.ServiceContext) (*jwtx.AccountClaims, error) {
	if svcCtx == nil || strings.TrimSpace(svcCtx.SigningKey) == "" {
		return nil, fmt.Errorf("driver api signing key missing")
	}
	token := strings.TrimSpace(conn.Request().URL.Query().Get("token"))
	if token == "" {
		token = extractBearerToken(conn.Request().Header.Get("Authorization"))
	}
	if token == "" {
		_ = conn.SetReadDeadline(time.Now().Add(pushAuthTimeout))
		var msg wsAuthMessage
		if err := websocket.JSON.Receive(conn, &msg); err != nil {
			return nil, fmt.Errorf("missing auth token")
		}
		_ = conn.SetReadDeadline(time.Time{})
		if strings.EqualFold(msg.Type, "auth") {
			token = strings.TrimSpace(msg.Token)
		}
	}
	claims, err := jwtx.ParseAccountToken(token, svcCtx.SigningKey)
	if err == jwtx.ErrTokenExpired {
		return nil, fmt.Errorf("auth token expired")
	}
	if err != nil || claims.AccountType != "driver" {
		return nil, fmt.Errorf("auth token invalid")
	}
	if claims.AccountStatus != int(driversproto.DriverStatus_DRIVER_STATUS_NORMAL) {
		return nil, fmt.Errorf("driver account not active")
	}
	return claims, nil
}

func keepWSAlive(conn *websocket.Conn) {
	done := make(chan struct{})
	go drainWSReads(conn, done)

	ticker := time.NewTicker(pushPingEvery)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-conn.Request().Context().Done():
			return
		case <-ticker.C:
			if err := sendWSJSON(conn, wsEnvelope{Type: "ping", Degraded: true, ServerTime: time.Now().Unix()}); err != nil {
				return
			}
		}
	}
}

func drainWSReads(conn *websocket.Conn, done chan<- struct{}) {
	defer close(done)
	for {
		var raw string
		if err := websocket.Message.Receive(conn, &raw); err != nil {
			return
		}
	}
}

func sendWSJSON(conn *websocket.Conn, value any) error {
	return websocket.JSON.Send(conn, value)
}

func extractBearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
