package execution_test

import (
	"testing"

	"github.com/ingot-agent/sdk/execution"
	"github.com/ingot-agent/sdk/session"
)

// TestScopeIsPlainValue verifies Scope is a plain immutable-by-convention
// value that carries the SessionID with no hidden context requirement.
func TestScopeIsPlainValue(t *testing.T) {
	scope := execution.Scope{SessionID: session.ID("session-a")}
	if scope.SessionID != session.ID("session-a") {
		t.Fatalf("scope session id = %q", scope.SessionID)
	}
	// Scope is a comparable, copyable value: copies do not alias.
	other := scope
	other.SessionID = session.ID("session-b")
	if scope.SessionID != session.ID("session-a") {
		t.Fatalf("scope mutated through copy: %q", scope.SessionID)
	}
}

// TestZeroScopeCarriesEmptySession verifies a zero Scope is representable and
// carries the zero SessionID so callers can detect unbound scopes explicitly.
func TestZeroScopeCarriesEmptySession(t *testing.T) {
	var scope execution.Scope
	if scope.SessionID != session.ID("") {
		t.Fatalf("zero scope session id = %q", scope.SessionID)
	}
}
