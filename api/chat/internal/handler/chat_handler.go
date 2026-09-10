package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"XiaoLong-Ridy/api/chat/internal/middleware"
	"XiaoLong-Ridy/api/chat/internal/svc"
	chatproto "XiaoLong-Ridy/rpc/chatsvc/proto"
	"XiaoLong-Ridy/common/jwtx"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// callerOf 从 JWT claims 推导调用方类型(1司机/2乘客)与 caller_id。
func callerOf(c *jwtx.AccountClaims) (callerType int32, callerId int64) {
	if c.AccountType == "driver" {
		callerType = 1
	} else {
		callerType = 2
	}
	callerId = int64(c.AccountID)
	return
}

// roleOf 将调用方类型转为 proto 的 caller_role 字符串（driver/passenger）。
func roleOf(t int32) string {
	if t == 1 {
		return "driver"
	}
	return "passenger"
}

// GetConversationHandler GET /api/chat/v1/conversation?orderId=
func GetConversationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "未登录")
			return
		}
		orderID, err := strconv.ParseInt(r.URL.Query().Get("orderId"), 10, 64)
		if err != nil || orderID <= 0 {
			writeError(w, http.StatusBadRequest, 40000, "orderId 参数无效")
			return
		}
		ct, cid := callerOf(claims)
		resp, err := svcCtx.ChatClient.GetOrCreateConversation(r.Context(), &chatproto.GetOrCreateConversationRequest{
			OrderId:    orderID,
			CallerRole: roleOf(ct),
			CallerId:   cid,
		})
		if err != nil {
			handleGRPCError(w, err)
			return
		}
		data := map[string]interface{}{
			"conversationId": resp.ConversationId,
			"orderId":        resp.OrderId,
			"status":         resp.Status,
			"lastMsg":        resp.LastMsg,
			"lastMsgAt":      resp.LastMsgAt,
			"unread":         resp.Unread,
			"opened":         resp.Opened,
		}
		if resp.Peer != nil {
			data["peer"] = map[string]interface{}{"id": resp.Peer.Id, "name": resp.Peer.Name}
		} else {
			data["peer"] = nil
		}
		writeJSON(w, data)
	}
}

// ListMessagesHandler GET /api/chat/v1/messages?conversationId=&cursor=&limit=
func ListMessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "未登录")
			return
		}
		convID, err := strconv.ParseInt(r.URL.Query().Get("conversationId"), 10, 64)
		if err != nil || convID <= 0 {
			writeError(w, http.StatusBadRequest, 40000, "conversationId 参数无效")
			return
		}
		cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 {
			limit = 20
		}
		ct, cid := callerOf(claims)
		resp, err := svcCtx.ChatClient.ListMessages(r.Context(), &chatproto.ListMessagesRequest{
			ConversationId: convID,
			Cursor:         cursor,
			Limit:          int32(limit),
			CallerRole:     roleOf(ct),
			CallerId:       cid,
		})
		if err != nil {
			handleGRPCError(w, err)
			return
		}
		msgs := make([]map[string]interface{}, 0, len(resp.Messages))
		for _, m := range resp.Messages {
			msgs = append(msgs, map[string]interface{}{
				"id":             m.Id,
				"conversationId": m.ConversationId,
				"orderId":        m.OrderId,
				"senderType":     m.SenderType,
				"senderId":       m.SenderId,
				"msgType":        m.MsgType,
				"content":        m.Content,
				"clientMsgId":    m.ClientMsgId,
				"createAt":       m.CreateAt,
			})
		}
		writeJSON(w, map[string]interface{}{
			"messages":   msgs,
			"nextCursor": resp.NextCursor,
			"hasMore":    resp.HasMore,
		})
	}
}

// SendMessageHandler POST /api/chat/v1/send
func SendMessageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "未登录")
			return
		}
		var body struct {
			ConversationId int64  `json:"conversationId"`
			MsgType        int32  `json:"msgType"`
			Content        string `json:"content"`
			ClientMsgId    string `json:"clientMsgId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ConversationId <= 0 || body.ClientMsgId == "" {
			writeError(w, http.StatusBadRequest, 40000, "请求参数无效")
			return
		}
		ct, cid := callerOf(claims)
		resp, err := svcCtx.ChatClient.SendMessage(r.Context(), &chatproto.SendMessageRequest{
			ConversationId: body.ConversationId,
			SenderType:     ct,
			SenderId:       cid,
			MsgType:        body.MsgType,
			Content:        body.Content,
			ClientMsgId:    body.ClientMsgId,
		})
		if err != nil {
			handleGRPCError(w, err)
			return
		}
		writeJSON(w, map[string]interface{}{"messageId": resp.MessageId, "createAt": resp.CreateAt})
	}
}

// MarkReadHandler POST /api/chat/v1/read
func MarkReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := middleware.ClaimsFromContext(r.Context())
		if claims == nil {
			writeError(w, http.StatusUnauthorized, 40102, "未登录")
			return
		}
		var body struct {
			ConversationId int64 `json:"conversationId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ConversationId <= 0 {
			writeError(w, http.StatusBadRequest, 40000, "请求参数无效")
			return
		}
		ct, cid := callerOf(claims)
		resp, err := svcCtx.ChatClient.MarkRead(r.Context(), &chatproto.MarkReadRequest{
			ConversationId: body.ConversationId,
			CallerRole:     roleOf(ct),
			CallerId:       cid,
		})
		if err != nil {
			handleGRPCError(w, err)
			return
		}
		writeJSON(w, map[string]interface{}{"ok": resp.Ok})
	}
}

// ---- 响应工具 ----

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":      0,
		"message":   "success",
		"data":      data,
		"timestamp": 0,
		"traceId":   "",
	})
}

func writeError(w http.ResponseWriter, httpStatus int, code int32, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":      code,
		"message":  message,
		"data":     nil,
		"timestamp": 0,
		"traceId":  "",
	})
}

func handleGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, 50000, "服务内部错误")
		return
	}
	switch st.Code() {
	case codes.NotFound:
		writeError(w, http.StatusNotFound, 40400, st.Message())
	case codes.PermissionDenied:
		writeError(w, http.StatusForbidden, 40301, st.Message())
	case codes.FailedPrecondition:
		writeError(w, http.StatusConflict, 40900, st.Message())
	case codes.InvalidArgument:
		writeError(w, http.StatusBadRequest, 40000, st.Message())
	case codes.Unavailable:
		// 下游（ordersvc/chatsvc）瞬时不可达：返回 503 + 友好提示，避免裸 500 让前端误判为崩溃
		writeError(w, http.StatusServiceUnavailable, 50300, "下游服务暂不可用，请稍后重试")
	default:
		writeError(w, http.StatusInternalServerError, 50000, st.Message())
	}
}
