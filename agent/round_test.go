package agent_test

import (
	"context"
	"testing"

	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/pipeline"
)

type roundInterceptor struct{}

func (roundInterceptor) Invoke(
	ctx context.Context,
	round agent.Round,
	next pipeline.Next[agent.Round, agent.RoundResult],
) (agent.RoundResult, error) {
	return next(ctx, round)
}

func TestRoundInterceptorContract(t *testing.T) {
	var interceptor agent.RoundInterceptor = roundInterceptor{}
	if interceptor == nil {
		t.Fatal("round interceptor is nil")
	}

	round := agent.Round{
		SessionID:  "session-1",
		Index:      2,
		Invocation: model.Request{Provider: "provider", Model: "model"},
		Response:   model.Response{Message: model.Message{Role: model.RoleAssistant}},
		Decision:   model.Message{Role: model.RoleAssistant},
	}
	if round.SessionID != "session-1" || round.Index != 2 {
		t.Fatalf("round=%#v", round)
	}
}
