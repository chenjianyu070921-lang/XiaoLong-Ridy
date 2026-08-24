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

	"golang.org/x/net/websocket"
)

const (
	pushAuthTimeout = 10 * time.Second
	pushPingEvery   = 25 * time.Second
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
		keepWSAlive(conn)
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

	done := make(chan struct{})
	go drainWSReads(conn, done)

	ticker := time.NewTicker(pushPingEvery)
	defer ticker.Stop()
	ch := pubsub.Channel()
	for {
		select {
		case <-done:
			return
		case <-conn.Request().Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if err := websocket.Message.Send(conn, msg.Payload); err != nil {
				return
			}
		case <-ticker.C:
			if err := sendWSJSON(conn, wsEnvelope{Type: "ping", ServerTime: time.Now().Unix()}); err != nil {
				return
			}
		}
	}
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
