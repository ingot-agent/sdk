package model_test

import (
	"context"
	"testing"

	"github.com/ingot-agent/sdk/model"
)

type resolverFunc func(context.Context, model.Request) (model.Request, error)

func TestStreamSemanticDefaultsToContent(t *testing.T) {
	var event model.StreamEvent
	if event.Semantic != model.StreamSemanticContent || model.StreamSemanticContent == model.StreamSemanticReasoning {
		t.Fatal("events must default to content, distinct from reasoning")
	}
}

func TestUsageReportedDistinguishesZeroFromUnavailable(t *testing.T) {
	if (model.Usage{}).Reported {
		t.Fatal("zero usage must be unavailable")
	}
	if !(model.Usage{Reported: true}).Reported {
		t.Fatal("reported zero usage must retain presence")
	}
}

func (f resolverFunc) ResolveRequest(ctx context.Context, request model.Request) (model.Request, error) {
	return f(ctx, request)
}

func TestReasoningEffortMaxIsExplicit(t *testing.T) {
	t.Parallel()
	explicit := []model.ReasoningEffort{
		model.ReasoningEffortNone,
		model.ReasoningEffortMinimal,
		model.ReasoningEffortLow,
		model.ReasoningEffortMedium,
		model.ReasoningEffortHigh,
		model.ReasoningEffortXHigh,
		model.ReasoningEffortMax,
		model.ReasoningEffortUltra,
	}
	for _, effort := range explicit {
		if !effort.Valid() {
			t.Fatalf("effort %q must be valid", effort)
		}
	}
	if model.ReasoningEffortMax != "max" || model.ReasoningEffortUltra != "ultra" {
		t.Fatalf("ReasoningEffortMax = %q, ReasoningEffortUltra = %q", model.ReasoningEffortMax, model.ReasoningEffortUltra)
	}
	if model.ReasoningEffortProviderDefault.Valid() || model.ReasoningEffort("").Valid() {
		t.Fatal("providerDefault and empty effort are request controls, not capabilities")
	}
}

func TestPublicRequestResolverContract(t *testing.T) {
	t.Parallel()
	var resolver model.RequestResolver = resolverFunc(func(_ context.Context, request model.Request) (model.Request, error) {
		request.Provider = "provider"
		request.Model = "model"
		return request, nil
	})
	resolved, err := resolver.ResolveRequest(context.Background(), model.Request{})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Provider != "provider" || resolved.Model != "model" {
		t.Fatalf("resolved = %#v", resolved)
	}
}
