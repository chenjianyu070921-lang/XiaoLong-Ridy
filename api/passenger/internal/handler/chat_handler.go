package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"XiaoLong-Ridy/api/passenger/internal/logic"
	"XiaoLong-Ridy/api/passenger/internal/svc"
	"XiaoLong-Ridy/api/passenger/internal/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ChatConversationHandler GET /api/passenger/v1/chat/conversation?order_id=
// 返回会话信息（司机接单前返回未开启）；peer.name 由网关补全。
func ChatConversationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, "仅支持GET请求")
			return
		}
		orderID, err := strconv.ParseInt(r.URL.Query().Get("order_id"), 10, 64)
		if err != nil || orderID <= 0 {
			writeError(w, http.StatusBadRequest, codeInvalidRequest, "invalid order_id")
			return
		}
		resp, err := logic.NewChatLogic(r.Context(), svcCtx, bearerToken(r)).GetConversation(orderID)
		if err != nil {
			writeChatGRPCError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ChatMessagesHandler GET /api/passenger/v1/chat/messages?conversation_id=&cursor=&limit=
func ChatMessagesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, "仅支持GET请求")
			return
		}
		conversationID, err := strconv.ParseInt(r.URL.Query().Get("conversation_id"), 10, 64)
		if err != nil || conversationID <= 0 {
			writeError(w, http.StatusBadRequest, codeInvalidRequest, "invalid conversation_id")
			return
		}
		cursor, _ := strconv.ParseInt(r.URL.Query().Get("cursor"), 10, 64)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		resp, err := logic.NewChatLogic(r.Context(), svcCtx, bearerToken(r)).ListMessages(conversationID, cursor, int32(limit))
		if err != nil {
			writeChatGRPCError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ChatSendHandler POST /api/passenger/v1/chat/send
func ChatSendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, "仅支持POST请求")
			return
		}
		var req types.ChatSendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, codeInvalidRequest, "invalid body")
			return
		}
		if req.ConversationID <= 0 || (req.MsgType != types.ChatMsgTypeText && req.MsgType != types.ChatMsgTypeQuick) ||
			strings.TrimSpace(req.Content) == "" || strings.TrimSpace(req.ClientMsgID) == "" {
			writeError(w, http.StatusBadRequest, codeInvalidRequest, "missing fields")
			return
		}
		resp, err := logic.NewChatLogic(r.Context(), svcCtx, bearerToken(r)).SendMessage(req.ConversationID, req.MsgType, req.Content, req.ClientMsgID)
		if err != nil {
			writeChatGRPCError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ChatReadHandler POST /api/passenger/v1/chat/read
func ChatReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, "仅支持POST请求")
			return
		}
		var req types.ChatReadRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.ConversationID <= 0 {
			writeError(w, http.StatusBadRequest, codeInvalidRequest, "invalid conversationId")
			return
		}
		resp, err := logic.NewChatLogic(r.Context(), svcCtx, bearerToken(r)).MarkRead(req.ConversationID)
		if err != nil {
			writeChatGRPCError(w, err)
			return
		}
		writeSuccess(w, resp)
	}
}

// ChatPhrasesHandler GET /api/passenger/v1/chat/phrases 返回 MVP 快捷短语。
func ChatPhrasesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, "仅支持GET请求")
			return
		}
		resp := logic.NewChatLogic(r.Context(), svcCtx, bearerToken(r)).QuickPhrases()
		writeSuccess(w, resp)
	}
}

// writeChatGRPCError 将 chatsvc 返回的 gRPC 错误映射为乘客端业务错误码（含敏感词 46000）。
// 先处理本进程 logic 层 sentinel（如 JWT 缺失/失效），再处理下游 gRPC status，保证鉴权失败返回 401 而非 500。
func writeChatGRPCError(w http.ResponseWriter, err error) {
	if errors.Is(err, logic.ErrUnauthorized) {
		writeError(w, http.StatusUnauthorized, codeInvalidToken, "Token invalid")
		return
	}
	if errors.Is(err, logic.ErrForbidden) {
		writeError(w, http.StatusForbidden, codeForbidden, "forbidden")
		return
	}
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, codeInternalServer, "服务异常")
		return
	}
	switch st.Code() {
	case codes.PermissionDenied:
		writeError(w, http.StatusForbidden, codeForbidden, "无权访问该会话")
	case codes.NotFound:
		writeError(w, http.StatusNotFound, codeNotFound, "会话不存在")
	case codes.FailedPrecondition:
		writeError(w, http.StatusConflict, codeConflict, "会话已结束，不可发送")
	case codes.InvalidArgument:
		if strings.Contains(st.Message(), "敏感词") {
			writeError(w, http.StatusBadRequest, codeSensitiveWord, "命中敏感词，消息已拦截")
			return
		}
		writeError(w, http.StatusBadRequest, codeInvalidRequest, st.Message())
	default:
		writeError(w, http.StatusInternalServerError, codeInternalServer, "服务异常")
	}
}
