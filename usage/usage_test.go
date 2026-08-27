package usage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/usage"
)

type counterFunc func(context.Context, usage.CountRequest) (usage.CountResult, error)

func (f counterFunc) CountInput(ctx context.Context, request usage.CountRequest) (usage.CountResult, error) {
	return f(ctx, request)
}

func TestPublicCounterContract(t *testing.T) {
	t.Parallel()
	var counter usage.Counter = counterFunc(func(_ context.Context, request usage.CountRequest) (usage.CountResult, error) {
		return usage.CountResult{
			InputTokens: 1,
			Accuracy:    usage.AccuracyEstimate,
			Source:      "test-v1",
			Provider:    request.Invocation.Provider,
			Model:       request.Invocation.Model,
		}, nil
	})
	result, err := counter.CountInput(context.Background(), usage.CountRequest{
		Invocation: model.Request{Provider: "provider", Model: "model"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InputTokens != 1 || result.Accuracy != usage.AccuracyEstimate || result.Provider != "provider" || result.Model != "model" {
		t.Fatalf("result = %#v", result)
	}
	if !errors.Is(usage.ErrUnsupportedModel, usage.ErrUnsupportedModel) {
		t.Fatal("unsupported model sentinel is not comparable through errors.Is")
	}
}
