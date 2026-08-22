// Package agent defines the agent turn runtime chokepoint.
package agent

import (
	"context"

	"github.com/ingot-agent/sdk/pipeline"
	"github.com/ingot-agent/sdk/session"
)

// Turn is one user input in a session.
type Turn struct {
	SessionID session.ID
	Input     string
}

// Result is the text result of an agent turn.
type Result struct {
	Output string
}

// Runtime executes agent turns. Turns for different sessions may run in
// parallel; turns for one session are serialized in call order.
type Runtime interface {
	Run(context.Context, Turn) (Result, error)
}

// Interceptor wraps an agent turn.
type Interceptor = pipeline.Interceptor[Turn, Result]
