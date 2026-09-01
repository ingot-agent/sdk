package observation

import (
	"time"

	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/tool"
)

// Event is one materialized, detached execution fact. Sequence is strictly
// increasing within one TurnID and starts at one; no ordering is promised
// across turns. Time records materialization by the Consumer, not delivery to
// an Observer.
type Event struct {
	Time        time.Time
	Sequence    uint64
	Correlation Correlation
	Detail      Detail
}

// Detail is the closed family of Agent execution observation facts. Only this
// package can define Detail implementations.
type Detail interface {
	observationDetail()
}

// Status identifies the terminal state of an execution scope.
type Status uint8

const (
	// StatusSucceeded means the scope produced its complete result.
	StatusSucceeded Status = iota + 1
	// StatusFailed means the scope ended with a non-cancellation error.
	StatusFailed
	// StatusCanceled means cancellation or a deadline ended the scope.
	StatusCanceled
)

// TurnStarted reports the immutable invocation snapshot that began a turn.
type TurnStarted struct {
	Turn agent.Turn
}

// TurnFinished reports exactly one terminal outcome for a started turn.
// Result is non-nil only for StatusSucceeded; Error is non-empty only for
// StatusFailed and StatusCanceled.
type TurnFinished struct {
	Status Status
	Result *agent.Result
	Error  string
}

// RoundStarted reports that the correlated zero-based round attempt began.
type RoundStarted struct{}

// RoundFinished reports exactly one terminal outcome for a started round.
// Result is the complete canonical round and is non-nil only on success.
type RoundFinished struct {
	Status Status
	Result *agent.RoundResult
	Error  string
}

// ModelStarted reports the actual post-compaction request submitted to the
// model runtime.
type ModelStarted struct {
	Request model.Request
}

// ModelProgress reports one provisional event produced by model streaming.
type ModelProgress struct {
	Progress model.StreamEvent
}

// ModelFinished reports exactly one terminal outcome for a started model
// invocation. Response is the original model execution fact and is non-nil
// only on success.
type ModelFinished struct {
	Status   Status
	Response *model.Response
	Error    string
}

// ToolStarted reports the canonical call submitted to the tool runtime.
type ToolStarted struct {
	Call tool.Call
}

// ToolProgress reports one transient fact emitted by a running Tool.
type ToolProgress struct {
	Progress tool.Progress
}

// ToolFinished reports exactly one terminal outcome for a started Tool call.
// Result is the direct tool runtime result and is non-nil only on success.
type ToolFinished struct {
	Status Status
	Result *tool.Result
	Error  string
}

func (TurnStarted) observationDetail()   {}
func (TurnFinished) observationDetail()  {}
func (RoundStarted) observationDetail()  {}
func (RoundFinished) observationDetail() {}
func (ModelStarted) observationDetail()  {}
func (ModelProgress) observationDetail() {}
func (ModelFinished) observationDetail() {}
func (ToolStarted) observationDetail()   {}
func (ToolProgress) observationDetail()  {}
func (ToolFinished) observationDetail()  {}
