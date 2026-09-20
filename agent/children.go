package agent

import (
	"context"
	"errors"
	"time"

	"github.com/ingot-agent/sdk/execution"
	"github.com/ingot-agent/sdk/session"
	"github.com/ingot-agent/sdk/workspace"
)

const (
	// AgentMetaNamespace is the Session Meta namespace owned by agent
	// implementations that support child Sessions.
	AgentMetaNamespace = "agent"
	// ChildSessionKind identifies a Session created for one child-agent Turn.
	ChildSessionKind = "subagent"
	// ChildMetaSchemaVersion is the current agent Session Meta schema.
	ChildMetaSchemaVersion = 1
)

// ChildState is the persisted business state of one single-Turn child
// Session. Terminal states never transition back to Queued or Working.
type ChildState string

const (
	ChildQueued      ChildState = "queued"
	ChildWorking     ChildState = "working"
	ChildInterrupted ChildState = "interrupted"
	ChildCanceled    ChildState = "canceled"
	ChildFailed      ChildState = "failed"
	ChildCompleted   ChildState = "completed"
)

// AgentTypeInfo describes one preconfigured child-agent type available to the
// caller. Tool names and system prompts are intentionally not exposed here.
type AgentTypeInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ChildDefinition is the immutable configuration snapshot used by one child
// Session. A child keeps this snapshot even if later process configuration
// changes.
type ChildDefinition struct {
	SystemPrompt      string   `json:"system_prompt"`
	Tools             []string `json:"tools"`
	AllowedChildTypes []string `json:"allowed_child_types"`
}

// ChildDiagnostic is a stable, presentation-neutral diagnostic attached to a
// child state or returned by a management operation.
type ChildDiagnostic struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// ChildSessionMeta is the agent-owned value stored at Session Meta key
// "agent". Root Sessions may contain only ChildSessionIDs. Kind=subagent
// records contain the complete immutable task snapshot and mutable lifecycle
// fields below.
type ChildSessionMeta struct {
	SchemaVersion int    `json:"schema_version,omitempty"`
	Kind          string `json:"kind,omitempty"`

	ParentSessionID session.ID   `json:"parent_session_id,omitempty"`
	RootSessionID   session.ID   `json:"root_session_id,omitempty"`
	ChildSessionIDs []session.ID `json:"child_session_ids"`
	Depth           uint32       `json:"depth,omitempty"`

	AgentType        string          `json:"agent_type,omitempty"`
	DefinitionDigest string          `json:"definition_digest,omitempty"`
	Definition       ChildDefinition `json:"definition,omitempty"`
	Task             string          `json:"task,omitempty"`
	Context          string          `json:"context,omitempty"`

	Ready            bool             `json:"ready,omitempty"`
	State            ChildState       `json:"state,omitempty"`
	Result           *string          `json:"result"`
	Error            *ChildDiagnostic `json:"error"`
	Outcome          *Outcome         `json:"outcome"`
	PreviousState    *ChildState      `json:"previous_state"`
	InterruptReason  *string          `json:"interrupt_reason"`
	ExecutionStopped *bool            `json:"execution_stopped"`
	StartedAt        *time.Time       `json:"started_at"`
	FinishedAt       *time.Time       `json:"finished_at"`
	UpdatedAt        time.Time        `json:"updated_at,omitempty"`
}

// ChildRequest creates one new persistent child Session and schedules exactly
// one Agent Turn. Workspace, when present, binds an already existing local
// directory; child management never creates or removes worktrees.
type ChildRequest struct {
	AgentType string
	Task      string
	Context   string
	Workspace *workspace.Binding
	Wait      bool
}

// ChildSnapshot is the management view of one child Session. State is nil
// only when persistence could not be confirmed for the current operation.
// Result is present only for a confirmed Completed state requested by the
// caller.
type ChildSnapshot struct {
	SessionID        session.ID       `json:"session_id"`
	ParentSessionID  session.ID       `json:"parent_session_id,omitempty"`
	RootSessionID    session.ID       `json:"root_session_id,omitempty"`
	Depth            uint32           `json:"depth,omitempty"`
	AgentType        string           `json:"agent_type,omitempty"`
	State            *ChildState      `json:"state"`
	StateConfirmed   bool             `json:"state_confirmed"`
	ExecutionStopped *bool            `json:"execution_stopped"`
	HasResult        *bool            `json:"has_result,omitempty"`
	Result           *string          `json:"result,omitempty"`
	Reason           string           `json:"reason,omitempty"`
	Error            *ChildDiagnostic `json:"error,omitempty"`
	PreviousState    *ChildState      `json:"previous_state,omitempty"`
	StartedAt        *time.Time       `json:"started_at,omitempty"`
	FinishedAt       *time.Time       `json:"finished_at,omitempty"`
}

// ChildrenPageRequest selects one page of direct children. ParentSessionID is
// empty to list the caller's own direct children. Cursor is an opaque value
// returned by the previous page. PageSize zero selects the implementation
// default.
type ChildrenPageRequest struct {
	ParentSessionID session.ID
	Cursor          string
	PageSize        int
}

// ChildrenPage is one deterministic page ordered by Session creation time and
// stable identity.
type ChildrenPage struct {
	Children   []ChildSnapshot `json:"children"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

// CancelResult reports the persisted target state immediately after an
// explicit branch cancellation. ExecutionStopped may remain false while a
// running tool or model call responds to cancellation.
type CancelResult struct {
	Snapshot ChildSnapshot `json:"snapshot"`
	Changed  bool          `json:"changed"`
}

// Children manages single-Turn child Sessions. Implementations authorize every
// operation from the explicit execution Scope. Check never waits. Wait waits
// only for an execution object owned by the current process; when none exists
// it performs one persisted-state read and returns immediately.
type Children interface {
	Types(context.Context, execution.Scope) ([]AgentTypeInfo, error)
	CreateChild(context.Context, execution.Scope, ChildRequest) (ChildSnapshot, error)
	Check(context.Context, execution.Scope, session.ID, bool) (ChildSnapshot, error)
	Wait(context.Context, execution.Scope, session.ID) (ChildSnapshot, error)
	List(context.Context, execution.Scope, ChildrenPageRequest) (ChildrenPage, error)
	Cancel(context.Context, execution.Scope, session.ID) (CancelResult, error)
	SubmitResult(context.Context, execution.Scope, string, string) error
}

// ChildSessionRecord is the authoritative persisted agent metadata for one
// Session together with its base Session metadata. Agent.Kind may be empty for
// an ordinary root Session.
type ChildSessionRecord struct {
	Session session.Metadata
	Agent   ChildSessionMeta
}

// ChildSessionCreateRequest is consumed by a child-aware Session persistence
// provider. Parent relationship registration and child row creation commit in
// one transaction.
type ChildSessionCreateRequest struct {
	ParentSessionID session.ID
	Title           string
	Depth           uint32
	Agent           ChildSessionMeta
}

// ChildSessionPageRequest selects one persisted page of direct child records.
type ChildSessionPageRequest struct {
	Cursor   string
	PageSize int
}

// ChildSessionPage is the persistence result used by child management.
type ChildSessionPage struct {
	Records    []ChildSessionRecord
	NextCursor string
}

// UpdateValue distinguishes an omitted update from setting a field to its
// zero value.
type UpdateValue[T any] struct {
	Set   bool
	Value T
}

// UpdateNullable distinguishes an omitted update, clearing a nullable field,
// and assigning a non-nil value.
type UpdateNullable[T any] struct {
	Set   bool
	Value *T
}

// ChildSessionUpdate conditionally changes only the agent namespace. A false
// Updated result from ChildSessionRepository.Update means the current record
// did not satisfy ExpectedStates or ExpectedReady; it is not a storage error.
type ChildSessionUpdate struct {
	ExpectedStates []ChildState
	ExpectedReady  *bool

	Ready            UpdateValue[bool]
	State            UpdateValue[ChildState]
	Result           UpdateNullable[string]
	Error            UpdateNullable[ChildDiagnostic]
	Outcome          UpdateNullable[Outcome]
	PreviousState    UpdateNullable[ChildState]
	InterruptReason  UpdateNullable[string]
	ExecutionStopped UpdateNullable[bool]
	StartedAt        UpdateNullable[time.Time]
	FinishedAt       UpdateNullable[time.Time]
}

// ChildBranchMode selects the atomic state change applied to one branch.
type ChildBranchMode string

const (
	BranchCancel    ChildBranchMode = "cancel"
	BranchInterrupt ChildBranchMode = "interrupt"
)

// ChildBranchRequest changes a target and all descendants, or descendants
// only. Cancel changes only Queued/Working records. Interrupt changes every
// non-Completed record and preserves existing error and outcome facts.
type ChildBranchRequest struct {
	Mode            ChildBranchMode
	Reason          string
	DescendantsOnly bool
}

// ChildRecoveryRequest settles process-local states left by a previous
// runtime. Queued/Working become Interrupted and stale false stop facts become
// unknown where actual external writers cannot be confirmed.
type ChildRecoveryRequest struct {
	Reason string
}

// ChildSessionRepository is the optional child-aware extension implemented by
// a Session persistence provider. It stores data in Session Meta rather than a
// separate task database. All methods are safe for concurrent use.
type ChildSessionRepository interface {
	CreateChild(context.Context, ChildSessionCreateRequest) (ChildSessionRecord, error)
	GetChildSession(context.Context, session.ID) (ChildSessionRecord, error)
	ListChildSessions(context.Context, session.ID, ChildSessionPageRequest) (ChildSessionPage, error)
	UpdateChildSession(context.Context, session.ID, ChildSessionUpdate) (ChildSessionRecord, bool, error)
	UpdateChildBranch(context.Context, session.ID, ChildBranchRequest) ([]ChildSessionRecord, error)
	RecoverChildSessions(context.Context, ChildRecoveryRequest) error
}

var (
	// ErrChildUnsupported indicates that the selected Session provider or Agent
	// composition does not support child Sessions.
	ErrChildUnsupported = errors.New("child agents unsupported")
	// ErrChildUnauthorized indicates that the caller is not the target or one
	// of its ancestors, or is not inside a current Agent execution.
	ErrChildUnauthorized = errors.New("child agent operation unauthorized")
	// ErrChildInvalidState indicates that the requested single-Turn transition
	// is not valid for the persisted child state.
	ErrChildInvalidState = errors.New("invalid child agent state")
	// ErrChildCapacity indicates that the bounded queue or active execution
	// capacity cannot accept another child.
	ErrChildCapacity = errors.New("child agent capacity exhausted")
	// ErrChildSubmission indicates an invalid, duplicate, mixed, or missing
	// final result submission.
	ErrChildSubmission = errors.New("invalid child agent result submission")
)
