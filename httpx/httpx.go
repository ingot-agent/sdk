// Package httpx defines the shared HTTP client capability.
package httpx

import (
	"context"
	"net/http"
)

// Client executes HTTP requests. The explicit ctx is the cancellation and
// deadline authority. Implementations must leave req unchanged and behave as
// if they cloned it with req.Clone(ctx) before dispatch.
//
// Implementations are safe for concurrent use.
type Client interface {
	Do(context.Context, *http.Request) (*http.Response, error)
}
