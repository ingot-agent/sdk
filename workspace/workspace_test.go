package workspace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ingot-agent/sdk/execution"
	"github.com/ingot-agent/sdk/session"
	"github.com/ingot-agent/sdk/workspace"
)

type fakeResolver struct {
	binding workspace.Binding
	err     error
}

func (f *fakeResolver) Resolve(_ context.Context, scope execution.Scope) (workspace.Binding, error) {
	if f.err != nil {
		return workspace.Binding{}, f.err
	}
	if scope.SessionID == session.ID("unknown") {
		return workspace.Binding{}, session.ErrNotFound
	}
	return f.binding, nil
}

// TestResolverIsExecutionSide verifies Resolver reads a Binding from an
// execution.Scope and that no context convention is required.
func TestResolverIsExecutionSide(t *testing.T) {
	resolver := &fakeResolver{binding: workspace.Binding{Root: "/repo/a"}}
	binding, err := resolver.Resolve(context.Background(), execution.Scope{SessionID: session.ID("session-a")})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if binding.Root != "/repo/a" {
		t.Fatalf("binding root = %q", binding.Root)
	}
	if _, err := resolver.Resolve(context.Background(), execution.Scope{SessionID: session.ID("unknown")}); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("unknown session resolve error = %v", err)
	}
}

type fakeManager struct {
	assigned map[session.ID]workspace.Binding
}

func (f *fakeManager) Assign(_ context.Context, id session.ID, binding workspace.Binding) error {
	if id == session.ID("missing") {
		return session.ErrNotFound
	}
	if binding.Root == "" || binding.Root[0] != '/' {
		return workspace.ErrInvalidBinding
	}
	if _, exists := f.assigned[id]; exists {
		return workspace.ErrAlreadyAssigned
	}
	f.assigned[id] = binding
	return nil
}

// TestManagerSentinelErrors verifies the immutable one-binding-per-session
// assignment semantics through the public sentinel errors.
func TestManagerSentinelErrors(t *testing.T) {
	manager := &fakeManager{assigned: make(map[session.ID]workspace.Binding)}
	if err := manager.Assign(context.Background(), session.ID("a"), workspace.Binding{Root: "/repo/a"}); err != nil {
		t.Fatalf("assign a: %v", err)
	}
	if err := manager.Assign(context.Background(), session.ID("a"), workspace.Binding{Root: "/repo/b"}); !errors.Is(err, workspace.ErrAlreadyAssigned) {
		t.Fatalf("duplicate assign error = %v", err)
	}
	if err := manager.Assign(context.Background(), session.ID("missing"), workspace.Binding{Root: "/repo/a"}); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("unknown session assign error = %v", err)
	}
	if err := manager.Assign(context.Background(), session.ID("b"), workspace.Binding{Root: "relative"}); !errors.Is(err, workspace.ErrInvalidBinding) {
		t.Fatalf("invalid binding error = %v", err)
	}
}
