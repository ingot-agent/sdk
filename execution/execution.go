// Package execution defines the explicit dynamic execution identity of one
// runtime call.
//
// The static Component Graph describes what a Component depends on. Dynamic
// Execution Scope describes which execution domain one invocation belongs to.
// Correctness-critical execution identity must be represented by public
// request or invocation contracts and must not be communicated through hidden
// context values: context.Context carries only cancellation, deadline, and
// tracing/observation implementation detail that cannot change business
// results.
package execution

import "github.com/ingot-agent/sdk/session"

// Scope is the immutable-by-convention execution identity of one invocation.
//
// Scope is execution identity, not a capability container, metadata bag,
// service locator, or a replacement for context.Value. Aggregate values are
// immutable by contract; callers must copy mutable data they retain.
type Scope struct {
	// SessionID identifies the session whose execution domain this invocation
	// belongs to. Implementations that receive a zero SessionID must reject the
	// invocation unless their contract explicitly allows an unbound scope.
	SessionID session.ID
}
