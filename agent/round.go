package agent

import (
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/pipeline"
	"github.com/ingot-agent/sdk/session"
)

// Round is one model invocation and the canonical assistant decision proposed
// from its response. Invocation and Response are execution facts; interceptors
// may only make contract-compliant changes to Decision.
//
// The aggregate and all nested mutable values passed to an interceptor are
// owned by that interceptor.
type Round struct {
	// SessionID identifies the session executing this round.
	SessionID session.ID
	// Index is zero-based within the current turn.
	Index int
	// Invocation is the exact request after compaction and before the model
	// runtime invocation.
	Invocation model.Request
	// Response is the original validated model response. It is an immutable
	// execution fact.
	Response model.Response
	// Decision is the canonical assistant decision currently proposed for
	// agent acceptance. Its initial value is a deep copy of Response.Message.
	Decision model.Message
}

// RoundResult is the canonical assistant decision and its ordered, persisted
// tool result messages. ToolMessages is empty for a final round.
//
// The returned aggregate and all nested mutable values are owned by the caller.
type RoundResult struct {
	// Decision is the canonical assistant decision accepted and persisted by
	// the agent.
	Decision model.Message
	// ToolMessages contains persisted tool result messages in execution order.
	ToolMessages []model.Message
}

// RoundInterceptor wraps the control and execution phase of an agent round.
// It observes a complete model decision before any tool side effect occurs.
type RoundInterceptor = pipeline.Interceptor[Round, RoundResult]
