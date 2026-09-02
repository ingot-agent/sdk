package agent

import "time"

// Execution is the complete settlement envelope for one started turn.
// Result is non-nil only when Outcome.Status is OutcomeSucceeded. A zero
// Execution means the turn lifecycle was not established.
type Execution struct {
	Result  *Result
	Outcome Outcome
}

// Outcome describes how a turn terminated and the execution activity known
// at that point. It does not imply rollback, retry safety, durability, or the
// state of external side effects.
type Outcome struct {
	Status     OutcomeStatus
	Duration   time.Duration
	Accounting Accounting
	Failure    *Failure
}

// OutcomeStatus identifies the terminal state of a started turn.
type OutcomeStatus uint8

const (
	// OutcomeSucceeded means the turn produced a canonical Result.
	OutcomeSucceeded OutcomeStatus = iota + 1
	// OutcomeFailed means a non-cancellation error ended the turn.
	OutcomeFailed
	// OutcomeCanceled means cancellation or a deadline ended the turn.
	OutcomeCanceled
)

// Accounting contains turn-level execution attempts and known
// provider-reported token usage.
type Accounting struct {
	Rounds           int
	ModelInvocations int
	ToolCalls        int

	Usage  TokenUsage
	Models []ModelAccounting
}

// TokenUsage aggregates authoritative provider-reported token counts.
// Coverage states whether every invocation that could have consumed model
// resources was authoritatively settled.
type TokenUsage struct {
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
	Coverage     UsageCoverage
}

// UsageCoverage describes how much of the possible model execution usage is
// represented by a TokenUsage value.
type UsageCoverage uint8

const (
	// UsageUnavailable means no authoritative execution usage is available.
	UsageUnavailable UsageCoverage = iota
	// UsagePartial means known usage is present but at least one invocation's
	// possible usage is not authoritatively settled.
	UsagePartial
	// UsageComplete means every invocation's possible usage is settled,
	// including explicit rejections that could not consume model resources.
	UsageComplete
)

// ModelAccounting contains authoritative completed invocations attributed to
// one provider/model pair. Failed invocations without an authoritative
// response are intentionally not attributed.
type ModelAccounting struct {
	Provider string
	Model    string

	CompletedInvocations int
	Usage                TokenUsage
}

// FailureStage identifies the execution boundary at which a turn terminated.
// It is not a durability, rollback, side-effect, or retryability verdict.
type FailureStage uint8

const (
	// FailureSessionGate identifies session serialization or its immediate
	// cancellation boundary.
	FailureSessionGate FailureStage = iota + 1
	// FailureHistoryLoad identifies persisted history loading.
	FailureHistoryLoad
	// FailureRecovery identifies interrupted-round recovery.
	FailureRecovery
	// FailureUserPersistence identifies persistence of the user message.
	FailureUserPersistence
	// FailurePrompt identifies prompt rendering.
	FailurePrompt
	// FailureCompaction identifies model-context compaction.
	FailureCompaction
	// FailureModel identifies a model runtime operation.
	FailureModel
	// FailureRoundControl identifies round validation, interception, or limits.
	FailureRoundControl
	// FailureAssistantPersistence identifies persistence of an assistant decision.
	FailureAssistantPersistence
	// FailureTool identifies a canonical tool runtime operation.
	FailureTool
	// FailureToolResultPersistence identifies persistence of a tool result.
	FailureToolResultPersistence
	// FailureTurnControl identifies turn interception or result validation.
	FailureTurnControl
	// FailureStreamConsumer identifies cancellation by the stream consumer.
	FailureStreamConsumer
)

// Failure identifies where execution terminated. RoundIndex is present for a
// round-scoped stage. ToolCallID is populated only for a tool-related stage.
// The concrete cause remains the operation's returned error.
type Failure struct {
	Stage      FailureStage
	RoundIndex *int
	ToolCallID string
}
