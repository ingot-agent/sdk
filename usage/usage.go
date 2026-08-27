// Package usage defines model input counting contracts shared by context
// budgeting, estimates, and usage monitoring.
package usage

import (
	"context"
	"errors"

	"github.com/ingot-agent/sdk/model"
)

// Accuracy describes what a caller may safely assume about a count.
type Accuracy string

const (
	// AccuracyExact means the tokenizer and request framing are implemented
	// exactly for the resolved provider and model.
	AccuracyExact Accuracy = "exact"
	// AccuracyUpperBound means the result is guaranteed not to be lower than
	// the actual model input count.
	AccuracyUpperBound Accuracy = "upper_bound"
	// AccuracyEstimate means the result is an approximation without an upper-
	// or lower-bound guarantee.
	AccuracyEstimate Accuracy = "estimate"
)

// CountRequest contains one immutable, complete model invocation.
type CountRequest struct {
	Invocation model.Request
}

// CountResult reports the model input count and how it was obtained.
type CountResult struct {
	InputTokens int64
	Accuracy    Accuracy
	Source      string
	Provider    string
	Model       string
}

// Counter computes model-aware input counts before a provider invocation.
// Implementations are safe for concurrent calls.
type Counter interface {
	CountInput(context.Context, CountRequest) (CountResult, error)
}

var (
	// ErrUnsupportedModel indicates that no configured counting strategy
	// supports the resolved provider and model.
	ErrUnsupportedModel = errors.New("usage counter does not support model")
)
