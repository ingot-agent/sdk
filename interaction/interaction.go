// Package interaction defines the structured effect boundary between plugins
// and their host environment.
package interaction

import (
	"context"
	"errors"
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

// FieldKind identifies the primitive value contract of a request field.
type FieldKind uint8

const (
	FieldString FieldKind = iota + 1
	FieldInteger
	FieldNumber
	FieldBoolean
	FieldChoice
	FieldMultiChoice
)

// Field describes one value requested from the host. Name is machine-facing;
// Label and Description are human-facing metadata. Options on FieldString are
// ordered suggestions and do not restrict free-form answers. Options on
// FieldChoice and FieldMultiChoice define the allowed values.
type Field struct {
	Name        string
	Label       string
	Description string

	Kind      FieldKind
	Required  bool
	Sensitive bool

	Default *Value
	Options []Option
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

// ValueKind identifies one primitive interaction value representation.
type ValueKind uint8

const (
	ValueString ValueKind = iota + 1
	ValueInteger
	ValueNumber
	ValueBoolean
	ValueStrings
)

// Value carries one primitive interaction value. The field selected by Kind
// is authoritative; all other representation fields are ignored.
type Value struct {
	Kind ValueKind

	String  string
	Integer int64
	Number  float64
	Boolean bool
	Strings []string
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

// ErrUnavailable indicates that no host facility can service an interaction.
var ErrUnavailable = errors.New("interaction unavailable")
