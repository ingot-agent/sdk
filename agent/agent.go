// Package agent defines the agent turn runtime chokepoint.
package agent

import (
	"context"

	"github.com/ingot-agent/sdk/content"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/pipeline"
	"github.com/ingot-agent/sdk/session"
)

// Turn is one user input in a session.
type Turn struct {
	SessionID   session.ID
	Input       string
	Attachments []content.Attachment
}

// Result is the ordered multimodal result of an agent turn. It is canonical
// only when the operation returns a nil error; callers must ignore it on any
// error. The returned aggregate and nested inline data are owned by the caller.
type Result struct {
	Output content.Content
}

// Runtime executes agent turns. Once a turn lifecycle is established,
// Execution.Outcome is valid even when Run returns an error; pre-lifecycle
// validation failures return a zero Execution. Turns for different sessions
// may run in parallel; turns for one session are serialized in call order.
// Cancellation stops future work but does not roll back durable progress or
// external side effects, and an error does not imply that retrying is safe.
type Runtime interface {
	Run(context.Context, Turn) (Execution, error)
}

// History loads the validated, persisted model messages for one session.
// Loads for the same session are serialized with Runtime.Run and
// StreamingRuntime.Stream; different sessions may be loaded concurrently.
// The returned aggregate and all nested
// mutable values are owned by the caller.
type History interface {
	Load(context.Context, session.ID) ([]model.Message, error)
}

// Interceptor wraps an agent turn.
type Interceptor = pipeline.Interceptor[Turn, Result]
