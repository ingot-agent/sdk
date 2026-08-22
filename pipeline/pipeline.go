// Package pipeline provides generic, typed interceptor composition.
package pipeline

import "context"

// Next is the next operation in an interceptor chain.
type Next[Req, Res any] func(context.Context, Req) (Res, error)

// Interceptor wraps a typed operation.
type Interceptor[Req, Res any] interface {
	Invoke(context.Context, Req, Next[Req, Res]) (Res, error)
}

// Compose wraps terminal with interceptors. The first interceptor is the
// outermost: its before logic runs first and its after logic runs last.
//
// Compose does not mutate the interceptor slice. A nil terminal or nil
// interceptor is a programming error and will panic when invoked.
func Compose[Req, Res any](
	terminal Next[Req, Res],
	interceptors ...Interceptor[Req, Res],
) Next[Req, Res] {
	next := terminal
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		following := next
		next = func(ctx context.Context, req Req) (Res, error) {
			return interceptor.Invoke(ctx, req, following)
		}
	}
	return next
}
