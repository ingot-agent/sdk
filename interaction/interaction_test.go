package interaction

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/ingot-agent/sdk/execution"
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

func TestNestedValueConstructorsCopyInputs(t *testing.T) {
	entries := []Entry{{Name: "host", Value: StringValue("example.com")}}
	object := ObjectValue(entries)
	if object.Kind != ValueObject || len(object.Entries) != 1 {
		t.Fatalf("object value = %#v", object)
	}
	entries[0] = Entry{Name: "changed"}
	if object.Entries[0].Name != "host" {
		t.Fatalf("ObjectValue retained caller slice: %#v", object.Entries)
	}

	items := []Value{StringValue("first")}
	list := ListValue(items)
	if list.Kind != ValueList || len(list.Items) != 1 {
		t.Fatalf("list value = %#v", list)
	}
	items[0] = StringValue("changed")
	if list.Items[0].String != "first" {
		t.Fatalf("ListValue retained caller slice: %#v", list.Items)
	}
}

// TestNestedFieldModelSupportsRepeatedObjects models the shape official
// plugins actually need: a repeated group whose members are themselves
// grouped, for example a provider list with per-provider model entries.
func TestNestedFieldModelSupportsRepeatedObjects(t *testing.T) {
	provider := Field{
		Name: "provider", Kind: FieldObject, Required: true,
		Fields: []Field{
			{Name: "name", Kind: FieldString, Required: true},
			{Name: "models", Kind: FieldList, Element: &Field{Name: "model", Kind: FieldString}},
			{Name: "auth", Kind: FieldObject, Fields: []Field{
				{Name: "token", Kind: FieldString, Sensitive: true},
			}},
		},
	}
	request := Request{Name: "example.setup", Fields: []Field{{Name: "providers", Kind: FieldList, Element: &provider}}}
	if request.Fields[0].Element.Kind != FieldObject || len(request.Fields[0].Element.Fields) != 3 {
		t.Fatalf("repeated object descriptor = %#v", request.Fields[0].Element)
	}
	value := ListValue([]Value{ObjectValue([]Entry{
		{Name: "name", Value: StringValue("openai")},
		{Name: "models", Value: ListValue([]Value{StringValue("gpt-4o-mini")})},
		{Name: "auth", Value: ObjectValue([]Entry{{Name: "token", Value: StringValue("secret")}})},
	})})
	if value.Kind != ValueList || value.Items[0].Kind != ValueObject {
		t.Fatalf("nested value tree = %#v", value)
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

func TestUnavailableExecutionBinder(t *testing.T) {
	binder := UnavailableBinder()
	if _, err := binder.Bind(execution.Scope{}); !errors.Is(err, ErrInvalidExecutionScope) {
		t.Fatalf("empty scope error = %v, want ErrInvalidExecutionScope", err)
	}
	channel, err := binder.Bind(execution.Scope{SessionID: "session"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := channel.Request(context.Background(), Request{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("request error = %v, want ErrUnavailable", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := channel.Emit(ctx, Event{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled error = %v, want context.Canceled", err)
	}
}
