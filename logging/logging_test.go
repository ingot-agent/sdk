package logging_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ingot-agent/sdk/logging"
)

type captureLogger struct {
	level   logging.Level
	records []captureRecord
	attrs   []logging.Attr
}

type captureRecord struct {
	level   logging.Level
	message string
	attrs   []logging.Attr
}

func (c *captureLogger) Enabled(level logging.Level) bool { return level >= c.level }

func (c *captureLogger) Debug(ctx context.Context, message string, attrs ...logging.Attr) {
	c.record(ctx, logging.LevelDebug, message, attrs)
}

func (c *captureLogger) Info(ctx context.Context, message string, attrs ...logging.Attr) {
	c.record(ctx, logging.LevelInfo, message, attrs)
}

func (c *captureLogger) Warn(ctx context.Context, message string, attrs ...logging.Attr) {
	c.record(ctx, logging.LevelWarn, message, attrs)
}

func (c *captureLogger) Error(ctx context.Context, message string, attrs ...logging.Attr) {
	c.record(ctx, logging.LevelError, message, attrs)
}

func (c *captureLogger) With(attrs ...logging.Attr) logging.Logger {
	clone := *c
	clone.attrs = append(append([]logging.Attr(nil), c.attrs...), attrs...)
	return &clone
}

// record models the required producer-side guard: a disabled level is never
// materialized.
func (c *captureLogger) record(_ context.Context, level logging.Level, message string, attrs []logging.Attr) {
	if !c.Enabled(level) {
		return
	}
	retained := make([]logging.Attr, 0, len(c.attrs)+len(attrs))
	retained = append(retained, c.attrs...)
	retained = append(retained, attrs...)
	c.records = append(c.records, captureRecord{level: level, message: message, attrs: retained})
}

var _ logging.Logger = (*captureLogger)(nil)

func TestParseLevel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  logging.Level
	}{
		{name: "debug", input: "debug", want: logging.LevelDebug},
		{name: "info", input: "info", want: logging.LevelInfo},
		{name: "warn", input: "warn", want: logging.LevelWarn},
		{name: "warning alias", input: "warning", want: logging.LevelWarn},
		{name: "error", input: "error", want: logging.LevelError},
		{name: "mixed case", input: "WaRn", want: logging.LevelWarn},
		{name: "surrounding space", input: "  debug\t", want: logging.LevelDebug},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := logging.ParseLevel(test.input)
			if err != nil {
				t.Fatalf("ParseLevel(%q) error = %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("ParseLevel(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}

func TestParseLevelRejectsUnknownNames(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"", " ", "trace", "fatal", "level(0)", "INFO4"} {
		if _, err := logging.ParseLevel(input); !errors.Is(err, logging.ErrInvalidLevel) {
			t.Fatalf("ParseLevel(%q) error = %v, want ErrInvalidLevel", input, err)
		}
	}
}

func TestLevelsAreOrderedBySeverity(t *testing.T) {
	t.Parallel()
	levels := []logging.Level{logging.LevelDebug, logging.LevelInfo, logging.LevelWarn, logging.LevelError}
	for index := 1; index < len(levels); index++ {
		if !(levels[index-1] < levels[index]) {
			t.Fatalf("%v is not less severe than %v", levels[index-1], levels[index])
		}
	}
}

func TestLevelString(t *testing.T) {
	t.Parallel()
	tests := map[logging.Level]string{
		logging.LevelDebug: "debug",
		logging.LevelInfo:  "info",
		logging.LevelWarn:  "warn",
		logging.LevelError: "error",
		logging.Level(3):   "level(3)",
	}
	for level, want := range tests {
		if got := level.String(); got != want {
			t.Fatalf("Level(%d).String() = %q, want %q", int(level), got, want)
		}
	}
}

func TestLevelNamesRoundTrip(t *testing.T) {
	t.Parallel()
	for _, level := range []logging.Level{logging.LevelDebug, logging.LevelInfo, logging.LevelWarn, logging.LevelError} {
		parsed, err := logging.ParseLevel(level.String())
		if err != nil {
			t.Fatalf("ParseLevel(%q) error = %v", level.String(), err)
		}
		if parsed != level {
			t.Fatalf("round trip of %v produced %v", level, parsed)
		}
	}
}

func TestValueConstructorsSelectKind(t *testing.T) {
	t.Parallel()
	at := time.Date(2026, time.November, 20, 12, 0, 0, 0, time.UTC)
	cause := errors.New("cause")
	tests := []struct {
		name  string
		value logging.Value
		kind  logging.Kind
	}{
		{name: "string", value: logging.StringValue("text"), kind: logging.KindString},
		{name: "int", value: logging.IntValue(-1), kind: logging.KindInt64},
		{name: "uint", value: logging.UintValue(1), kind: logging.KindUint64},
		{name: "float", value: logging.FloatValue(1.5), kind: logging.KindFloat64},
		{name: "bool", value: logging.BoolValue(true), kind: logging.KindBool},
		{name: "duration", value: logging.DurationValue(time.Second), kind: logging.KindDuration},
		{name: "time", value: logging.TimeValue(at), kind: logging.KindTime},
		{name: "error", value: logging.ErrorValue(cause), kind: logging.KindError},
		{name: "group", value: logging.GroupValue(logging.String("k", "v")), kind: logging.KindGroup},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.value.Kind != test.kind {
				t.Fatalf("Kind = %v, want %v", test.value.Kind, test.kind)
			}
		})
	}
}

func TestErrorValueKeepsNilSafe(t *testing.T) {
	t.Parallel()
	value := logging.ErrorValue(nil)
	if value.Kind != logging.KindError {
		t.Fatalf("Kind = %v, want KindError", value.Kind)
	}
	if value.Error != nil {
		t.Fatalf("Error = %v, want nil", value.Error)
	}
}

func TestGroupValueCopiesEntries(t *testing.T) {
	t.Parallel()
	entries := []logging.Attr{logging.String("first", "one")}
	value := logging.GroupValue(entries...)
	entries[0] = logging.String("mutated", "two")
	if value.Group[0].Key != "first" {
		t.Fatalf("group entry changed through caller slice: %q", value.Group[0].Key)
	}
}

func TestGroupCopiesAttributes(t *testing.T) {
	t.Parallel()
	attrs := []logging.Attr{logging.Int("count", 1)}
	attr := logging.Group("nested", attrs...)
	attrs[0] = logging.Int("count", 2)
	if attr.Key != "nested" || attr.Value.Kind != logging.KindGroup {
		t.Fatalf("unexpected group attribute: %+v", attr)
	}
	if attr.Value.Group[0].Value.Int64 != 1 {
		t.Fatalf("group value changed through caller slice: %d", attr.Value.Group[0].Value.Int64)
	}
}

func TestErrUsesErrorKey(t *testing.T) {
	t.Parallel()
	attr := logging.Err(errors.New("failed"))
	if attr.Key != "error" {
		t.Fatalf("Key = %q, want error", attr.Key)
	}
	if attr.Value.Kind != logging.KindError {
		t.Fatalf("Kind = %v, want KindError", attr.Value.Kind)
	}
}

func TestAttributeConstructorsSetKeyAndKind(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		attr logging.Attr
		kind logging.Kind
	}{
		{name: "string", attr: logging.String("k", "v"), kind: logging.KindString},
		{name: "int", attr: logging.Int("k", 1), kind: logging.KindInt64},
		{name: "uint", attr: logging.Uint("k", 1), kind: logging.KindUint64},
		{name: "float", attr: logging.Float("k", 1), kind: logging.KindFloat64},
		{name: "bool", attr: logging.Bool("k", true), kind: logging.KindBool},
		{name: "duration", attr: logging.Duration("k", time.Second), kind: logging.KindDuration},
		{name: "time", attr: logging.Time("k", time.Now()), kind: logging.KindTime},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.attr.Key != "k" {
				t.Fatalf("Key = %q, want k", test.attr.Key)
			}
			if test.attr.Value.Kind != test.kind {
				t.Fatalf("Kind = %v, want %v", test.attr.Value.Kind, test.kind)
			}
		})
	}
}

func TestDiscardEmitsNothing(t *testing.T) {
	t.Parallel()
	logger := logging.Discard()
	for _, level := range []logging.Level{logging.LevelDebug, logging.LevelInfo, logging.LevelWarn, logging.LevelError} {
		if logger.Enabled(level) {
			t.Fatalf("Discard().Enabled(%v) = true", level)
		}
	}
	ctx := context.Background()
	logger.Debug(ctx, "debug")
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
	if got := logger.With(logging.String("k", "v")); got.Enabled(logging.LevelError) {
		t.Fatalf("With on Discard enabled a level")
	}
}

func TestProducerGuardSkipsDisabledLevels(t *testing.T) {
	t.Parallel()
	logger := &captureLogger{level: logging.LevelWarn}
	if logger.Enabled(logging.LevelInfo) {
		t.Fatal("Enabled(LevelInfo) = true below the minimum level")
	}
	logger.Info(context.Background(), "hidden")
	logger.Warn(context.Background(), "visible")
	if len(logger.records) != 1 {
		t.Fatalf("records = %d, want 1", len(logger.records))
	}
	if logger.records[0].level != logging.LevelWarn {
		t.Fatalf("level = %v, want LevelWarn", logger.records[0].level)
	}
}

func TestWithRetainsAttributes(t *testing.T) {
	t.Parallel()
	logger := (&captureLogger{level: logging.LevelDebug}).With(logging.String("component", "test"))
	logger.Info(context.Background(), "message", logging.Int("count", 2))
	records := logger.(*captureLogger).records
	if len(records) != 1 {
		t.Fatalf("records = %d, want 1", len(records))
	}
	if len(records[0].attrs) != 2 {
		t.Fatalf("attrs = %v, want retained and record attributes", records[0].attrs)
	}
}
