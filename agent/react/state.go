package react

import "github.com/cloudwego/eino/schema"

// AgentState is the shared memory passed through every ReAct graph node.
type AgentState struct {
	UserQuestion string
	Messages     []*schema.Message
	Thought      string
	ToolCalls    []schema.ToolCall
	Observations []Observation
	FinalAnswer  string
	LoopCount    int
	MaxLoops     int
	Done         bool
}

// Observation records one tool result and its corresponding call.
type Observation struct {
	ToolCallID string
	ToolName   string
	Result     string
}

func newAgentState(question string, maxLoops int) *AgentState {
	if maxLoops < 1 {
		maxLoops = 1
	}
	return &AgentState{
		UserQuestion: question,
		Messages:     []*schema.Message{schema.UserMessage(question)},
		MaxLoops:     maxLoops,
	}
}
