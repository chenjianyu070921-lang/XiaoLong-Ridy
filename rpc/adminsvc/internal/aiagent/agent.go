package aiagent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	// ErrDemoDenied 表示生产环境拒绝演示模式。
	ErrDemoDenied = errors.New("demo mode disabled")
	// ErrSceneMismatch 表示会话场景与请求场景不一致。
	ErrSceneMismatch = errors.New("conversation scene mismatch")
	// ErrNotOwned 表示会话不属于当前管理员。
	ErrNotOwned = errors.New("conversation not owned by admin")
)

// Engine 是 aiagent 编排器，串起场景路由、事实查询、模型调用与降级。
type Engine struct {
	query QueryPort
	model ModelClient
	store SessionStore
	cfg   Config
}

// AskRequest 是一次受限问答请求。
type AskRequest struct {
	Scene          Scene
	Question       string
	ConversationID string
	StartTime      string
	EndTime        string
	DemoMode       bool
	AdminID        int64
}

// NewEngine 构造编排器；model 或 store 为 nil 时使用安全默认值。
func NewEngine(query QueryPort, model ModelClient, store SessionStore, cfg Config) *Engine {
	if model == nil {
		model = DisabledModel{}
	}
	if store == nil {
		store = NewInMemoryStore()
	}
	if cfg.Conversation.MaxRounds <= 0 {
		cfg.Conversation.MaxRounds = 6
	}
	if cfg.Conversation.TTLSeconds <= 0 {
		cfg.Conversation.TTLSeconds = 24 * 3600
	}
	return &Engine{query: query, model: model, store: store, cfg: cfg}
}

// Ask 处理一次受限问答，返回标准化回答。
func (e *Engine) Ask(ctx context.Context, req AskRequest) (*Answer, error) {
	if !req.Scene.Valid() {
		return nil, ErrOutOfScope
	}
	if req.DemoMode && !e.cfg.DemoEnabled {
		return nil, ErrDemoDenied
	}
	if strings.TrimSpace(req.Question) == "" {
		return nil, errors.New("question required")
	}

	conv, err := e.loadOrCreate(ctx, req)
	if err != nil {
		return nil, err
	}

	f, sourceMode, err := e.collectFacts(ctx, req)
	if err != nil {
		return nil, err
	}

	answer := e.generate(ctx, req.Scene, f, sourceMode)
	answer.ConversationID = conv.ID
	answer.TraceID = uuid.NewString()

	conv.Rounds = appendRound(conv.Rounds, Round{Question: req.Question, Answer: *answer}, e.cfg.Conversation.MaxRounds)
	conv.UpdatedAt = time.Now().Unix()
	_ = e.store.Save(ctx, conv)
	return answer, nil
}

// Suggestions 返回三个场景的快捷问题。
func (e *Engine) Suggestions() []Suggestion {
	return Suggestions()
}

// ListConversations 返回某管理员的会话摘要列表。
func (e *Engine) ListConversations(ctx context.Context, adminID int64) ([]ConversationSummary, error) {
	convs, err := e.store.ListByAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	out := make([]ConversationSummary, 0, len(convs))
	for _, c := range convs {
		sum := ConversationSummary{ConversationID: c.ID, Scene: c.Scene, UpdatedAt: c.UpdatedAt}
		if len(c.Rounds) > 0 {
			last := c.Rounds[len(c.Rounds)-1]
			sum.Summary = last.Answer.Summary
			sum.SourceMode = last.Answer.SourceMode
		}
		out = append(out, sum)
	}
	return out, nil
}

// DeleteConversation 删除某管理员的指定会话。
func (e *Engine) DeleteConversation(ctx context.Context, adminID int64, conversationID string) error {
	if conversationID == "" {
		return errors.New("conversation id required")
	}
	c, err := e.store.Load(ctx, conversationID)
	if err != nil {
		return err
	}
	if c == nil {
		return nil // 幂等：不存在视为已删除
	}
	if c.AdminID != adminID {
		return ErrNotOwned
	}
	return e.store.Delete(ctx, conversationID)
}

// Feedback 记录回答是否有帮助。V1 仅记录结构化日志，后续用于提示词与工具评估。
func (e *Engine) Feedback(adminID int64, conversationID, traceID string, helpful bool) {
	logx.Infow("ai feedback",
		logx.Field("admin_id", adminID),
		logx.Field("conversation_id", conversationID),
		logx.Field("trace_id", traceID),
		logx.Field("helpful", helpful))
}

// loadOrCreate 加载会话或创建新会话；校验会话归属与场景绑定。
func (e *Engine) loadOrCreate(ctx context.Context, req AskRequest) (*Conversation, error) {
	if req.ConversationID == "" {
		return &Conversation{ID: uuid.NewString(), AdminID: req.AdminID, Scene: req.Scene}, nil
	}
	c, err := e.store.Load(ctx, req.ConversationID)
	if err != nil || c == nil {
		// 会话不存在或已过期：沿用客户端 ID 重建，保持前端一致性。
		return &Conversation{ID: req.ConversationID, AdminID: req.AdminID, Scene: req.Scene}, nil
	}
	if c.Scene != req.Scene {
		return nil, ErrSceneMismatch
	}
	if c.AdminID != req.AdminID {
		return nil, ErrNotOwned
	}
	return c, nil
}

// collectFacts 读取结构化事实：演示走快照，否则走只读查询端口。
func (e *Engine) collectFacts(ctx context.Context, req AskRequest) (*facts, SourceMode, error) {
	if req.DemoMode {
		return demoFacts(req.Scene), SourceDemoSnapshot, nil
	}
	f, err := e.queryFacts(ctx, req)
	if err != nil {
		// 生产环境真实查询失败必须返回错误，禁止用快照伪造线上结论；
		// 仅在演示环境开启快照时允许降级到快照。
		if e.cfg.DemoEnabled {
			return demoFacts(req.Scene), SourceDemoSnapshot, nil
		}
		return nil, "", err
	}
	return f, SourceRealtime, nil
}

func (e *Engine) queryFacts(ctx context.Context, req AskRequest) (*facts, error) {
	switch req.Scene {
	case SceneOverview:
		m, err := e.query.QueryMetrics(ctx, req.StartTime, req.EndTime)
		if err != nil {
			return nil, err
		}
		return &facts{Metrics: m}, nil
	case SceneAbnormalOrder:
		orders, err := e.query.ListAbnormalOrders(ctx, req.StartTime, req.EndTime, 3)
		if err != nil {
			return nil, err
		}
		return &facts{Abnormal: orders}, nil
	case SceneRiskReview:
		risks, err := e.query.GetRiskSummary(ctx, 3)
		if err != nil {
			return nil, err
		}
		return &facts{Risks: risks}, nil
	default:
		return nil, ErrOutOfScope
	}
}

// generate 优先调用 LLM 组织回答，失败或输出无依据时降级到本地模板。
func (e *Engine) generate(ctx context.Context, scene Scene, f *facts, mode SourceMode) *Answer {
	gctx := ctx
	if e.cfg.Model.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		gctx, cancel = context.WithTimeout(ctx, time.Duration(e.cfg.Model.TimeoutSeconds)*time.Second)
		defer cancel()
	}
	if raw, err := e.model.Generate(gctx, systemPrompt, buildUserPrompt(scene, f)); err == nil {
		if a := validateLLMAnswer(scene, raw, f, mode); a != nil {
			return a
		}
	}
	a := buildTemplateAnswer(scene, f)
	if mode == SourceDemoSnapshot {
		a.SourceMode = SourceDemoSnapshot
	} else {
		a.SourceMode = SourceTemplateFallback
	}
	return a
}

// validateLLMAnswer 解析并校验模型输出：
// 优先对象的 ID 必须来自本次事实，无依据的对象被丢弃；关键字段缺失则返回 nil 触发降级。
func validateLLMAnswer(scene Scene, raw string, f *facts, mode SourceMode) *Answer {
	var a Answer
	if err := json.Unmarshal([]byte(trimJSON(raw)), &a); err != nil {
		return nil
	}
	if strings.TrimSpace(a.Summary) == "" {
		return nil
	}
	ids := f.priorityIDs()
	valid := a.Priorities[:0]
	for _, p := range a.Priorities {
		if ids[p.ID] {
			valid = append(valid, p)
		}
	}
	a.Priorities = valid
	a.SourceMode = mode
	return &a
}

// trimJSON 去除模型输出中可能的 markdown 代码块包裹与前后噪声。
func trimJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if i := strings.LastIndex(s, "}"); i >= 0 {
		s = s[:i+1]
	}
	return s
}
