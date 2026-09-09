package react

import (
	"context"
	"testing"
)

func TestAgentRunsCompleteToolCallingFlow(t *testing.T) {
	tools, err := NewDefaultToolRegistry()
	if err != nil {
		t.Fatal(err)
	}
	agent, err := NewAgent(context.Background(), ScriptedModel{}, tools, 3)
	if err != nil {
		t.Fatal(err)
	}
	state, err := agent.Run(context.Background(), "帮我查商品1001多少钱")
	if err != nil {
		t.Fatal(err)
	}
	if !state.Done || state.FinalAnswer == "" {
		t.Fatalf("agent did not finish: %+v", state)
	}
	if state.LoopCount != 2 {
		t.Fatalf("LoopCount = %d, want 2", state.LoopCount)
	}
	if len(state.Messages) != 4 {
		t.Fatalf("message count = %d, want user/assistant/tool/assistant", len(state.Messages))
	}
	if len(state.Observations) != 1 || state.Observations[0].Result != "商品1001价格为299元" {
		t.Fatalf("observations = %+v", state.Observations)
	}
	if len(state.Messages[1].ToolCalls) != 1 || state.Messages[1].ToolCalls[0].Function.Name != "get_product_price" {
		t.Fatalf("tool calls = %+v", state.Messages[1].ToolCalls)
	}
}

func TestAgentStopsAtMaximumLoopCount(t *testing.T) {
	tools, err := NewDefaultToolRegistry()
	if err != nil {
		t.Fatal(err)
	}
	agent, err := NewAgent(context.Background(), ScriptedModel{}, tools, 1)
	if err != nil {
		t.Fatal(err)
	}
	state, err := agent.Run(context.Background(), "帮我查商品1001多少钱")
	if err != nil {
		t.Fatal(err)
	}
	if !state.Done || state.LoopCount != 1 || len(state.Observations) != 0 {
		t.Fatalf("bounded state = %+v", state)
	}
}
