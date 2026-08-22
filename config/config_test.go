package config_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ingot-agent/sdk/config"
)

type nestedConfig struct {
	Enabled bool   `toml:"enabled"`
	Label   string `toml:"label"`
}

type pluginConfig struct {
	Endpoint       string       `toml:"endpoint"`
	TimeoutSeconds int          `toml:"timeout_seconds"`
	Nested         nestedConfig `toml:"nested"`
}

func TestDecodeStrictTOML(t *testing.T) {
	t.Parallel()

	got, err := config.Decode[pluginConfig]([]byte(`
endpoint = "https://example.test/v1"
timeout_seconds = 3

[nested]
enabled = true
label = "test"
`))
	if err != nil {
		t.Fatal(err)
	}
	if got.Endpoint != "https://example.test/v1" || got.TimeoutSeconds != 3 {
		t.Fatalf("decoded config = %#v", got)
	}
	if !got.Nested.Enabled || got.Nested.Label != "test" {
		t.Fatalf("decoded nested config = %#v", got.Nested)
	}
}

func TestDecodeRejectsUnknownField(t *testing.T) {
	t.Parallel()

	_, err := config.Decode[pluginConfig]([]byte(`
endpoint = "https://example.test/v1"
unknown = true
`))
	if err == nil {
		t.Fatal("Decode accepted an unknown field")
	}
}

func TestDecodeRejectsInvalidTOML(t *testing.T) {
	t.Parallel()

	_, err := config.Decode[pluginConfig]([]byte(`endpoint = [`))
	if err == nil {
		t.Fatal("Decode accepted invalid TOML")
	}
}

func TestStateDirScope(t *testing.T) {
	t.Parallel()

	parent := context.Background()
	ctx := config.WithStateDir(parent, "/var/lib/ingot/plugin-key")

	got, err := config.StateDir(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/var/lib/ingot/plugin-key" {
		t.Fatalf("StateDir() = %q", got)
	}
	if _, err := config.StateDir(parent); !errors.Is(err, config.ErrStateDirUnavailable) {
		t.Fatalf("parent StateDir error = %v", err)
	}
}

func TestStateDirUnavailable(t *testing.T) {
	t.Parallel()

	for _, ctx := range []context.Context{
		nil,
		context.Background(),
		config.WithStateDir(context.Background(), ""),
	} {
		if _, err := config.StateDir(ctx); !errors.Is(err, config.ErrStateDirUnavailable) {
			t.Fatalf("StateDir(%v) error = %v", ctx, err)
		}
	}
}

func TestResolveTables(t *testing.T) {
	t.Parallel()

	data := []byte(`
[plugins.short]
value = "selected"

[plugins.extra]
value = "ignored"
`)
	tables, err := config.ResolveTables(data, []config.PluginReference{{ID: "github.com/example/plugin", Name: "short"}})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := config.Decode[struct {
		Value string `toml:"value"`
	}](tables["github.com/example/plugin"])
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Value != "selected" {
		t.Fatalf("unexpected value %q", decoded.Value)
	}
}

func TestResolveTablesRequiresExactlyOneReferenceKey(t *testing.T) {
	t.Parallel()

	reference := []config.PluginReference{{ID: "github.com/example/plugin", Name: "short"}}
	_, err := config.ResolveTables([]byte("[plugins.extra]\n"), reference)
	if !errors.Is(err, config.ErrMissingPluginConfig) {
		t.Fatalf("expected missing config error, got %v", err)
	}
	_, err = config.ResolveTables([]byte("[plugins.short]\n[plugins.\"github.com/example/plugin\"]\n"), reference)
	if !errors.Is(err, config.ErrDuplicatePluginConfig) {
		t.Fatalf("expected duplicate config error, got %v", err)
	}
}
