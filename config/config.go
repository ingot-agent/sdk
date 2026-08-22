// Package config provides strict plugin configuration decoding and access to
// the plugin-scoped persistent state directory.
package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

// ErrStateDirUnavailable indicates that a context has no plugin state scope.
var ErrStateDirUnavailable = errors.New("plugin state directory unavailable")

var (
	// ErrMissingPluginConfig indicates that a locked plugin has no table under
	// [plugins] using either its canonical ID or short name.
	ErrMissingPluginConfig = errors.New("missing plugin config")
	// ErrDuplicatePluginConfig indicates that both the canonical ID and short
	// name tables are present for one plugin.
	ErrDuplicatePluginConfig = errors.New("duplicate plugin config")
)

// PluginReference is the locked identity used to locate one runtime config
// table. ID and Name must each be unique within a reference set.
type PluginReference struct {
	ID   string
	Name string
}

// ResolveTables locates exactly one table for every locked plugin. Extra
// tables belonging to plugins outside the current image are ignored. Returned
// bytes can be passed to Decode.
func ResolveTables(data []byte, references []PluginReference) (map[string][]byte, error) {
	var document struct {
		Plugins map[string]map[string]any `toml:"plugins"`
	}
	if err := toml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse runtime config: %w", err)
	}
	result := make(map[string][]byte, len(references))
	ids := make(map[string]bool, len(references))
	names := make(map[string]bool, len(references))
	for index, reference := range references {
		if reference.ID == "" || reference.Name == "" {
			return nil, fmt.Errorf("plugin reference at index %d has empty id or name", index)
		}
		if ids[reference.ID] || names[reference.Name] {
			return nil, fmt.Errorf("plugin reference at index %d is not unique", index)
		}
		ids[reference.ID], names[reference.Name] = true, true
		byID, hasID := document.Plugins[reference.ID]
		byName, hasName := document.Plugins[reference.Name]
		if hasID && hasName {
			return nil, fmt.Errorf("plugin %s (%s): %w", reference.ID, reference.Name, ErrDuplicatePluginConfig)
		}
		if !hasID && !hasName {
			return nil, fmt.Errorf("plugin %s (%s): %w", reference.ID, reference.Name, ErrMissingPluginConfig)
		}
		table := byName
		if hasID {
			table = byID
		}
		encoded, err := toml.Marshal(table)
		if err != nil {
			return nil, fmt.Errorf("encode config table for plugin %s: %w", reference.ID, err)
		}
		result[reference.ID] = encoded
	}
	return result, nil
}

// Decode strictly decodes one resolved plugin TOML table into T. Unknown
// fields are rejected so configuration typos fail during startup.
func Decode[T any](data []byte) (T, error) {
	var value T
	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode plugin config: %w", err)
	}
	return value, nil
}

type stateDirKey struct{}

// WithStateDir returns a child context carrying the state directory assigned
// to a plugin. It is intended for generated wiring; components should retrieve
// the directory with StateDir.
//
// An empty dir deliberately creates no usable scope and causes StateDir to
// return ErrStateDirUnavailable.
func WithStateDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, stateDirKey{}, dir)
}

// StateDir returns the state directory assigned to the calling plugin.
func StateDir(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrStateDirUnavailable
	}
	dir, ok := ctx.Value(stateDirKey{}).(string)
	if !ok || dir == "" {
		return "", ErrStateDirUnavailable
	}
	return dir, nil
}
