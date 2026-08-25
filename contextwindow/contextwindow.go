// Package contextwindow defines model-context compaction capabilities.
package contextwindow

import (
	"context"

	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/session"
)

// CompactionRequest describes one model invocation whose message context may
// be compacted. Invocation is immutable-by-contract.
type CompactionRequest struct {
	SessionID  session.ID
	Invocation model.Request
}

// CompactionResult contains the complete message sequence to use for the
// invocation. Ownership of Messages passes to the caller when Compact returns.
type CompactionResult struct {
	Messages []model.Message
	Changed  bool
}

// Compactor prepares a bounded message context for a model invocation.
// Implementations are safe for concurrent calls and preserve any required
// per-session ordering when they reuse or persist compaction state.
type Compactor interface {
	Compact(context.Context, CompactionRequest) (CompactionResult, error)
}
