// Package tool defines tool providers and the tool runtime chokepoint.
package tool

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ingot-agent/sdk/content"
	"github.com/ingot-agent/sdk/execution"
	"github.com/ingot-agent/sdk/pipeline"
)

// Definition describes a tool and the JSON Schema accepted by its arguments.
type Definition struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

// Call is one model- or user-originated tool invocation. It is the durable,
// domain-level payload of a tool call: it enters model messages and session
// history, can be serialized, restored, and displayed by observation. Call
// deliberately carries no execution metadata; runtime execution identity lives
// in Invocation.
type Call struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// Invocation is the runtime execution envelope for one tool call. It carries
// the explicit dynamic execution Scope of this call together with the durable
// domain payload Call.
//
// Scope is immutable execution identity; a Call payload is a domain value.
// Runtime and Tool implementations must treat Invocation.Scope as read-only
// execution identity that they must not modify or discard.
type Invocation struct {
	Scope execution.Scope
	Call  Call
}

// Result is the ordered multimodal, authoritative outcome of a tool invocation.
// It is valid only when the operation returns a nil error. Known business
// outcomes, including unsuccessful ones, should be represented as Result with
// a nil error. The returned aggregate and nested inline data are owned by the
// caller.
type Result struct {
	Content content.Content
}

// Tool defines and executes one tool. Implementations are safe for concurrent
// calls unless they serialize access internally as part of their semantics. A
// Result is authoritative only when Invoke returns nil; an error does not imply
// that external side effects did not occur or that retrying is safe.
//
// Invoke receives the full execution envelope; Invocation.Scope identifies the
// execution domain this call belongs to.
type Tool interface {
	Definition() Definition
	Invoke(context.Context, Invocation) (Result, error)
}

// Runtime is the standard lookup, validation, interception, and invocation
// chokepoint for tools. A non-nil error means no reliable Result exists and,
// except for the documented pre-dispatch sentinels below, does not prove that
// external side effects did not occur or that retrying is safe.
//
// Call receives the full execution envelope; Invocation.Scope is propagated
// unchanged to every Tool and Interceptor.
type Runtime interface {
	Definitions() []Definition
	Call(context.Context, Invocation) (Result, error)
}

// Interceptor wraps a tool invocation.
type Interceptor = pipeline.Interceptor[Invocation, Result]

var (
	// ErrNotFound indicates that lookup rejected the call before Tool.Invoke was
	// dispatched. A standard Runtime must not return it after dispatch.
	ErrNotFound = errors.New("tool not found")
	// ErrInvalidArguments indicates that JSON or schema validation rejected the
	// call before Tool.Invoke was dispatched. A standard Runtime must not return
	// it after dispatch.
	ErrInvalidArguments = errors.New("invalid tool arguments")
)
