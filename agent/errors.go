package agent

import "errors"

var (
	// ErrNilStreamHandler indicates that Stream was called without a handler.
	ErrNilStreamHandler = errors.New("nil agent stream handler")
	// ErrMaxRounds indicates that another round cannot be completed within the
	// configured turn limit.
	ErrMaxRounds = errors.New("maximum agent rounds exceeded")
	// ErrInvalidRound indicates that an interceptor changed immutable round
	// identity or execution facts.
	ErrInvalidRound = errors.New("invalid agent round")
	// ErrInvalidRoundDecision indicates that an interceptor produced a
	// canonical decision outside the permitted mutation contract.
	ErrInvalidRoundDecision = errors.New("invalid agent round decision")
	// ErrInvalidRoundResult indicates an invalid short-circuit result or a
	// rewrite of a result after durable round execution.
	ErrInvalidRoundResult = errors.New("invalid agent round result")
)
