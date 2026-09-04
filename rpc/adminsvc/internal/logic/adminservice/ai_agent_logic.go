package adminservicelogic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"XiaoLong-Ridy/rpc/adminsvc/adminsvc"
	"XiaoLong-Ridy/rpc/adminsvc/internal/aiagent"
	"XiaoLong-Ridy/rpc/adminsvc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AiAgentLogic 处理 AI 运营助手相关 RPC。
type AiAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewAiAgentLogic 创建 AI 运营助手逻辑对象。
func NewAiAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AiAgentLogic {
	return &AiAgentLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// engine 构造绑定本次查询端口的 AI 编排引擎；会话存储由 svcCtx 共享，保证多轮延续。
func (l *AiAgentLogic) engine() *aiagent.Engine {
	return aiagent.NewEngine(adminQueryPort{l.svcCtx}, l.svcCtx.AiAgentModel, l.svcCtx.AiAgentStore, l.svcCtx.AiAgentCfg)
}

// Ask 处理受限运营问答。
func (l *AiAgentLogic) Ask(in *adminsvc.AiAskRequest) (*adminsvc.AiAnswerResponse, error) {
	// AI 入口显式接入独立只读权限 ai:insight:view，仅对阅读型角色（超管/运营/客服）开放。
	if err := requireAdminRoles(l.ctx, l.svcCtx, 1, 2, 3); err != nil {
		return nil, err
	}
	admin, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	engine := l.engine()
	answer, err := engine.Ask(l.ctx, aiagent.AskRequest{
		Scene:          aiagent.Scene(in.GetScene()),
		Question:       in.GetQuestion(),
		ConversationID: in.GetConversationId(),
		StartTime:      in.GetStartTime(),
		EndTime:        in.GetEndTime(),
		DemoMode:       in.GetDemoMode(),
		AdminID:        admin.ID,
	})
	durationMs := time.Since(start).Milliseconds()
	if err != nil {
		return nil, mapAiAgentErr(err)
	}
	// 落库审计：仅保留脱敏摘要（问题哈希 + 结果摘要），不落原始问题全文，满足设计文档 ai:insight:view 可追溯要求。
	l.writeAiAskAudit(admin.ID, in, answer, durationMs)
	return toAiAnswerResponse(answer), nil
}

// Suggestions 返回三个快捷问题。
func (l *AiAgentLogic) Suggestions(in *adminsvc.AiSuggestionsRequest) (*adminsvc.AiSuggestionsResponse, error) {
	if _, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx); err != nil {
		return nil, err
	}
	engine := l.engine()
	items := engine.Suggestions()
	out := make([]*adminsvc.AiSuggestion, 0, len(items))
	for _, s := range items {
		out = append(out, &adminsvc.AiSuggestion{Scene: string(s.Scene), QuickPrompt: s.QuickPrompt})
	}
	return &adminsvc.AiSuggestionsResponse{Items: out}, nil
}

// History 返回当前管理员的会话摘要。
func (l *AiAgentLogic) History(in *adminsvc.AiHistoryRequest) (*adminsvc.AiHistoryResponse, error) {
	admin, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	engine := l.engine()
	sums, err := engine.ListConversations(l.ctx, admin.ID)
	if err != nil {
		return nil, err
	}
	out := make([]*adminsvc.AiConversationSummary, 0, len(sums))
	for _, s := range sums {
		out = append(out, &adminsvc.AiConversationSummary{
			ConversationId: s.ConversationID,
			Scene:          string(s.Scene),
			SourceMode:     string(s.SourceMode),
			Summary:        s.Summary,
			UpdatedAt:      strconv.FormatInt(s.UpdatedAt, 10),
		})
	}
	return &adminsvc.AiHistoryResponse{Items: out}, nil
}

// Feedback 记录回答是否有帮助。
func (l *AiAgentLogic) Feedback(in *adminsvc.AiFeedbackRequest) (*adminsvc.CommonResponse, error) {
	admin, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	engine := l.engine()
	engine.Feedback(admin.ID, in.GetConversationId(), in.GetTraceId(), in.GetHelpful())
	l.writeAiFeedbackAudit(admin.ID, in)
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// DeleteConversation 结束并清空指定会话。
func (l *AiAgentLogic) DeleteConversation(in *adminsvc.AiConversationRequest) (*adminsvc.CommonResponse, error) {
	admin, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	engine := l.engine()
	if err := engine.DeleteConversation(l.ctx, admin.ID, in.GetConversationId()); err != nil {
		return nil, mapAiAgentErr(err)
	}
	l.writeAiDeleteAudit(admin.ID, in.GetConversationId())
	return &adminsvc.CommonResponse{Message: "ok"}, nil
}

// mapAiAgentErr 将 aiagent 内部错误映射为 gRPC status。
func mapAiAgentErr(err error) error {
	switch {
	case errors.Is(err, aiagent.ErrOutOfScope):
		return status.Error(codes.InvalidArgument, "问题超出 AI 助手支持范围")
	case errors.Is(err, aiagent.ErrDemoDenied):
		return status.Error(codes.FailedPrecondition, "演示模式未开启")
	case errors.Is(err, aiagent.ErrSceneMismatch):
		return status.Error(codes.InvalidArgument, "会话场景不匹配")
	case errors.Is(err, aiagent.ErrNotOwned):
		return status.Error(codes.PermissionDenied, "无权访问该会话")
	default:
		return status.Error(codes.Internal, "AI 助手处理失败")
	}
}

// toAiAnswerResponse 将 aiagent.Answer 转换为 proto 响应。
func toAiAnswerResponse(a *aiagent.Answer) *adminsvc.AiAnswerResponse {
	if a == nil {
		return &adminsvc.AiAnswerResponse{}
	}
	resp := &adminsvc.AiAnswerResponse{
		Summary:        a.Summary,
		SourceMode:     string(a.SourceMode),
		Citations:      a.Citations,
		ConversationId: a.ConversationID,
		TraceId:        a.TraceID,
	}
	for _, ev := range a.Evidence {
		resp.Evidence = append(resp.Evidence, &adminsvc.AiEvidence{Label: ev.Label, Value: ev.Value, Comparison: ev.Comparison})
	}
	for _, p := range a.Priorities {
		resp.Priorities = append(resp.Priorities, &adminsvc.AiPriority{Type: p.Type, Id: p.ID, Level: p.Level, Reasons: p.Reasons, Route: p.Route})
	}
	for _, act := range a.Actions {
		resp.Actions = append(resp.Actions, &adminsvc.AiAction{Type: act.Type, Label: act.Label, Route: act.Route})
	}
	return resp
}

// aiAuditTarget 统一 AI 审计记录的归属对象类型。
const aiAuditTarget = "ai_conversation"

// writeAiAudit 写一条 AI 模块审计日志。审计失败只记录日志，不影响业务响应。
func (l *AiAgentLogic) writeAiAudit(adminID int64, action, detail string) {
	if l.svcCtx == nil || l.svcCtx.MySQL == nil {
		return
	}
	if err := writeAuditAfterCommitted(l.ctx, l.svcCtx, adminID, "ai", action, aiAuditTarget, 0, detail, ""); err != nil {
		logx.Errorf("ai audit failed: action=%s admin_id=%d err=%v", action, adminID, err)
	}
}

// hashQuestion 计算问题摘要哈希，避免审计落库原始问题全文。
func hashQuestion(q string) string {
	sum := sha256.Sum256([]byte(q))
	return hex.EncodeToString(sum[:])[:16]
}

// aiToolName 返回场景对应的白名单只读工具名，用于审计记录。
func aiToolName(scene aiagent.Scene) string {
	switch scene {
	case aiagent.SceneOverview:
		return "metrics_query"
	case aiagent.SceneAbnormalOrder:
		return "abnormal_order_query"
	case aiagent.SceneRiskReview:
		return "risk_summary_query"
	default:
		return "unknown"
	}
}

// aiModelOK 根据来源模式推断模型调用是否成功及错误类别。
func aiModelOK(mode aiagent.SourceMode) (bool, string) {
	switch mode {
	case aiagent.SourceRealtime:
		return true, "ok"
	case aiagent.SourceTemplateFallback:
		return false, "model_fallback"
	case aiagent.SourceDemoSnapshot:
		return false, "demo_snapshot"
	default:
		return false, "unknown"
	}
}

// writeAiAskAudit 记录一次受限问答的脱敏审计摘要。
func (l *AiAgentLogic) writeAiAskAudit(adminID int64, in *adminsvc.AiAskRequest, answer *aiagent.Answer, durationMs int64) {
	modelOK, errCat := aiModelOK(answer.SourceMode)
	detail := map[string]any{
		"question_hash":   hashQuestion(in.GetQuestion()),
		"scene":           in.GetScene(),
		"tools":           aiToolName(aiagent.Scene(in.GetScene())),
		"source_mode":     string(answer.SourceMode),
		"summary":         truncateAuditText(answer.Summary, 500),
		"conversation_id": answer.ConversationID,
		"trace_id":        answer.TraceID,
		"duration_ms":     durationMs,
		"model_ok":        modelOK,
		"error_category":  errCat,
	}
	if b, err := json.Marshal(detail); err == nil {
		l.writeAiAudit(adminID, "ask", string(b))
	}
}

// writeAiFeedbackAudit 记录用户对回答的反馈。
func (l *AiAgentLogic) writeAiFeedbackAudit(adminID int64, in *adminsvc.AiFeedbackRequest) {
	detail := map[string]any{
		"conversation_id": in.GetConversationId(),
		"trace_id":        in.GetTraceId(),
		"helpful":         in.GetHelpful(),
	}
	if b, err := json.Marshal(detail); err == nil {
		l.writeAiAudit(adminID, "feedback", string(b))
	}
}

// writeAiDeleteAudit 记录清空会话动作。
func (l *AiAgentLogic) writeAiDeleteAudit(adminID int64, conversationID string) {
	detail := map[string]any{"conversation_id": conversationID}
	if b, err := json.Marshal(detail); err == nil {
		l.writeAiAudit(adminID, "delete_conversation", string(b))
	}
}

// truncateAuditText 将审计文本截断到指定长度，避免超过 admin_operation_log.detail 字段上限。
func truncateAuditText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
