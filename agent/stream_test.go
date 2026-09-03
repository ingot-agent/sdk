package agent_test

import (
	"context"
	"testing"

	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/content"
)

// Deliberately has no Run method: streaming is an independent capability.
type streamingOnly struct{}

func (streamingOnly) Stream(_ context.Context, _ agent.Turn, handler agent.StreamHandler) (agent.Execution, error) {
	if handler == nil {
		return agent.Execution{}, agent.ErrNilStreamHandler
	}
	for _, event := range []agent.StreamEvent{
		{Kind: agent.StreamReasoningDelta, TextDelta: "thinking"},
		{Kind: agent.StreamOutputDelta, TextDelta: "working"},
	} {
		if err := handler(event); err != nil {
			return agent.Execution{}, err
		}
	}
	return agent.Execution{
		Result:  &agent.Result{Output: content.FromText("final")},
		Outcome: agent.Outcome{Status: agent.OutcomeSucceeded},
	}, nil
}

func TestIndependentStreamingContract(t *testing.T) {
	var runtime agent.StreamingRuntime = streamingOnly{}
	var events []agent.StreamEvent
	result, err := runtime.Stream(context.Background(), agent.Turn{}, func(event agent.StreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil || len(events) != 2 || result.Result == nil || result.Result.Output[0].Text != "final" ||
		result.Outcome.Status != agent.OutcomeSucceeded {
		t.Fatalf("result=%v events=%v err=%v", result, events, err)
	}
}
