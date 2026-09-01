package observation

import (
	"context"

	"github.com/ingot-agent/sdk/session"
)

// ID is the opaque identity of one Agent turn. The Agent runtime creates it at
// turn start and propagates it to every child execution scope.
type ID string

// Correlation identifies the execution scope in which a Detail was produced.
// RoundIndex is authoritative for round, model, and tool events. ToolCallID is
// authoritative only for tool events. The remaining fields are ignored where
// they do not apply to a Detail type.
type Correlation struct {
	SessionID  session.ID
	TurnID     ID
	RoundIndex int
	ToolCallID string
}

type correlationContextKey struct{}

// WithCorrelation returns a child context carrying correlation for execution
// observation. The value is immutable and replaces any correlation already in
// ctx.
func WithCorrelation(ctx context.Context, correlation Correlation) context.Context {
	return context.WithValue(ctx, correlationContextKey{}, correlation)
}

// CorrelationFromContext returns the execution correlation carried by ctx.
// The boolean is false when ctx is nil or no correlation has been attached.
func CorrelationFromContext(ctx context.Context) (Correlation, bool) {
	if ctx == nil {
		return Correlation{}, false
	}
	correlation, ok := ctx.Value(correlationContextKey{}).(Correlation)
	return correlation, ok
}
