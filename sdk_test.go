package sdk_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ingot-agent/sdk"
)

func TestOptionalDistinguishesAbsentAndZeroValue(t *testing.T) {
	t.Parallel()

	none := sdk.None[int]()
	if none.Valid || none.Value != 0 {
		t.Fatalf("None[int]() = %#v", none)
	}

	some := sdk.Some(0)
	if !some.Valid || some.Value != 0 {
		t.Fatalf("Some(0) = %#v", some)
	}
}

func TestCheckUniqueNames(t *testing.T) {
	t.Parallel()

	if err := sdk.CheckUniqueNames([]sdk.Named[int]{
		{Name: "primary", Value: 1},
		{Name: "fallback", Value: 2},
	}); err != nil {
		t.Fatalf("unique names: %v", err)
	}
}

func TestCheckUniqueNamesRejectsEmptyName(t *testing.T) {
	t.Parallel()

	err := sdk.CheckUniqueNames([]sdk.Named[int]{{Value: 1}})
	if !errors.Is(err, sdk.ErrEmptyName) {
		t.Fatalf("error = %v, want ErrEmptyName", err)
	}
}

func TestCheckUniqueNamesRejectsDuplicate(t *testing.T) {
	t.Parallel()

	err := sdk.CheckUniqueNames([]sdk.Named[int]{
		{Name: "same", Value: 1},
		{Name: "same", Value: 2},
	})
	if !errors.Is(err, sdk.ErrDuplicateName) {
		t.Fatalf("error = %v, want ErrDuplicateName", err)
	}
}

func TestCleanupHasSpecifiedSignature(t *testing.T) {
	t.Parallel()

	called := false
	cleanup := sdk.Cleanup(func(context.Context) error {
		called = true
		return nil
	})
	if err := cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("cleanup was not called")
	}
}
