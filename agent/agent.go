// Package agent defines the agent turn runtime chokepoint.
package agent

import (
	"context"

	"github.com/ingot-agent/sdk/content"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/pipeline"
	"github.com/ingot-agent/sdk/session"
)

// Turn is one execution lifecycle. Inputs contains the ordered submissions
// admitted to it; pending submissions and preparation failures are excluded.
type Turn struct {
	// RootSessionID owns cumulative usage across nested executions. An empty
	// value defaults to SessionID for ordinary turns; child and followup
	// dispatchers must explicitly supply their outermost accounting owner.
	RootSessionID session.ID
	SessionID     session.ID
	Inputs        []Input
}

// Result is the ordered multimodal result of an agent turn. It is canonical
// only when the operation returns a nil error; callers must ignore it on any
// error. The returned aggregate and nested inline data are owned by the caller.
type Result struct {
	Output content.Content
}

// Runtime accepts messages and owns their scheduling. SendMessage returns once
// accepted, not when execution finishes. Its context controls admission only;
// accepted work runs under the runtime's context and is canceled through Control.
// A successful submission need not yet be persisted or seen by the model.
type Runtime interface {
	SendMessage(context.Context, MessageRequest) error
}

// History loads the validated, persisted model messages for one session.
// Loads for the same session are serialized with complete turn lifecycles;
// SendMessage and state queries do not wait for that execution gate.
// Agents using a PluginInputProjector project recognized records to user messages.
// Inputs appended within an incomplete tool round are deferred until its results are
// complete or recovered. Load itself is read-only and does not perform recovery.
// The returned aggregate and all nested
// mutable values are owned by the caller.
type History interface {
	Load(context.Context, session.ID) ([]model.Message, error)
}

// Interceptor wraps an agent turn.
type Interceptor = pipeline.Interceptor[Turn, Result]
