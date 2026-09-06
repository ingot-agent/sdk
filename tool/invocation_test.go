package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ingot-agent/sdk/execution"
	"github.com/ingot-agent/sdk/pipeline"
	"github.com/ingot-agent/sdk/session"
	"github.com/ingot-agent/sdk/tool"
)

// recordingTool records the last Invocation it received and returns a fixed
// result, mirroring an independent tool implementation.
type recordingTool struct {
	last tool.Invocation
}

func (*recordingTool) Definition() tool.Definition {
	return tool.Definition{Name: "demo", InputSchema: json.RawMessage(`{"type":"object"}`)}
}

func (r *recordingTool) Invoke(_ context.Context, invocation tool.Invocation) (tool.Result, error) {
	r.last = invocation
	return tool.Result{}, nil
}

// TestInvocationCarriesScopeToTool verifies the execution Scope is delivered to
// the Tool through the public Invocation envelope.
func TestInvocationCarriesScopeToTool(t *testing.T) {
	toolImpl := &recordingTool{}
	scope := execution.Scope{SessionID: session.ID("session-a")}
	call := tool.Call{ID: "call-1", Name: "demo", Arguments: json.RawMessage(`{"x":1}`)}
	invocation := tool.Invocation{Scope: scope, Call: call}
	if _, err := toolImpl.Invoke(context.Background(), invocation); err != nil {
		t.Fatal(err)
	}
	if toolImpl.last.Scope != scope {
		t.Fatalf("tool scope = %#v want %#v", toolImpl.last.Scope, scope)
	}
	if toolImpl.last.Call.ID != call.ID || toolImpl.last.Call.Name != call.Name || string(toolImpl.last.Call.Arguments) != string(call.Arguments) {
		t.Fatalf("tool call = %#v", toolImpl.last.Call)
	}
}

// captureInterceptor reads an Invocation and forwards it unchanged.
type captureInterceptor struct {
	seen []tool.Invocation
}

func (c *captureInterceptor) Invoke(
	ctx context.Context,
	invocation tool.Invocation,
	next pipeline.Next[tool.Invocation, tool.Result],
) (tool.Result, error) {
	c.seen = append(c.seen, invocation)
	return next(ctx, invocation)
}

// TestInterceptorObservesAndPreservesScope verifies an Interceptor can see the
// Scope and does not accidentally lose it when forwarding the Invocation.
func TestInterceptorObservesAndPreservesScope(t *testing.T) {
	toolImpl := &recordingTool{}
	terminal := func(_ context.Context, invocation tool.Invocation) (tool.Result, error) {
		return toolImpl.Invoke(context.Background(), invocation)
	}
	interceptor := &captureInterceptor{}
	next := pipeline.Compose[tool.Invocation, tool.Result](terminal, interceptor)
	scope := execution.Scope{SessionID: session.ID("session-b")}
	call := tool.Call{ID: "call-2", Name: "demo", Arguments: json.RawMessage(`{}`)}
	if _, err := next(context.Background(), tool.Invocation{Scope: scope, Call: call}); err != nil {
		t.Fatal(err)
	}
	if len(interceptor.seen) != 1 || interceptor.seen[0].Scope != scope {
		t.Fatalf("interceptor seen = %#v", interceptor.seen)
	}
	if toolImpl.last.Scope != scope {
		t.Fatalf("tool scope after interceptor = %#v", toolImpl.last.Scope)
	}
}
