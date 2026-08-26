// Package application defines process-scoped application runtime controls.
package application

import (
	"context"
	"errors"
	"reflect"
)

// ErrUnavailable indicates that a Context does not carry a process control.
var ErrUnavailable = errors.New("application process control unavailable")

// Process exposes immutable invocation metadata and a controlled process
// shutdown request. Implementations are concurrent-safe. Arguments returns a
// caller-owned copy. Shutdown is idempotent; the first request determines the
// process result, with a nil error indicating normal completion.
type Process interface {
	Arguments() []string
	Check() bool
	Shutdown(error)
}

type processContextKey struct{}

// WithProcess returns a child Context carrying process-scoped application
// control. It panics when ctx is nil or process is nil, including typed nil.
func WithProcess(ctx context.Context, process Process) context.Context {
	if ctx == nil {
		panic("application.WithProcess: nil context")
	}
	if isNil(process) {
		panic("application.WithProcess: nil process")
	}
	return context.WithValue(ctx, processContextKey{}, process)
}

// FromContext returns the process control assigned to ctx.
func FromContext(ctx context.Context) (Process, error) {
	if ctx == nil {
		return nil, ErrUnavailable
	}
	process, ok := ctx.Value(processContextKey{}).(Process)
	if !ok || isNil(process) {
		return nil, ErrUnavailable
	}
	return process, nil
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
