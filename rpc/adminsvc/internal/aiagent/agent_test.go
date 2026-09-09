package aiagent

import (
	"context"
	"errors"
	"testing"
)

// fakeQueryPort 是 QueryPort 的固定事实实现。
type fakeQueryPort struct {
	metrics  *MetricsFact
	abnormal []AbnormalOrderFact
	risks    []RiskHitFact
	err      error
}

func (f fakeQueryPort) QueryMetrics(context.Context, string, string) (*MetricsFact, error) {
	return f.metrics, f.err
}
func (f fakeQueryPort) ListAbnormalOrders(context.Context, string, string, int) ([]AbnormalOrderFact, error) {
	return f.abnormal, f.err
}
func (f fakeQueryPort) GetRiskSummary(context.Context, int) ([]RiskHitFact, error) {
	return f.risks, f.err
}

// fakeModel 是 ModelClient 的可控实现。
type fakeModel struct {
	output string
	err    error
}

func (f fakeModel) Generate(context.Context, string, string) (string, error) { return f.output, f.err }

func sampleQuery() fakeQueryPort {
	return fakeQueryPort{
		metrics: &MetricsFact{OrderCount: 100, CompletedCount: 80, AbnormalCount: 5, CancelRate: "20%", CompletionRate: "80%", TimeoutCount: 3},
		abnormal: []AbnormalOrderFact{
			{OrderNo: "A001", AbnormalType: "dispatch"},
			{OrderNo: "A002", AbnormalType: "payment"},
		},
		risks: []RiskHitFact{
			{ID: 1, RiskLevel: 3, HitReason: "多设备登录", HandleStatus: "pending"},
			{ID: 2, RiskLevel: 2, HitReason: "高频取消", HandleStatus: "pending"},
		},
	}
}

func newTestEngine(q QueryPort, m ModelClient, cfg Config) *Engine {
	if q == nil {
		q = sampleQuery()
	}
	if m == nil {
		m = DisabledModel{}
	}
	return NewEngine(q, m, NewInMemoryStore(), cfg)
}

func TestAsk_OutOfScope(t *testing.T) {
	e := newTestEngine(nil, nil, Config{})
	if _, err := e.Ask(context.Background(), AskRequest{Scene: Scene("chat"), Question: "hi", AdminID: 1}); !errors.Is(err, ErrOutOfScope) {
		t.Fatalf("want ErrOutOfScope, got %v", err)
	}
}

func TestAsk_DemoDenied(t *testing.T) {
	e := newTestEngine(nil, nil, Config{DemoEnabled: false})
	if _, err := e.Ask(context.Background(), AskRequest{Scene: SceneOverview, Question: "q", DemoMode: true, AdminID: 1}); !errors.Is(err, ErrDemoDenied) {
		t.Fatalf("want errDemoDenied, got %v", err)
	}
}

func TestAsk_OverviewTemplateFallback(t *testing.T) {
	e := newTestEngine(nil, nil, Config{})
	a, err := e.Ask(context.Background(), AskRequest{Scene: SceneOverview, Question: "今日风险", StartTime: "2026-09-01 00:00:00", EndTime: "2026-09-01 23:59:59", AdminID: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if a.SourceMode != SourceTemplateFallback {
		t.Fatalf("want template_fallback, got %s", a.SourceMode)
	}
	if a.Summary == "" || len(a.Evidence) == 0 {
		t.Fatalf("template answer incomplete: %+v", a)
	}
	if a.ConversationID == "" || a.TraceID == "" {
		t.Fatalf("missing ids: %+v", a)
	}
}

func TestAsk_DemoSnapshot(t *testing.T) {
	e := newTestEngine(nil, nil, Config{DemoEnabled: true})
	a, err := e.Ask(context.Background(), AskRequest{Scene: SceneAbnormalOrder, Question: "识别异常订单", DemoMode: true, AdminID: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if a.SourceMode != SourceDemoSnapshot {
		t.Fatalf("want demo_snapshot, got %s", a.SourceMode)
	}
	if len(a.Priorities) == 0 {
		t.Fatalf("demo snapshot should have priorities")
	}
}

func TestAsk_AbnormalPriorities(t *testing.T) {
	e := newTestEngine(nil, nil, Config{})
	a, err := e.Ask(context.Background(), AskRequest{Scene: SceneAbnormalOrder, Question: "异常订单", AdminID: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(a.Priorities) == 0 {
		t.Fatalf("want priorities, got none")
	}
	for _, p := range a.Priorities {
		if p.ID != "A001" && p.ID != "A002" {
			t.Fatalf("priority id not from facts: %s", p.ID)
		}
	}
}

func TestAsk_QueryFailureFallsBackToDemoWhenEnabled(t *testing.T) {
	q := fakeQueryPort{err: errors.New("downstream down")}
	e := newTestEngine(q, nil, Config{DemoEnabled: true})
	a, err := e.Ask(context.Background(), AskRequest{Scene: SceneOverview, Question: "q", AdminID: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if a.SourceMode != SourceDemoSnapshot {
		t.Fatalf("want demo_snapshot on query failure, got %s", a.SourceMode)
	}
}

func TestAsk_QueryFailureReturnsErrorWhenDemoDisabled(t *testing.T) {
	q := fakeQueryPort{err: errors.New("downstream down")}
	e := newTestEngine(q, nil, Config{DemoEnabled: false})
	if _, err := e.Ask(context.Background(), AskRequest{Scene: SceneOverview, Question: "q", AdminID: 1}); err == nil {
		t.Fatal("want error on query failure in production, got nil")
	}
}

func TestValidateLLMAnswer_FiltersUnknownID(t *testing.T) {
	f := &facts{Abnormal: []AbnormalOrderFact{{OrderNo: "A001", AbnormalType: "dispatch"}}}
	raw := `{"summary":"ok","priorities":[{"type":"order","id":"A001","level":"high","reasons":["x"],"route":"/orders/abnormal"},{"type":"order","id":"FAKE","level":"high","reasons":["y"],"route":"/orders/abnormal"}]}`
	a := validateLLMAnswer(SceneAbnormalOrder, raw, f, SourceRealtime)
	if a == nil {
		t.Fatal("want valid answer")
	}
	if len(a.Priorities) != 1 || a.Priorities[0].ID != "A001" {
		t.Fatalf("want only A001 retained, got %+v", a.Priorities)
	}
}

func TestValidateLLMAnswer_InvalidJSON(t *testing.T) {
	f := &facts{}
	if a := validateLLMAnswer(SceneOverview, "not json", f, SourceRealtime); a != nil {
		t.Fatalf("want nil on invalid json, got %+v", a)
	}
}

func TestValidateLLMAnswer_EmptySummary(t *testing.T) {
	f := &facts{}
	if a := validateLLMAnswer(SceneOverview, `{"summary":""}`, f, SourceRealtime); a != nil {
		t.Fatalf("want nil on empty summary, got %+v", a)
	}
}

func TestAsk_SceneMismatch(t *testing.T) {
	store := NewInMemoryStore()
	e := NewEngine(sampleQuery(), DisabledModel{}, store, Config{})
	a1, _ := e.Ask(context.Background(), AskRequest{Scene: SceneOverview, Question: "q", AdminID: 1})
	// 用 overview 的会话 ID 请求 abnormal_order 场景，应拒绝。
	if _, err := e.Ask(context.Background(), AskRequest{Scene: SceneAbnormalOrder, Question: "q", ConversationID: a1.ConversationID, AdminID: 1}); !errors.Is(err, ErrSceneMismatch) {
		t.Fatalf("want errSceneMismatch, got %v", err)
	}
}

func TestAsk_NotOwned(t *testing.T) {
	store := NewInMemoryStore()
	e := NewEngine(sampleQuery(), DisabledModel{}, store, Config{})
	a1, _ := e.Ask(context.Background(), AskRequest{Scene: SceneOverview, Question: "q", AdminID: 1})
	if _, err := e.Ask(context.Background(), AskRequest{Scene: SceneOverview, Question: "q", ConversationID: a1.ConversationID, AdminID: 2}); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("want errNotOwned, got %v", err)
	}
}

func TestAppendRound_Truncate(t *testing.T) {
	var rounds []Round
	for i := 0; i < 10; i++ {
		rounds = appendRound(rounds, Round{Question: "q"}, 6)
	}
	if len(rounds) != 6 {
		t.Fatalf("want 6 rounds, got %d", len(rounds))
	}
}
