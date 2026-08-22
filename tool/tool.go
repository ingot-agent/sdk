// Package tool defines tool providers and the tool runtime chokepoint.
package tool

import (
	"context"
	"encoding/json"
	"errors"

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

// Result is the text result of a tool invocation.
type Result struct {
	Content string
}

// Tool defines and executes one tool. Implementations are safe for concurrent
// calls unless they serialize access internally as part of their semantics.
type Tool interface {
	Definition() Definition
	Invoke(context.Context, Call) (Result, error)
}

// Runtime is the standard lookup, validation, interception, and invocation
// chokepoint for tools.
type Runtime interface {
	Definitions() []Definition
	Call(context.Context, Call) (Result, error)
}

// Interceptor wraps a tool call.
type Interceptor = pipeline.Interceptor[Call, Result]

var (
	// ErrNotFound indicates that no tool has the requested name.
	ErrNotFound = errors.New("tool not found")
	// ErrInvalidArguments indicates that a call does not satisfy the tool's
	// input schema.
	ErrInvalidArguments = errors.New("invalid tool arguments")
)
