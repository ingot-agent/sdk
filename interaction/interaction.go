// Package interaction defines a single logical frontend interaction channel.
package interaction

import (
	"context"
	"errors"

	"github.com/ingot-agent/sdk/tool"
)

// Channel provides synchronous user interaction and event rendering.
// Render is concurrent-safe. Ask and ReadLine are serialized with each other
// and observe context cancellation while waiting for serialization or input.
type Channel interface {
	Ask(context.Context, AskRequest) (AskResponse, error)
	ReadLine(context.Context, string) (string, error)
	Render(context.Context, Event) error
}

// AskRequest requests structured text input.
type AskRequest struct {
	Prompt string
}

// AskResponse contains structured text input.
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
