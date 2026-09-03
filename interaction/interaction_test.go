package interaction

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestValueConstructors(t *testing.T) {
	strings := []string{"one", "two"}
	tests := []struct {
		name  string
		value Value
		kind  ValueKind
	}{
		{name: "string", value: StringValue("value"), kind: ValueString},
		{name: "integer", value: IntegerValue(42), kind: ValueInteger},
		{name: "number", value: NumberValue(0.5), kind: ValueNumber},
		{name: "boolean", value: BooleanValue(true), kind: ValueBoolean},
		{name: "strings", value: StringsValue(strings), kind: ValueStrings},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.value.Kind != test.kind {
				t.Fatalf("kind=%v want=%v", test.value.Kind, test.kind)
			}
		})
	}

	strings[0] = "changed"
	if tests[4].value.Strings[0] != "one" {
		t.Fatalf("StringsValue retained caller slice: %#v", tests[4].value.Strings)
	}
}

func TestUnavailableChannel(t *testing.T) {
	channel := Unavailable()
	operations := []struct {
		name string
		call func(context.Context) error
	}{
		{name: "request", call: func(ctx context.Context) error {
			_, err := channel.Request(ctx, Request{})
			return err
		}},
		{name: "emit", call: func(ctx context.Context) error { return channel.Emit(ctx, Event{}) }},
		{name: "set", call: func(ctx context.Context) error { return channel.Set(ctx, State{}) }},
		{name: "clear", call: func(ctx context.Context) error { return channel.Clear(ctx, "state") }},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			if err := operation.call(context.Background()); !errors.Is(err, ErrUnavailable) {
				t.Fatalf("error = %v, want ErrUnavailable", err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := operation.call(ctx); !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled error = %v, want context.Canceled", err)
			}
		})
	}
}

func TestUnavailableChannelConcurrentUse(t *testing.T) {
	channel := Unavailable()
	var calls sync.WaitGroup
	for range 32 {
		calls.Add(1)
		go func() {
			defer calls.Done()
			if err := channel.Emit(context.Background(), Event{}); !errors.Is(err, ErrUnavailable) {
				t.Errorf("Emit error = %v, want ErrUnavailable", err)
			}
		}()
	}
	calls.Wait()
}
