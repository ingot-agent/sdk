// Package interaction defines a single logical frontend interaction channel.
package interaction

import (
	"context"
	"errors"

	"github.com/ingot-agent/sdk/tool"
)

// Channel provides synchronous user interaction and event rendering.
// Render is concurrent-safe. Ask operations are serialized with each other
// and observe context cancellation while waiting for serialization or input.
type Channel interface {
	Ask(context.Context, AskRequest) (AskResponse, error)
	Render(context.Context, Event) error
}

// AskOption describes one ordered choice presented by an interaction frontend.
// Label is the value returned in AskResponse.Text when the option is selected.
type AskOption struct {
	Label       string
	Description string
}

// AskRequest requests synchronous user input. When Options is non-empty, the
// frontend presents them in order. When Options is empty the frontend asks a
// plain free-form question; the user's input is returned as-is in
// AskResponse.Text. AllowTextInput requires an additional free-form choice;
// selecting it returns the user's original text.
type AskRequest struct {
	Prompt         string
	Options        []AskOption
	AllowTextInput bool
}

// AskResponse contains either the selected option label or free-form input.
type AskResponse struct {
	Text string
}

// Event is the closed set of renderable SDK interaction events.
type Event interface {
	interactionEvent()
}

// TextEvent renders ordinary text.
type TextEvent struct {
	Text string
}

func (TextEvent) interactionEvent() {}

// StatusEvent renders transient or durable status text.
type StatusEvent struct {
	Text string
}

func (StatusEvent) interactionEvent() {}

// ErrorEvent renders an error.
type ErrorEvent struct {
	Err error
}

func (ErrorEvent) interactionEvent() {}

// ToolCallEvent announces a tool invocation.
type ToolCallEvent struct {
	Call tool.Call
}

func (ToolCallEvent) interactionEvent() {}

// ToolResultEvent announces the result of a tool invocation.
type ToolResultEvent struct {
	Call   tool.Call
	Result tool.Result
}

func (ToolResultEvent) interactionEvent() {}

// ErrUnavailable indicates that no interaction frontend can service an
// operation.
var ErrUnavailable = errors.New("interaction unavailable")
