// Package interaction defines the structured effect boundary between plugins
// and their host environment.
package interaction

import (
	"context"
	"errors"

	"github.com/ingot-agent/sdk/execution"
)

// Channel lets a plugin request values from, emit events to, and publish
// current state to its host environment. Implementations decide whether those
// effects are handled by a UI, CLI, policy, configuration, recorder, remote
// controller, or another host facility.
type Channel interface {
	Request(context.Context, Request) (Response, error)
	Emit(context.Context, Event) error
	Set(context.Context, State) error
	Clear(context.Context, string) error
}

// ExecutionBinder derives an interaction Channel with an immutable binding to
// one explicit dynamic execution scope. It is the execution-side capability
// for plugins whose interaction effects must be routed to the Session that owns
// a runtime invocation.
//
// Bind is concurrent-safe and does not perform a blocking operation. It treats
// scope as immutable and must return an error wrapping ErrInvalidExecutionScope
// when scope has no SessionID. The returned Channel is concurrent-safe and its
// binding cannot change. Its business routing is determined by the bound scope,
// never by context values; implementations may use context-carried tracing or
// observation correlation only as optional metadata that cannot replace or
// override the explicit scope or change the operation's business result.
type ExecutionBinder interface {
	Bind(execution.Scope) (Channel, error)
}

// Unavailable returns a concurrent-safe Channel with no host facility behind
// it. Each method returns an already-canceled context error when present and
// otherwise returns ErrUnavailable. It is suitable for non-interactive hosts
// that must supply an explicit Channel.
func Unavailable() Channel { return unavailableChannel{} }

// UnavailableBinder returns a concurrent-safe ExecutionBinder for a host with
// no interaction facility. Bind rejects a missing SessionID with
// ErrInvalidExecutionScope; otherwise it returns an Unavailable Channel so the
// eventual effect call preserves context cancellation and deadline errors.
func UnavailableBinder() ExecutionBinder { return unavailableExecutionBinder{} }

// Level describes semantic severity without prescribing host behavior.
type Level uint8

const (
	// LevelUnspecified indicates that the producer expressed no severity.
	LevelUnspecified Level = iota
	// LevelInfo marks an informational interaction.
	LevelInfo
	// LevelWarning marks a warning interaction.
	LevelWarning
	// LevelError marks an error-severity interaction.
	LevelError
)

// Request asks the host environment to provide a set of named values. Name is
// a stable protocol identity, not a per-call correlation ID.
type Request struct {
	Name        string
	Level       Level
	Description string
	Fields      []Field
}

// FieldKind identifies the value contract of a request field.
type FieldKind uint8

const (
	FieldString FieldKind = iota + 1
	FieldInteger
	FieldNumber
	FieldBoolean
	FieldChoice
	FieldMultiChoice
	// FieldObject requests a nested group of named fields. The group's
	// members are described by Field.Fields; the collected answer is a
	// ValueObject.
	FieldObject
	// FieldList requests an ordered, possibly empty repetition of one
	// element descriptor. The element contract is described by
	// Field.Element; the collected answer is a ValueList. A list element may
	// itself be an object or another list, which is how repeated nested
	// structures are expressed without inventing per-plugin protocols.
	FieldList
)

// Field describes one value requested from the host. Name is machine-facing;
// Label and Description are human-facing metadata. Options on FieldString are
// ordered suggestions and do not restrict free-form answers. Options on
// FieldChoice and FieldMultiChoice define the allowed values.
//
// Fields is set exactly for FieldObject and describes the group's members in
// presentation order. Element is set exactly for FieldList and describes one
// repeated element. Both may nest arbitrarily deep; a host must therefore
// render and validate recursively rather than assuming a flat form.
type Field struct {
	Name        string
	Label       string
	Description string

	Kind      FieldKind
	Required  bool
	Sensitive bool

	Default *Value
	Options []Option

	Fields  []Field
	Element *Field
}

// Option describes one ordered choice. Value is the stable protocol value;
// Label and Description are human-facing metadata.
type Option struct {
	Value       string
	Label       string
	Description string
}

// Response contains the values supplied by the host for a Request.
type Response struct {
	Values []Answer
}

// Answer supplies one field value by its machine-facing field name.
type Answer struct {
	Name  string
	Value Value
}

// Event reports that a named fact just occurred. Events have no continuing
// lifecycle; durable current facts belong in State.
type Event struct {
	Name    string
	Level   Level
	Message string
}

// State replaces the current snapshot of a named state object. Name is unique
// within the producing component instance; the runtime may add component
// identity when constructing a global key.
type State struct {
	Name        string
	Level       Level
	Description string
	Values      []Entry
}

// Entry describes one value in a State snapshot.
type Entry struct {
	Name        string
	Label       string
	Description string
	Value       Value
}

// ValueKind identifies one interaction value representation.
type ValueKind uint8

const (
	ValueString ValueKind = iota + 1
	ValueInteger
	ValueNumber
	ValueBoolean
	ValueStrings
	// ValueObject carries one nested group value. Entries holds the group
	// members and is authoritative; the scalar fields are ignored.
	ValueObject
	// ValueList carries one ordered repetition. Items holds the elements and
	// is authoritative; the scalar fields are ignored.
	ValueList
)

// Value carries one interaction value. The field selected by Kind is
// authoritative; all other representation fields are ignored.
//
// ValueObject and ValueList nest recursively, so a value tree mirrors the
// requesting Field tree one-to-one.
type Value struct {
	Kind ValueKind

	String  string
	Integer int64
	Number  float64
	Boolean bool
	Strings []string

	Entries []Entry
	Items   []Value
}

// ObjectValue constructs a ValueObject value. Entries are copied so
// subsequent caller mutation cannot change the returned Value.
func ObjectValue(entries []Entry) Value {
	return Value{Kind: ValueObject, Entries: append([]Entry(nil), entries...)}
}

// ListValue constructs a ValueList value. Items are copied so subsequent
// caller mutation cannot change the returned Value.
func ListValue(items []Value) Value {
	return Value{Kind: ValueList, Items: append([]Value(nil), items...)}
}

// StringValue constructs a string Value.
func StringValue(value string) Value {
	return Value{Kind: ValueString, String: value}
}

// IntegerValue constructs an integer Value.
func IntegerValue(value int64) Value {
	return Value{Kind: ValueInteger, Integer: value}
}

// NumberValue constructs a floating-point Value.
func NumberValue(value float64) Value {
	return Value{Kind: ValueNumber, Number: value}
}

// BooleanValue constructs a boolean Value.
func BooleanValue(value bool) Value {
	return Value{Kind: ValueBoolean, Boolean: value}
}

// StringsValue constructs a string-list Value and copies value so subsequent
// caller mutation cannot change the returned Value.
func StringsValue(value []string) Value {
	return Value{Kind: ValueStrings, Strings: append([]string(nil), value...)}
}

var (
	// ErrUnavailable indicates that no host facility can service an interaction.
	ErrUnavailable = errors.New("interaction unavailable")
	// ErrInvalidExecutionScope indicates that an interaction Channel could not
	// be bound because its explicit dynamic execution scope is invalid.
	ErrInvalidExecutionScope = errors.New("invalid interaction execution scope")
)

type unavailableChannel struct{}

type unavailableExecutionBinder struct{}

var _ Channel = unavailableChannel{}
var _ ExecutionBinder = unavailableExecutionBinder{}

func (unavailableExecutionBinder) Bind(scope execution.Scope) (Channel, error) {
	if scope.SessionID == "" {
		return nil, ErrInvalidExecutionScope
	}
	return unavailableChannel{}, nil
}

func (unavailableChannel) Request(ctx context.Context, _ Request) (Response, error) {
	return Response{}, unavailableError(ctx)
}

func (unavailableChannel) Emit(ctx context.Context, _ Event) error {
	return unavailableError(ctx)
}

func (unavailableChannel) Set(ctx context.Context, _ State) error {
	return unavailableError(ctx)
}

func (unavailableChannel) Clear(ctx context.Context, _ string) error {
	return unavailableError(ctx)
}

func unavailableError(ctx context.Context) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return ErrUnavailable
}
