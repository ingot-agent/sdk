package agent_test

import (
	"testing"
	"time"

	"github.com/ingot-agent/sdk/agent"
)

func TestExecutionOutcomeContracts(t *testing.T) {
	round := 2
	execution := agent.Execution{Outcome: agent.Outcome{
		Status:   agent.OutcomeFailed,
		Duration: time.Second,
		Accounting: agent.Accounting{
			Rounds: 3, ModelInvocations: 4, ToolCalls: 2,
			Usage: agent.TokenUsage{TotalTokens: 7, Coverage: agent.UsagePartial},
			Models: []agent.ModelAccounting{{
				Provider: "provider", Model: "model", CompletedInvocations: 3,
				Usage: agent.TokenUsage{TotalTokens: 7, Coverage: agent.UsageComplete},
			}},
		},
		Failure: &agent.Failure{Stage: agent.FailureModel, RoundIndex: &round},
	}}
	if execution.Result != nil || execution.Outcome.Failure.RoundIndex == nil ||
		*execution.Outcome.Failure.RoundIndex != round {
		t.Fatalf("execution=%#v", execution)
	}
	if (agent.Execution{}).Outcome.Status != 0 || agent.UsageUnavailable != 0 {
		t.Fatal("zero values must represent no established outcome and unavailable usage")
	}
}
