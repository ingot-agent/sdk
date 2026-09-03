// Package tool defines tool providers and the tool runtime chokepoint.
package tool

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ingot-agent/sdk/content"
	"github.com/ingot-agent/sdk/pipeline"
)

// Definition describes a tool and the JSON Schema accepted by its arguments.
type Definition struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

// Call is one model- or user-originated tool invocation.
type Call struct {
	ID        string
	Name      string
	Arguments json.RawMessage
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
type Tool interface {
	Definition() Definition
	Invoke(context.Context, Call) (Result, error)
}

// Runtime is the standard lookup, validation, interception, and invocation
// chokepoint for tools. A non-nil error means no reliable Result exists and,
// except for the documented pre-dispatch sentinels below, does not prove that
// external side effects did not occur or that retrying is safe.
type Runtime interface {
	Definitions() []Definition
	Call(context.Context, Call) (Result, error)
}

// Interceptor wraps a tool call.
type Interceptor = pipeline.Interceptor[Call, Result]

var (
	// ErrNotFound indicates that lookup rejected the call before Tool.Invoke was
	// dispatched. A standard Runtime must not return it after dispatch.
	ErrNotFound = errors.New("tool not found")
	// ErrInvalidArguments indicates that JSON or schema validation rejected the
	// call before Tool.Invoke was dispatched. A standard Runtime must not return
	// it after dispatch.
	ErrInvalidArguments = errors.New("invalid tool arguments")
)
