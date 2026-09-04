package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"XiaoLong-Ridy/agent/react"
	"XiaoLong-Ridy/api/driver/internal/types"
)

const (
	agentQuestionMaxLength = 1000
	agentRunTimeout        = 5 * time.Second
)

func AgentChatHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AgentChatRequest
		if !decodeJSON(w, r, &req) {
			return
		}

		question := strings.TrimSpace(req.Question)
		if question == "" || len([]rune(question)) > agentQuestionMaxLength {
			writeError(w, http.StatusBadRequest, 50000, "question must contain 1 to 1000 characters")
			return
		}

		registry, err := react.NewDefaultToolRegistry()
		if err != nil {
			writeError(w, http.StatusInternalServerError, 50003, "agent tools are unavailable")
			return
		}
		agent, err := react.NewAgent(r.Context(), react.ScriptedModel{}, registry, 3)
		if err != nil {
			writeError(w, http.StatusInternalServerError, 50003, "agent initialization failed")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), agentRunTimeout)
		defer cancel()
		state, err := agent.Run(ctx, question)
		if err != nil {
			writeError(w, http.StatusInternalServerError, 50003, "agent execution failed")
			return
		}

		observations := make([]types.AgentObservation, 0, len(state.Observations))
		for _, observation := range state.Observations {
			observations = append(observations, types.AgentObservation{
				ToolName: observation.ToolName,
				Result:   observation.Result,
			})
		}
		writeSuccess(w, &types.AgentChatResponse{
			Answer:       state.FinalAnswer,
			LoopCount:    state.LoopCount,
			Observations: observations,
			Mode:         "scripted",
		})
	}
}
