// Package sdk contains the composition primitives shared by ingot components.
package sdk

import (
	"context"
	"errors"
	"fmt"
)

// Cleanup releases the resources owned by one component instance.
//
// A Cleanup must observe ctx.Done and return promptly when the context is
// cancelled. Generated wiring calls component cleanups synchronously in the
// reverse of component creation order.
type Cleanup func(context.Context) error

// Optional represents an OPTIONAL component dependency. Valid distinguishes a
// provided zero value from an absent value.
type Optional[T any] struct {
	Value T
	Valid bool
}

// None returns an Optional with no value.
func None[T any]() Optional[T] {
	return Optional[T]{}
}

// Some returns an Optional containing value.
func Some[T any](value T) Optional[T] {
	return Optional[T]{Value: value, Valid: true}
}

// Named gives a runtime capability instance a stable name. The name is scoped
// to the collection in which the value is exported.
type Named[T any] struct {
	Name  string
	Value T
}

var (
	// ErrEmptyName indicates that a Named value has no runtime instance name.
	ErrEmptyName = errors.New("empty capability name")
	// ErrDuplicateName indicates that two Named values in one collection use
	// the same runtime instance name.
	ErrDuplicateName = errors.New("duplicate capability name")
)

// CheckUniqueNames verifies that every item has a non-empty, unique name.
func CheckUniqueNames[T any](items []Named[T]) error {
	seen := make(map[string]int, len(items))
	for i, item := range items {
		if item.Name == "" {
			return fmt.Errorf("named capability at index %d: %w", i, ErrEmptyName)
		}
		if previous, ok := seen[item.Name]; ok {
			return fmt.Errorf(
				"named capability %q at index %d duplicates index %d: %w",
				item.Name,
				i,
				previous,
				ErrDuplicateName,
			)
		}
		seen[item.Name] = i
	}
	return nil
}
