package agent_test

import (
	"context"
	"testing"

	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/content"
)

// Deliberately has no Run method: streaming is an independent capability.
type streamingOnly struct{}

func (streamingOnly) Stream(_ context.Context, _ agent.Turn, handler agent.StreamHandler) (agent.Result, error) {
	if handler == nil {
		return agent.Result{}, agent.ErrNilStreamHandler
	}
	for _, event := range []agent.StreamEvent{
		{Kind: agent.StreamReasoningDelta, TextDelta: "thinking"},
		{Kind: agent.StreamOutputDelta, TextDelta: "working"},
	} {
		if err := handler(event); err != nil {
			return agent.Result{}, err
		}
	}
	return agent.Result{Output: content.FromText("final")}, nil
}

func TestIndependentStreamingContract(t *testing.T) {
	var runtime agent.StreamingRuntime = streamingOnly{}
	var events []agent.StreamEvent
	result, err := runtime.Stream(context.Background(), agent.Turn{}, func(event agent.StreamEvent) error {
		events = append(events, event)
		return nil
	})
	if err != nil || len(events) != 2 || result.Output[0].Text != "final" {
		t.Fatalf("result=%v events=%v err=%v", result, events, err)
	}
}
