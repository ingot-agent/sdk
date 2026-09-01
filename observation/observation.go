// Package observation defines passive observation of the Agent execution
// hierarchy. Observation reports execution facts and is independent of the
// interceptor control plane.
package observation

import "context"

// Consumer materializes execution details using correlation carried by ctx.
// Emit must not report observer failures to the producer and must limit work on
// the execution path to local ingestion. Implementations are safe for
// concurrent use.
type Consumer interface {
	Emit(context.Context, Detail)
}

// Observer receives self-contained execution events. Observe does not receive
// the execution context because terminal events remain observable after that
// context is canceled. Implementations may maintain ordered local state: an
// observation hub must not invoke the same Observer instance concurrently.
// Observer panics must be isolated by the hub and cannot affect execution.
type Observer interface {
	Observe(Event)
}
