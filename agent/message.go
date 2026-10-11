package agent

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/ingot-agent/sdk/content"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/session"
	"github.com/ingot-agent/sdk/tool"
)

// Input is a single user submission, including all of its attachments.
type Input struct {
	Text        string
	Attachments []InputAttachment
}

// InputAttachment describes a local file, media, or both. Path is communicated
// as an ordinary user file notice; only Media becomes model media content.
type InputAttachment struct {
	Name     string
	MIMEType string
	Size     int64
	Path     string
	Media    *content.Attachment
}

// MediaAttachments wraps existing media attachments as input attachments. It
// copies the media and inline bytes and does not invent local file paths.
func MediaAttachments(media []content.Attachment) []InputAttachment {
	owned := content.CloneAttachments(media)
	result := make([]InputAttachment, len(owned))
	for i := range owned {
		result[i] = InputAttachment{Name: owned[i].Media.Name, MIMEType: owned[i].Media.MIMEType, Media: &owned[i]}
	}
	return result
}

type MessageRequest struct {
	RootSessionID session.ID
	SessionID     session.ID
	Input         Input
	// RequestID reuses the caller's existing request receipt key. Empty disables
	// request deduplication. It never associates media with rounds or messages.
	RequestID string
}

var (
	ErrInvalidMessage  = errors.New("invalid agent message")
	ErrClosed          = errors.New("agent runtime closed")
	ErrRequestConflict = errors.New("requestId already belongs to a different message")
	ErrMessageNotFound = errors.New("agent message not found")
	ErrMessageStarted  = errors.New("message has been claimed; stop its execution instead")
	ErrSessionRunning  = errors.New("session is running")
)

func ValidateInput(input Input) error {
	if !utf8.ValidString(input.Text) {
		return ErrInvalidMessage
	}
	for i, a := range input.Attachments {
		if (a.Path == "" && a.Media == nil) || a.Size < 0 || !utf8.ValidString(a.Path) || !utf8.ValidString(a.Name) || !utf8.ValidString(a.MIMEType) {
			return fmt.Errorf("attachment %d: %w", i, ErrInvalidMessage)
		}
		if a.Media != nil {
			if err := content.ValidateAttachments([]content.Attachment{*a.Media}); err != nil {
				return fmt.Errorf("attachment %d: %w: %w", i, ErrInvalidMessage, err)
			}
			if a.Name != "" && a.Media.Media.Name != "" && a.Name != a.Media.Media.Name || a.MIMEType != "" && a.Media.Media.MIMEType != "" && a.MIMEType != a.Media.Media.MIMEType {
				return fmt.Errorf("attachment %d metadata mismatch: %w", i, ErrInvalidMessage)
			}
		}
	}
	return nil
}

func CloneInput(input Input) Input {
	result := Input{Text: input.Text, Attachments: append([]InputAttachment(nil), input.Attachments...)}
	for i := range result.Attachments {
		if a := result.Attachments[i].Media; a != nil {
			media := content.CloneAttachments([]content.Attachment{*a})[0]
			result.Attachments[i].Media = &media
		}
	}
	return result
}

func CloneTurn(turn Turn) Turn {
	result := turn
	result.Inputs = make([]Input, len(turn.Inputs))
	for i := range turn.Inputs {
		result.Inputs[i] = CloneInput(turn.Inputs[i])
	}
	return result
}

// MessageState uses the existing submission receipt identity, independent of a
// TurnID. A failed attachment submission can coexist with a successful turn.
type MessageState struct {
	ID        string
	Request   MessageRequest
	Sequence  uint64
	Revision  uint64
	Status    string
	Started   bool
	TurnID    string
	Error     string
	Persisted bool
	// HistoryEnd anchors a settled submission to the complete persisted history,
	// including failures that never produced a user message.
	HistoryEnd *int
}

// ExecutionBlock preserves input/round order without assigning block IDs.
// Round blocks contain provisional text and, after completion, canonical messages.
type ExecutionBlock struct {
	Kind       string // input or round
	InputIndex int
	RoundIndex int
	Messages   []model.Message
	Reasoning  string
	Output     string
	Tools      []ToolState
}

type ToolState struct {
	Call    tool.Call
	Status  string
	Content content.Content
	Error   string
}

type TurnState struct {
	TurnID     string
	Turn       Turn
	Status     string
	Blocks     []ExecutionBlock
	Execution  Execution
	Error      string
	HistoryEnd *int
}

type QueueState struct {
	SessionID   session.ID
	ActiveID    string // existing submission receipt that started this execution
	Paused      bool
	PauseReason string
	Revision    uint64
}

// RuntimeState is process-local, detached and caller-owned. Completed records
// have bounded retention; pending and active submissions are never evicted.
type RuntimeState struct {
	Revision uint64
	Messages []MessageState
	Turns    []TurnState
	Queues   []QueueState
}

// TurnProgress contains coalesced increments for an existing submission and
// round. MessageID is the submission receipt, not a new input/attachment ID.
type TurnProgress struct {
	SessionID      session.ID
	TurnID         string
	MessageID      string
	RoundIndex     int
	OutputDelta    string
	ReasoningDelta string
	Tools          []ToolProgress
}

type ToolProgress struct {
	ToolCallID string
	Content    content.Content
}

// RuntimeUpdate is one ordered, caller-owned update. A structural update
// upserts only the affected Messages, Turns and Queues. Progress appends to the
// referenced round instead. Both share RuntimeState.Revision: ignore revisions
// already covered by a newer State snapshot. Reset replaces the local runtime
// state when starting a watch or recovering beyond the bounded update window.
type RuntimeUpdate struct {
	RuntimeState
	Reset    bool
	Progress *TurnProgress
}

// Control queries authoritative runtime state and controls scheduling. Watch
// waits for the next update after a revision. State remains the full snapshot
// for bootstrap/recovery. Observer failures cannot affect execution.
type Control interface {
	State(context.Context) (RuntimeState, error)
	Watch(context.Context, uint64) (RuntimeUpdate, error)
	// CachedHistory returns the last complete persisted boundary without waiting
	// for the active execution. false means no boundary has been captured yet.
	CachedHistory(context.Context, session.ID) ([]model.Message, bool, error)
	SetPaused(context.Context, session.ID, bool) error
	CancelMessage(context.Context, string, bool) error
	Stop(context.Context, session.ID) error
	ForgetSession(context.Context, session.ID) error
}
