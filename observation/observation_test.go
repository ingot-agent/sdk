package observation_test

import (
	"context"
	"testing"
	"time"

	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/observation"
	"github.com/ingot-agent/sdk/session"
)

type consumer struct{}

func (consumer) Emit(context.Context, observation.Detail) {}

type observer struct{}

func (observer) Observe(observation.Event) {}

func TestPublicContractsAndCorrelation(t *testing.T) {
	var _ observation.Consumer = consumer{}
	var _ observation.Observer = observer{}

	want := observation.Correlation{
		SessionID: session.ID("session-1"), TurnID: observation.ID("turn-1"),
		RoundIndex: 2, ToolCallID: "call-1",
	}
	ctx := observation.WithCorrelation(context.Background(), want)
	got, ok := observation.CorrelationFromContext(ctx)
	if !ok || got != want {
		t.Fatalf("correlation=%#v ok=%v", got, ok)
	}
	if _, ok := observation.CorrelationFromContext(context.Background()); ok {
		t.Fatal("background context unexpectedly carried correlation")
	}
	if _, ok := observation.CorrelationFromContext(nil); ok {
		t.Fatal("nil context unexpectedly carried correlation")
	}

	event := observation.Event{
		Time: time.Unix(1, 2), Sequence: 1, Correlation: want,
		Detail: observation.TurnStarted{Turn: agent.Turn{SessionID: "session-1", Input: "hello"}},
	}
	if event.Sequence != 1 || event.Detail == nil {
		t.Fatalf("event=%#v", event)
	}
}

func TestDetailFamilyAndStatuses(t *testing.T) {
	details := []observation.Detail{
		observation.TurnStarted{}, observation.TurnFinished{},
		observation.RoundStarted{}, observation.RoundFinished{},
		observation.ModelStarted{}, observation.ModelProgress{}, observation.ModelFinished{},
		observation.ToolStarted{}, observation.ToolProgress{}, observation.ToolFinished{},
	}
	if len(details) != 10 {
		t.Fatalf("details=%d", len(details))
	}
	if observation.StatusSucceeded == 0 || observation.StatusFailed == observation.StatusCanceled {
		t.Fatal("invalid status constants")
	}
}
