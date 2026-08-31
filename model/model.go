// Package model defines model providers and complete/stream runtime
// chokepoints.
package model

import (
	"context"
	"errors"

	"github.com/ingot-agent/sdk/content"
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

// Message is an ordered multimodal or tool-calling conversation message.
type Message struct {
	Role       Role
	Content    content.Content
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

// RequestResolver materializes the provider and model defaults for an
// invocation without calling a provider or executing model interceptors.
// Implementations return a caller-owned deep copy and are safe for concurrent
// requests.
type RequestResolver interface {
	ResolveRequest(context.Context, Request) (Request, error)
}

// Interceptor wraps a complete model request.
type Interceptor = pipeline.Interceptor[Request, Response]

// StreamEventKind identifies a content part lifecycle event.
type StreamEventKind uint8

const (
	StreamPartStart StreamEventKind = iota + 1
	StreamPartDelta
	StreamPartEnd
)

// StreamEvent is one ordered event in a streamed response. PartIndex is
// zero-based and identifies the corresponding part in the final response.
// DataDelta is immutable only for the duration of a handler call; handlers
// that retain it must copy it.
type StreamEvent struct {
	Kind      StreamEventKind
	PartIndex int

	PartKind content.Kind
	MIMEType string
	Name     string

	TextDelta string
	DataDelta []byte
}

// StreamHandler receives events synchronously in delivery order. Returning an
// error stops streaming immediately and that error is propagated unchanged.
type StreamHandler func(StreamEvent) error

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
	// ErrCapabilitiesUnavailable indicates that the selected provider does not
	// expose model capability information.
	ErrCapabilitiesUnavailable = errors.New("model capabilities unavailable")
)

// CapabilityRequest selects a provider and model whose content capabilities
// should be resolved using the runtime's normal defaulting rules.
type CapabilityRequest struct {
	Provider string
	Model    string
}

// ContentCapability describes support for one content kind, its source forms,
// and the message roles for which it is accepted.
type ContentCapability struct {
	Kind    content.Kind
	Sources []content.SourceKind
	Roles   []Role
}

// Capabilities describes model input, output, and streaming output support.
// Returned slices and all nested slices are owned by the caller.
type Capabilities struct {
	Input           []ContentCapability
	Output          []ContentCapability
	StreamingOutput []content.Kind
}

// CapabilityResolver resolves model capabilities without invoking a model or
// mutating provider configuration. Implementations are safe for concurrent
// use and return caller-owned aggregates.
type CapabilityResolver interface {
	ResolveCapabilities(context.Context, CapabilityRequest) (Capabilities, error)
}

// CapabilityProvider is an optional read-only extension implemented by model
// providers that expose capability information. It does not change Provider.
type CapabilityProvider interface {
	Provider
	Capabilities(context.Context, string) (Capabilities, error)
}
