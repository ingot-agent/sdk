// Package workspace defines the session-scoped local workspace binding
// capability.
//
// A Workspace Binding maps one Session to exactly one local working root.
// Workspace is session-scoped, not process-scoped or turn-scoped: one Runtime
// can carry many Sessions with different Workspaces concurrently. The binding
// is immutable after assignment. Workspace capability is separate from the
// session capability even when one implementation provides both.
//
// Resolver is the execution-side read capability: authoritative Workspace
// access flows only through it. Manager is the application-side mutation
// capability used by Application ingress such as a Web UI or IDE. Root
// represents a local working directory in this initial contract. A Workspace
// Root is a command-relative execution base; it does not imply filesystem
// confinement, shell sandboxing, container isolation, or path escape
// protection, which are independent security capabilities.
package workspace

import (
	"context"
	"errors"

	"github.com/ingot-agent/sdk/execution"
	"github.com/ingot-agent/sdk/session"
)

// Binding is the immutable-by-convention assignment of one Session to one
// local workspace root.
type Binding struct {
	// Root is an absolute local working directory. Implementations must reject
	// non-absolute or otherwise invalid roots at assignment time.
	Root string
}

// Resolver resolves the Workspace Binding of one dynamic execution scope.
// Implementations are safe for concurrent use. Resolve must verify that the
// target Session exists and that a Binding is present. Aggregate outputs are
// caller-owned.
type Resolver interface {
	Resolve(context.Context, execution.Scope) (Binding, error)
}

// Manager assigns Workspace Bindings to Sessions. Assign commits durably: a
// Session is assigned at most once in this contract, and repeated assignment
// fails with ErrAlreadyAssigned. Assigning to an unknown Session fails with a
// wrapped session.ErrNotFound. Assigning an invalid Binding fails with
// ErrInvalidBinding. The check-and-insert decision must be atomic so
// concurrent Assign and lifecycle operations cannot create a partial state.
type Manager interface {
	Assign(context.Context, session.ID, Binding) error
}

var (
	// ErrAlreadyAssigned indicates that the target Session already has an
	// immutable Workspace Binding.
	ErrAlreadyAssigned = errors.New("workspace already assigned")
	// ErrInvalidBinding indicates that the supplied Binding is not a legal
	// assignment for the target Session.
	ErrInvalidBinding = errors.New("invalid workspace binding")
	// ErrNotAssigned indicates that the target Session exists but has no
	// Workspace Binding. Resolver implementations return it from Resolve when
	// the Session is valid but unbound.
	ErrNotAssigned = errors.New("workspace not assigned")
)
