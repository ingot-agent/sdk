// Package operation defines externally invocable request-response capabilities
// exposed by plugins to application ingress components.
//
// Operations form an open set at an application's external boundary. They are
// not a service locator for calls between ordinary components; reusable
// component-to-component behavior belongs in a typed capability contract.
// Application ingress components snapshot all definitions before accepting
// calls, reject malformed or duplicate definitions, and validate each input
// and successful output against the corresponding schema.
package operation

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ingot-agent/sdk/interaction"
	"github.com/ingot-agent/sdk/session"
)

// Definition describes one externally invocable operation. Name is its stable
// lowercase ASCII protocol identity. It matches
// ^[a-z][a-z0-9]*([._-][a-z0-9]+)*$ and does not include transport syntax such
// as a leading slash or an HTTP path. Description is non-empty UTF-8 text.
// InputSchema and OutputSchema use JSON Schema Draft 2020-12 and each describes
// a JSON object; an omitted $schema keyword is interpreted as Draft 2020-12.
// Values returned by Operation.Definition are caller-owned snapshots,
// including both schema byte slices.
type Definition struct {
	Name         string
	Description  string
	InputSchema  json.RawMessage
	OutputSchema json.RawMessage
}

// Request is one invocation supplied by an application ingress. An empty
// SessionID means the invocation is not scoped to an agent session. Input is an
// immutable JSON object for the duration of Invoke. Interaction must be non-nil
// and is scoped to this invocation; an Operation must not retain or call it
// after Invoke returns.
type Request struct {
	SessionID   session.ID
	Input       json.RawMessage
	Interaction interaction.Channel
}

// Result is the final machine-readable result of an invocation. Output is a
// JSON object satisfying Definition.OutputSchema and is owned by the caller on
// return. When Invoke returns an error, Output has no defined meaning.
type Result struct {
	Output json.RawMessage
}

// Operation describes and executes one externally invocable operation.
// Implementations are safe for concurrent calls; an implementation may
// serialize calls internally when required by the resource or session it
// manages. Invoke observes context cancellation while waiting or executing and
// does not return while any operation-owned work still uses Request values.
type Operation interface {
	// Definition returns the same semantic definition for the lifetime of the
	// Operation while transferring ownership of returned schema bytes.
	Definition() Definition
	// Invoke executes one request and returns its final machine-readable result.
	Invoke(context.Context, Request) (Result, error)
}

// ErrUnavailable indicates that an operation is currently or conditionally
// unavailable in the requested scope.
var ErrUnavailable = errors.New("operation unavailable")
