package agent

import "errors"

var (
	// ErrStreamingUnsupported indicates unavailable model streaming support.
	ErrStreamingUnsupported = errors.New("agent streaming unsupported")
	// ErrNilStreamHandler indicates that Stream was called without a handler.
	ErrNilStreamHandler = errors.New("nil agent stream handler")
)
