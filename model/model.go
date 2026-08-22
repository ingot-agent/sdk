// Package model defines model providers and complete/stream runtime
// chokepoints.
package model

import (
	"context"
	"errors"

	"github.com/ingot-agent/sdk/pipeline"
	"github.com/ingot-agent/sdk/tool"
)

// Role identifies a message participant.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is a text or tool-calling conversation message.
type Message struct {
	Role       Role
	Content    string
	Name       string
	ToolCallID string
	ToolCalls  []tool.Call
}

// Request is one model invocation. Provider chooses a named provider instance;
// Model chooses a model exposed by that provider.
type Request struct {
	Provider    string
	Model       string
	Messages    []Message
	Tools       []tool.Definition
	Temperature *float64
	MaxTokens   *int
	Stop        []string
}

// Usage reports token counts for a model response.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// Response is the final result of a complete or streaming model invocation.
type Response struct {
	Message      Message
	FinishReason string
	Usage        Usage
	Provider     string
	Model        string
}

// Provider completes model requests. Implementations are safe for concurrent
// requests.
type Provider interface {
	Complete(context.Context, Request) (Response, error)
}

// Runtime selects a named provider and executes complete requests through its
// interceptor chain.
type Runtime interface {
	Complete(context.Context, Request) (Response, error)
}

// Interceptor wraps a complete model request.
type Interceptor = pipeline.Interceptor[Request, Response]

// StreamChunk is one streaming text delta.
type StreamChunk struct {
	TextDelta string
}

// StreamHandler receives chunks in delivery order. Returning an error stops
// streaming immediately and that error is propagated to the caller.
type StreamHandler func(StreamChunk) error

// StreamingProvider is a provider capable of both complete and streaming
// requests.
type StreamingProvider interface {
	Provider
	Stream(context.Context, Request, StreamHandler) (Response, error)
}

// StreamingRuntime selects a named streaming provider and executes requests
// through the streaming interceptor chain.
type StreamingRuntime interface {
	Stream(context.Context, Request, StreamHandler) (Response, error)
}

// StreamNext is the next operation in a streaming interceptor chain.
type StreamNext func(
	context.Context,
	Request,
	StreamHandler,
) (Response, error)

// StreamInterceptor wraps a streaming model request.
type StreamInterceptor interface {
	InvokeStream(
		context.Context,
		Request,
		StreamHandler,
		StreamNext,
	) (Response, error)
}

var (
	// ErrStreamingUnsupported indicates that the selected provider does not
	// implement StreamingProvider.
	ErrStreamingUnsupported = errors.New("streaming unsupported")
	// ErrProviderNotFound indicates that no provider has the requested name.
	ErrProviderNotFound = errors.New("provider not found")
	// ErrModelNotFound indicates that the selected provider does not expose the
	// requested model.
	ErrModelNotFound = errors.New("model not found")
)
