package react

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ThinkModel is the replaceable LLM boundary for this minimal skeleton.
// Tools are passed on every thinking turn so a real chat model can bind them.
type ThinkModel interface {
	Think(context.Context, []*schema.Message, []*schema.ToolInfo) (*schema.Message, error)
}

// ScriptedModel makes the complete flow deterministic without an external model.
// Replace it with an Eino ChatModel adapter when a provider is configured.
type ScriptedModel struct{}

func (ScriptedModel) Think(_ context.Context, messages []*schema.Message, tools []*schema.ToolInfo) (*schema.Message, error) {
	if len(tools) == 0 {
		return nil, fmt.Errorf("no tools registered")
	}
	for _, message := range messages {
		if message.Role == schema.Tool && message.Content != "" {
			return schema.AssistantMessage("根据工具查询结果，商品1001的价格是299元。", nil), nil
		}
	}
	return schema.AssistantMessage("", []schema.ToolCall{{
		ID:   "call-product-price-1",
		Type: "function",
		Function: schema.FunctionCall{
			Name:      "get_product_price",
			Arguments: `{"product_id":"1001"}`,
		},
	}}), nil
}

// Agent is a reusable, bounded ReAct runner.
type Agent struct {
	model    ThinkModel
	tools    *ToolRegistry
	maxLoops int
	runnable compose.Runnable[*AgentState, *AgentState]
}

func NewAgent(ctx context.Context, model ThinkModel, tools *ToolRegistry, maxLoops int) (*Agent, error) {
	if model == nil || tools == nil {
		return nil, fmt.Errorf("model and tools are required")
	}
	if maxLoops < 1 {
		maxLoops = 1
	}

	graph := compose.NewGraph[*AgentState, *AgentState]()
	if err := graph.AddLambdaNode("init", compose.InvokableLambda(func(_ context.Context, state *AgentState) (*AgentState, error) {
		if state == nil {
			return nil, fmt.Errorf("agent state is nil")
		}
		if state.MaxLoops < 1 {
			state.MaxLoops = maxLoops
		}
		if len(state.Messages) == 0 {
			state.Messages = append(state.Messages, schema.UserMessage(state.UserQuestion))
		}
		return state, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("llm_think", compose.InvokableLambda(func(ctx context.Context, state *AgentState) (*AgentState, error) {
		state.LoopCount++
		if state.LoopCount > state.MaxLoops {
			state.Done = true
			state.FinalAnswer = "已达到最大思考轮次，停止继续调用工具。"
			return state, nil
		}
		message, err := model.Think(ctx, state.Messages, tools.Infos())
		if err != nil {
			return nil, err
		}
		state.Messages = append(state.Messages, message)
		state.Thought = message.Content
		state.ToolCalls = append([]schema.ToolCall(nil), message.ToolCalls...)
		if len(state.ToolCalls) == 0 {
			state.Done = true
			state.FinalAnswer = message.Content
		} else if state.LoopCount >= state.MaxLoops {
			state.Done = true
			state.FinalAnswer = "已达到最大思考轮次，未执行更多工具调用。"
		}
		return state, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("tools", compose.InvokableLambda(func(ctx context.Context, state *AgentState) (*AgentState, error) {
		for _, call := range state.ToolCalls {
			result, err := tools.Execute(ctx, call)
			if err != nil {
				return nil, err
			}
			state.Observations = append(state.Observations, Observation{ToolCallID: call.ID, ToolName: call.Function.Name, Result: result})
			state.Messages = append(state.Messages, schema.ToolMessage(result, call.ID))
		}
		state.ToolCalls = nil
		return state, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, "init"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("init", "llm_think"); err != nil {
		return nil, err
	}
	if err := graph.AddBranch("llm_think", compose.NewGraphBranch(func(_ context.Context, state *AgentState) (string, error) {
		if state.Done || state.LoopCount >= state.MaxLoops {
			return compose.END, nil
		}
		return "tools", nil
	}, map[string]bool{"tools": true, compose.END: true})); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("tools", "llm_think"); err != nil {
		return nil, err
	}
	runnable, err := graph.Compile(ctx, compose.WithMaxRunSteps(maxLoops*3+2))
	if err != nil {
		return nil, err
	}
	return &Agent{model: model, tools: tools, maxLoops: maxLoops, runnable: runnable}, nil
}

func (a *Agent) Run(ctx context.Context, question string) (*AgentState, error) {
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}
	return a.runnable.Invoke(ctx, newAgentState(question, a.maxLoops))
}
