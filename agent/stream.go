package agent

// StreamEventKind identifies a transient agent output delta.
type StreamEventKind uint8

const (
	// StreamReasoningDelta is reasoning/thinking text explicitly returned by
	// the model provider. Agents must not infer reasoning from ordinary output.
	StreamReasoningDelta StreamEventKind = iota + 1
	// StreamOutputDelta is user-visible assistant text from any model round,
	// including rounds that subsequently invoke tools.
	StreamOutputDelta
)

// StreamEvent is transient text for live consumption, not a persisted message.
// Concatenating events does not reconstruct Result or session history.
type StreamEvent struct {
	Kind      StreamEventKind
	TextDelta string
}

// StreamHandler receives events synchronously, in order, without concurrent
// calls within a turn. Returning an error aborts the turn immediately; the
// original error is returned and no later events or tools may be dispatched.
// Every invocation of the handler is externally observable stream progress,
// including an invocation that returns an error. Already completed side effects
// are not rolled back.
type StreamHandler func(StreamEvent) error
