package interaction

import "testing"

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
