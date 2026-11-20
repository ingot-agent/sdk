// Package logging defines leveled, structured log emission contracts shared by
// runtime components and interchangeable logging implementations.
//
// A Logger accepts records from any component that produces diagnostics. The
// contract deliberately says nothing about formatting, destinations, retention
// or rotation: those are implementation policy selected by the operator. It
// does fix the vocabulary (levels, attributes and context correlation) so that
// producers and implementations can be exchanged independently.
//
// Logging is observability, not a business effect. A Logger must never be used
// to report data that a caller is expected to act on, and implementations must
// never let a record change the outcome of the operation that emitted it.
package logging

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Level is the semantic severity of a record. Larger values are more severe,
// so an implementation that keeps records at or above a minimum level compares
// levels numerically. Debug, Info, Warn and Error are the portable levels; a
// producer should not emit other values unless an implementation documents
// support for them.
type Level int8

const (
	// LevelDebug marks detailed diagnostics that are normally disabled.
	LevelDebug Level = -4
	// LevelInfo marks routine operational progress.
	LevelInfo Level = 0
	// LevelWarn marks a recoverable problem or degraded behavior.
	LevelWarn Level = 4
	// LevelError marks a failed operation or an unrecoverable problem.
	LevelError Level = 8
)

// ErrInvalidLevel indicates that a level name is not recognized.
var ErrInvalidLevel = errors.New("invalid log level")

// ParseLevel returns the Level named by name. Names are case-insensitive and
// surrounding space is ignored; "debug", "info", "warn" and "error" are
// recognized. The zero value and unknown names return ErrInvalidLevel so that
// configuration mistakes are reported instead of silently selecting a level.
func ParseLevel(name string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return 0, fmt.Errorf("%q: %w", name, ErrInvalidLevel)
	}
}

// String returns the lowercase portable name of level, or a numeric fallback
// for a value outside the portable set.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return fmt.Sprintf("level(%d)", int(l))
	}
}

// Kind identifies the value contract carried by a Value.
type Kind uint8

const (
	// KindString marks a UTF-8 text value.
	KindString Kind = iota + 1
	// KindInt64 marks a signed integer value.
	KindInt64
	// KindUint64 marks an unsigned integer value.
	KindUint64
	// KindFloat64 marks a floating-point value.
	KindFloat64
	// KindBool marks a boolean value.
	KindBool
	// KindDuration marks a time interval.
	KindDuration
	// KindTime marks an instant.
	KindTime
	// KindError marks an error value.
	KindError
	// KindGroup marks an ordered, nested collection of attributes.
	KindGroup
)

// Value is one immutable attribute value. Only the field selected by Kind is
// meaningful. Implementations render values they understand and may render an
// unrecognized Kind as its numeric or textual fallback; they must not retain
// pointers into the value.
type Value struct {
	Kind Kind

	String   string
	Int64    int64
	Uint64   uint64
	Float64  float64
	Boolean  bool
	Duration time.Duration
	Time     time.Time
	Error    error
	Group    []Attr
}

// StringValue constructs a string Value.
func StringValue(value string) Value { return Value{Kind: KindString, String: value} }

// IntValue constructs a signed integer Value.
func IntValue(value int64) Value { return Value{Kind: KindInt64, Int64: value} }

// UintValue constructs an unsigned integer Value.
func UintValue(value uint64) Value { return Value{Kind: KindUint64, Uint64: value} }

// FloatValue constructs a floating-point Value.
func FloatValue(value float64) Value { return Value{Kind: KindFloat64, Float64: value} }

// BoolValue constructs a boolean Value.
func BoolValue(value bool) Value { return Value{Kind: KindBool, Boolean: value} }

// DurationValue constructs a duration Value.
func DurationValue(value time.Duration) Value { return Value{Kind: KindDuration, Duration: value} }

// TimeValue constructs an instant Value. The timezone of value is preserved.
func TimeValue(value time.Time) Value { return Value{Kind: KindTime, Time: value} }

// ErrorValue constructs an error Value. A nil error is valid and renders as an
// empty string, so an implementation never dereferences a nil error.
func ErrorValue(value error) Value { return Value{Kind: KindError, Error: value} }

// GroupValue constructs a nested group Value. Entries are copied so later
// caller mutation cannot change the returned Value.
func GroupValue(entries ...Attr) Value {
	return Value{Kind: KindGroup, Group: append([]Attr(nil), entries...)}
}

// Attr is one named attribute of a record. Key is the machine-facing identity
// of the attribute; producers should use stable, lowercase, dot-separated
// names. An attribute with an empty Key is ignored by implementations.
type Attr struct {
	Key   string
	Value Value
}

// String returns a string attribute.
func String(key, value string) Attr { return Attr{Key: key, Value: StringValue(value)} }

// Int returns a signed integer attribute.
func Int(key string, value int64) Attr { return Attr{Key: key, Value: IntValue(value)} }

// Uint returns an unsigned integer attribute.
func Uint(key string, value uint64) Attr { return Attr{Key: key, Value: UintValue(value)} }

// Float returns a floating-point attribute.
func Float(key string, value float64) Attr { return Attr{Key: key, Value: FloatValue(value)} }

// Bool returns a boolean attribute.
func Bool(key string, value bool) Attr { return Attr{Key: key, Value: BoolValue(value)} }

// Duration returns a duration attribute.
func Duration(key string, value time.Duration) Attr {
	return Attr{Key: key, Value: DurationValue(value)}
}

// Time returns an instant attribute.
func Time(key string, value time.Time) Attr { return Attr{Key: key, Value: TimeValue(value)} }

// Group returns an attribute whose value is the ordered nested group of attrs.
// The attributes are copied so later caller mutation cannot change the result.
func Group(key string, attrs ...Attr) Attr {
	return Attr{Key: key, Value: GroupValue(attrs...)}
}

// Err returns an error attribute named "error". A nil error is accepted and
// produces an empty string value; callers that need a different key compose
// Attr{Key: key, Value: ErrorValue(err)} directly.
func Err(err error) Attr { return Attr{Key: "error", Value: ErrorValue(err)} }

// Logger accepts leveled, structured records from the component that owns the
// diagnostic. Implementations are safe for concurrent use by multiple
// goroutines and must not panic for any input, including a nil ctx.
//
// A record is emitted only when Enabled reports true for its level. Producers
// should therefore guard expensive attribute construction with Enabled.
//
// Implementations must attach the execution identity carried by ctx: when
// observation.CorrelationFromContext reports a correlation, the record carries
// session_id, and, when present, turn_id, round and tool_call_id. A nil or
// background context is valid and carries no correlation. Producers therefore
// pass the operation's context and never repeat correlation attributes
// themselves.
//
// Implementations must not retain the message, the context or the attributes
// after a call returns, and must not drop a record only because ctx is done:
// cancellation may end an implementation's internal wait for output, but the
// record has already been accepted. Implementations own formatting, filtering,
// destinations, retention and any buffering; a buffered implementation must
// flush or release its buffer from the Component Cleanup that owns it.
type Logger interface {
	// Enabled reports whether a record at level would be emitted. It is a
	// cheap, best-effort check: the effective level may change concurrently.
	Enabled(level Level) bool
	// Debug emits a record at LevelDebug.
	Debug(ctx context.Context, message string, attrs ...Attr)
	// Info emits a record at LevelInfo.
	Info(ctx context.Context, message string, attrs ...Attr)
	// Warn emits a record at LevelWarn.
	Warn(ctx context.Context, message string, attrs ...Attr)
	// Error emits a record at LevelError.
	Error(ctx context.Context, message string, attrs ...Attr)
	// With returns a Logger that adds attrs to every subsequent record. The
	// returned Logger shares the implementation's destination and level
	// policy. Attributes are copied, so later caller mutation cannot change
	// the returned Logger. A later record whose own attribute repeats a key
	// takes precedence over the retained value.
	With(attrs ...Attr) Logger
}

// Discard returns a Logger that emits nothing. It is the explicit choice for a
// component whose logging is not configured, not a substitute for a
// destination.
func Discard() Logger { return discardLogger{} }

type discardLogger struct{}

func (discardLogger) Enabled(Level) bool                     { return false }
func (discardLogger) Debug(context.Context, string, ...Attr) {}
func (discardLogger) Info(context.Context, string, ...Attr)  {}
func (discardLogger) Warn(context.Context, string, ...Attr)  {}
func (discardLogger) Error(context.Context, string, ...Attr) {}
func (d discardLogger) With(...Attr) Logger                  { return d }
