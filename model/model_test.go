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
		t.Fatal("legacy events must default to content, distinct from reasoning")
	}
}

func (f resolverFunc) ResolveRequest(ctx context.Context, request model.Request) (model.Request, error) {
	return f(ctx, request)
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
