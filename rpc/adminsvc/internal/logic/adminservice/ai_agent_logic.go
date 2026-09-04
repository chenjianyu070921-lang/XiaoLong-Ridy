package adminservicelogic

import (
	"context"
	"errors"
	"strconv"

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
	admin, err := ValidateAdminTokenFromContext(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
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
	if err != nil {
		return nil, mapAiAgentErr(err)
	}
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
