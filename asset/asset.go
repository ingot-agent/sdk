// Package asset defines immutable binary asset storage and resolution.
package asset

import (
	"context"
	"io"
)

// Reference is an opaque, globally unique, persistent asset identity. Its ID
// has no caller-visible structure; the zero value does not identify an asset.
type Reference struct {
	ID string
}

// Info describes stored bytes without assigning media semantics.
type Info struct {
	Size uint64
}

// PutRequest streams exactly Size bytes into a Store. Put consumes Body before
// returning and does not close it. Size zero is valid.
type PutRequest struct {
	Body io.Reader
	Size uint64
}

// Resolver opens immutable assets. Implementations are safe for concurrent
// calls. Each successful Open returns an independent reader that the caller
// must close.
type Resolver interface {
	Stat(context.Context, Reference) (Info, error)
	Open(context.Context, Reference) (io.ReadCloser, error)
}

// Store imports immutable bytes and resolves resulting references. A
// successful Put consumes and verifies exactly PutRequest.Size bytes before
// returning. Implementations never reuse a Reference for different bytes.
type Store interface {
	Resolver
	Put(context.Context, PutRequest) (Reference, Info, error)
}
