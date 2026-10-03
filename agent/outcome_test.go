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
		Failure:  &agent.Failure{Stage: agent.FailureModel, RoundIndex: &round},
	}}
	if execution.Result != nil || execution.Outcome.Status != agent.OutcomeFailed ||
		execution.Outcome.Duration != time.Second || execution.Outcome.Failure.RoundIndex == nil ||
		*execution.Outcome.Failure.RoundIndex != round {
		t.Fatalf("execution=%#v", execution)
	}
	if (agent.Execution{}).Outcome.Status != 0 {
		t.Fatal("zero value must represent no established outcome")
	}
}
