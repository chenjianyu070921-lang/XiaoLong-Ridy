package handler

import (
	"net/http"
	"strings"

	"XiaoLong-Ridy/api/admin/internal/types"
	adminclient "XiaoLong-Ridy/rpc/adminsvc/client/adminservice"
)

// registerAiRoutes 注册 AI 运营助手相关路由，全部经过 authRequired 鉴权。
func (r *Router) registerAiRoutes() {
	r.mux.HandleFunc("/admin/v1/ai-agent/ask", r.authRequired(r.handleAiAsk))
	r.mux.HandleFunc("/admin/v1/ai-agent/suggestions", r.authRequired(r.handleAiSuggestions))
	r.mux.HandleFunc("/admin/v1/ai-agent/history", r.authRequired(r.handleAiHistory))
	r.mux.HandleFunc("/admin/v1/ai-agent/feedback", r.authRequired(r.handleAiFeedback))
	r.mux.HandleFunc("/admin/v1/ai-agent/conversations/", r.authRequired(r.handleAiConversation))
}

// handleAiAsk 提交受限运营问答并返回结构化回答。
func (r *Router) handleAiAsk(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.AiAskRequest
	if err := decodeJSON(w, req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	resp, err := r.ctx.AdminSvc.AskAiAgent(req.Context(), &adminclient.AiAskRequest{
		Scene:          body.Scene,
		Question:       body.Question,
		ConversationId: body.ConversationID,
		StartTime:      body.StartTime,
		EndTime:        body.EndTime,
		DemoMode:       body.DemoMode,
	})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, toAiAnswerDTO(resp))
}

// handleAiSuggestions 读取当前管理员可见的三个快捷问题。
func (r *Router) handleAiSuggestions(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := r.ctx.AdminSvc.GetAiSuggestions(req.Context(), &adminclient.AiSuggestionsRequest{})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	items := make([]types.AiSuggestionDTO, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		items = append(items, types.AiSuggestionDTO{Scene: it.GetScene(), QuickPrompt: it.GetQuickPrompt()})
	}
	writeSuccess(w, map[string]any{"items": items})
}

// handleAiHistory 查询当前管理员自己的 AI 问答会话摘要。
func (r *Router) handleAiHistory(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	resp, err := r.ctx.AdminSvc.GetAiHistory(req.Context(), &adminclient.AiHistoryRequest{})
	if err != nil {
		r.writeBizError(w, err)
		return
	}
	items := make([]types.AiConversationSummaryDTO, 0, len(resp.GetItems()))
	for _, it := range resp.GetItems() {
		items = append(items, types.AiConversationSummaryDTO{
			ConversationID: it.GetConversationId(),
			Scene:          it.GetScene(),
			SourceMode:     it.GetSourceMode(),
			Summary:        it.GetSummary(),
			UpdatedAt:      it.GetUpdatedAt(),
		})
	}
	writeSuccess(w, map[string]any{"items": items})
}

// handleAiFeedback 记录回答是否有帮助。
func (r *Router) handleAiFeedback(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	var body types.AiFeedbackRequest
	if err := decodeJSON(w, req, &body); err != nil {
		writeError(w, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	if _, err := r.ctx.AdminSvc.AiFeedback(req.Context(), &adminclient.AiFeedbackRequest{
		ConversationId: body.ConversationID,
		TraceId:        body.TraceID,
		Helpful:        body.Helpful,
	}); err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// handleAiConversation 结束并清空指定 AI 会话（DELETE）。
func (r *Router) handleAiConversation(w http.ResponseWriter, req *http.Request) {
	id := strings.Trim(strings.TrimPrefix(req.URL.Path, "/admin/v1/ai-agent/conversations/"), "/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusBadRequest, 40001, "invalid conversation id")
		return
	}
	if req.Method != http.MethodDelete {
		writeMethodNotAllowed(w)
		return
	}
	if _, err := r.ctx.AdminSvc.DeleteAiConversation(req.Context(), &adminclient.AiConversationRequest{ConversationId: id}); err != nil {
		r.writeBizError(w, err)
		return
	}
	writeSuccess(w, types.CommonResponse{Message: "ok"})
}

// toAiAnswerDTO 将 adminsvc 响应转换为网关结构化回答。
func toAiAnswerDTO(resp *adminclient.AiAnswerResponse) types.AiAnswerDTO {
	if resp == nil {
		return types.AiAnswerDTO{}
	}
	out := types.AiAnswerDTO{
		Summary:        resp.GetSummary(),
		SourceMode:     resp.GetSourceMode(),
		Citations:      resp.GetCitations(),
		ConversationID: resp.GetConversationId(),
		TraceID:        resp.GetTraceId(),
	}
	for _, ev := range resp.GetEvidence() {
		out.Evidence = append(out.Evidence, types.AiEvidenceDTO{Label: ev.GetLabel(), Value: ev.GetValue(), Comparison: ev.GetComparison()})
	}
	for _, p := range resp.GetPriorities() {
		out.Priorities = append(out.Priorities, types.AiPriorityDTO{Type: p.GetType(), ID: p.GetId(), Level: p.GetLevel(), Reasons: p.GetReasons(), Route: p.GetRoute()})
	}
	for _, act := range resp.GetActions() {
		out.Actions = append(out.Actions, types.AiActionDTO{Type: act.GetType(), Label: act.GetLabel(), Route: act.GetRoute()})
	}
	return out
}
