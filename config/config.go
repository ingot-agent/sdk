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
