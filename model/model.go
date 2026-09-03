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

// Usage reports token counts for a model response. Reported distinguishes an
// explicitly reported zero from unavailable provider execution usage.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Reported     bool
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
// requests. A Response is authoritative only when the returned error is nil;
// an error does not imply that generation, billing, or other provider effects
// did not occur or that retrying is safe.
type Provider interface {
	Complete(context.Context, Request) (Response, error)
}

// Runtime selects a named provider and executes complete requests through its
// interceptor chain. A Response is authoritative only when the returned error
// is nil. A non-nil error does not imply that provider-side effects did not
// occur or that retrying the request is safe.
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

// StreamSemantic describes the purpose of a streamed part independently of
// its content modality.
type StreamSemantic uint8

const (
	// StreamSemanticContent is ordinary response content. Its zero value keeps
	// existing providers compatible.
	StreamSemanticContent StreamSemantic = iota
	// StreamSemanticReasoning is provider-explicit reasoning/thinking text.
	// It is transient and does not enter Response.Message.Content.
	StreamSemanticReasoning
)

// StreamEvent is one ordered event in a streamed response. PartIndex is
// zero-based within each Semantic. Content indices identify the corresponding
// part in the final response; reasoning indices identify transient text parts.
// Each semantic has one active part at a time; the two streams may interleave.
// Semantic is set on every start/delta/end event. PartKind, MIMEType and Name
// are set only on start; delta/end retain the part's index and semantic.
// DataDelta is immutable only for the duration of a handler call; handlers
// that retain it must copy it.
type StreamEvent struct {
	Kind      StreamEventKind
	PartIndex int
	Semantic  StreamSemantic

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
// requests. A Response is authoritative only when Stream returns nil. Events
// delivered before an error are transient progress, not a canonical response.
type StreamingProvider interface {
	Provider
	Stream(context.Context, Request, StreamHandler) (Response, error)
}

// StreamingRuntime selects a named streaming provider and executes requests
// through the streaming interceptor chain. A Response is authoritative only
// when the returned error is nil. Delivered events are transient progress and
// are not a canonical response when Stream returns an error.
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
	// ErrStreamingUnsupported is an explicit pre-stream rejection indicating
	// that streaming mode is unavailable. It is not a general invocation
	// failure. Callers may use it for a documented mode fallback only when no
	// externally observable stream progress has been delivered.
	ErrStreamingUnsupported = errors.New("streaming unsupported")
	// ErrProviderNotFound indicates that no provider has the requested name.
	ErrProviderNotFound = errors.New("provider not found")
	// ErrModelNotFound indicates that the selected provider does not expose the
	// requested model.
	ErrModelNotFound = errors.New("model not found")
)
