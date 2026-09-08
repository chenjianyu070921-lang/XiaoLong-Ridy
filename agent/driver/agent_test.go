package driver

import (
	"context"
	"strings"
	"testing"
)

type stubModel struct {
	raw string
	err error
}

func (s stubModel) Generate(_ context.Context, system, user string) (string, error) {
	_ = system
	_ = user
	return s.raw, s.err
}

type stubKnowledge struct {
	text string
}

func (s *stubKnowledge) Retrieve(_ context.Context, _ OrderInput) (string, error) {
	return s.text, nil
}

func TestRateUsesValidLLMOutput(t *testing.T) {
	model := stubModel{raw: "好的，以下是结果：\n```json\n{\"star\":4,\"tags\":[\"接驾迟到\",\"自造标签\"],\"reason\":\"轻微迟到\"}\n```"}
	a := NewAgent(model, nil)
	got, err := a.Rate(context.Background(), OrderInput{OrderID: "o-1"})
	if err != nil {
		t.Fatalf("Rate err=%v", err)
	}
	if got.Star != 4 || got.Source != SourceLLM {
		t.Fatalf("got %+v", got)
	}
	if len(got.Tags) != 1 || got.Tags[0] != TagLatePickup {
		t.Fatalf("白名单外标签未过滤: %v", got.Tags)
	}
}

func TestRateFallsBackOnInvalidJSON(t *testing.T) {
	a := NewAgent(stubModel{raw: "抱歉我无法输出 JSON"}, nil)
	got, err := a.Rate(context.Background(), OrderInput{OrderID: "o-1"})
	if err != nil {
		t.Fatalf("Rate err=%v", err)
	}
	if got.Source != SourceRuleFallback || got.Star != 5 {
		t.Fatalf("got %+v", got)
	}
	if got.Tags == nil || len(got.Tags) != 0 {
		t.Fatalf("无标签时应返回空数组: %#v", got.Tags)
	}
}

func TestRateFallsBackOnStarOutOfRange(t *testing.T) {
	a := NewAgent(stubModel{raw: `{"star":9,"tags":["接驾准时"],"reason":"越界"}`}, nil)
	got, _ := a.Rate(context.Background(), OrderInput{OrderID: "o-1"})
	if got.Source != SourceRuleFallback {
		t.Fatalf("星级越界应降级: %+v", got)
	}
}

func TestReasonTruncatedTo100Runes(t *testing.T) {
	long := strings.Repeat("长", 150)
	a := NewAgent(stubModel{raw: `{"star":5,"tags":[],"reason":"` + long + `"}`}, nil)
	got, _ := a.Rate(context.Background(), OrderInput{OrderID: "o-1"})
	if n := len([]rune(got.Reason)); n != 100 {
		t.Fatalf("reason 应截断到 100 字, got %d", n)
	}
}

func TestFallbackCriticalInduceCancel(t *testing.T) {
	a := NewAgent(nil, nil) // 模型不可用，直接走规则引擎
	got, err := a.Rate(context.Background(), OrderInput{
		OrderID:   "o-2",
		Behaviors: []BehaviorEvent{{Name: BehaviorInduceCancel}},
	})
	if err != nil {
		t.Fatalf("Rate err=%v", err)
	}
	if got.Star != 1 || got.Source != SourceRuleFallback {
		t.Fatalf("got %+v", got)
	}
	if len(got.Tags) != 1 || got.Tags[0] != TagInduceCancel {
		t.Fatalf("tags=%v", got.Tags)
	}
}

func TestFallbackCommentPriority(t *testing.T) {
	a := NewAgent(stubModel{raw: ""}, nil)
	got, _ := a.Rate(context.Background(), OrderInput{
		OrderID:          "o-3",
		PassengerComment: "司机绕路了，而且态度很差",
	})
	if got.Star != 2 { // 绕路(3星档) + 态度差(2星档) → 取最严重 2 星
		t.Fatalf("star=%d", got.Star)
	}
	if len(got.Tags) != 2 {
		t.Fatalf("tags=%v", got.Tags)
	}
}

func TestFallbackFiveStarPositiveTags(t *testing.T) {
	a := NewAgent(nil, nil)
	got, _ := a.Rate(context.Background(), OrderInput{
		OrderID:   "o-4",
		Behaviors: []BehaviorEvent{{Name: BehaviorOnTime}, {Name: BehaviorCleanCar}},
	})
	if got.Star != 5 {
		t.Fatalf("star=%d", got.Star)
	}
	for _, want := range []string{TagOnTime, TagCleanCar} {
		found := false
		for _, tag := range got.Tags {
			if tag == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("缺正向标签 %s: %v", want, got.Tags)
		}
	}
}

func TestUnknownBehaviorIgnored(t *testing.T) {
	a := NewAgent(nil, nil)
	got, _ := a.Rate(context.Background(), OrderInput{
		OrderID:   "o-5",
		Behaviors: []BehaviorEvent{{Name: "不存在的行为"}},
	})
	if got.Star != 5 || len(got.Tags) != 0 {
		t.Fatalf("未知行为应被忽略: %+v", got)
	}
}

// 模型输出合法性：系统提示词包含花小猪规则、知识库片段被注入用户提示词。
func TestPromptsInjected(t *testing.T) {
	var gotSystem, gotUser string
	model := promptCapture{capture: func(system, user string) { gotSystem, gotUser = system, user }}
	a := NewAgent(model, &stubKnowledge{text: "乘客偏好安静，车内禁烟"})
	_, err := a.Rate(context.Background(), OrderInput{OrderID: "o-6"})
	if err != nil {
		t.Fatalf("Rate err=%v", err)
	}
	if !strings.Contains(gotSystem, "花小猪") {
		t.Fatal("system prompt 缺少花小猪评分规则")
	}
	if !strings.Contains(gotUser, "乘客偏好安静") {
		t.Fatal("知识库片段未注入用户提示词")
	}
}

type promptCapture struct {
	capture func(system, user string)
}

func (m promptCapture) Generate(_ context.Context, system, user string) (string, error) {
	m.capture(system, user)
	return `{"star":5,"tags":[],"reason":"ok"}`, nil
}

func TestRateRequiresOrderID(t *testing.T) {
	a := NewAgent(nil, nil)
	if _, err := a.Rate(context.Background(), OrderInput{}); err == nil {
		t.Fatal("缺少 order_id 应返回错误")
	}
}
