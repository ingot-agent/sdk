package agent

import (
	"context"
	"time"
)

// StreamEventKind identifies transient output or a retry control event.
type StreamEventKind uint8

const (
	// StreamReasoningDelta is reasoning/thinking text explicitly returned by
	// the model provider. Agents must not infer reasoning from ordinary output.
	StreamReasoningDelta StreamEventKind = iota + 1
	// StreamOutputDelta is user-visible assistant text from any model round,
	// including rounds that subsequently invoke tools.
	StreamOutputDelta
	// StreamRetry retracts the failed model attempt's provisional output and
	// reasoning and announces a planned retry. It never retracts earlier model
	// invocations, persisted messages, or completed tool effects.
	StreamRetry
	// StreamRecovered ends retry status after a model invocation succeeds.
	// All event fields except Kind are zero. It does not alter accumulated text.
	StreamRecovered
)

// StreamEvent is transient output or retry status, not a persisted message.
// Concatenating events does not reconstruct Result or session history.
type StreamEvent struct {
	Kind      StreamEventKind
	TextDelta string

	// Retry fields are set only on StreamRetry; TextDelta is then empty.
	// Discard counts remove UTF-8 bytes from the end of the accumulated output
	// and reasoning respectively, always at rune boundaries. Counts cover only
	// the failed attempt and are zero when it emitted no text. Consumers must
	// apply the retraction before accepting further deltas, or return an error.
	// Attempt includes the original request and is at least 2. RetryDelay is the
	// nonnegative planned wait; cancellation can prevent the next attempt.
	DiscardOutputBytes    int
	DiscardReasoningBytes int
	Attempt               int
	RetryDelay            time.Duration
}

// StreamHandler receives events synchronously, in order, without concurrent
// calls within a turn. Returning an error aborts the turn immediately; the
// original error is returned and no later events or tools may be dispatched.
// Every invocation of the handler is externally observable stream progress,
// including an invocation that returns an error. Already completed side effects
// are not rolled back.
type StreamHandler func(StreamEvent) error

// StreamingRuntime independently exposes incremental agent output. It does not
// require Runtime. Implementations supporting both share turn execution and
// interceptors; only the model invocation and incremental observation differ.
// Turns for one session are serialized with Run and History.Load; different
// sessions may run concurrently. Input aggregates are immutable.
//
// Stream returns an Execution with a canonical, caller-owned Result only when
// its error is nil. Once the turn lifecycle starts, Outcome remains valid on
// failure or cancellation. The Result can differ from concatenated deltas.
// Reasoning need not be present and is not persisted merely because it was
// streamed. Retry controls are included; tool and turn lifecycle events are
// excluded. Implementations may
// complete successfully without delivering any StreamEvent.
//
// A nil handler returns ErrNilStreamHandler. Implementations may transparently
// use a non-streaming model path when streaming is unavailable, provided their
// execution contract permits it. Cancellation and deadlines stop future work
// without rolling back completed durable or external effects, and do not imply
// that retrying the turn is safe.
type StreamingRuntime interface {
	Stream(context.Context, Turn, StreamHandler) (Execution, error)
}
