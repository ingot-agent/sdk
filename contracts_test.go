package sdk_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/asset"
	"github.com/ingot-agent/sdk/content"
	"github.com/ingot-agent/sdk/contextwindow"
	"github.com/ingot-agent/sdk/execution"
	"github.com/ingot-agent/sdk/httpx"
	"github.com/ingot-agent/sdk/interaction"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/observation"
	"github.com/ingot-agent/sdk/operation"
	"github.com/ingot-agent/sdk/pipeline"
	"github.com/ingot-agent/sdk/prompt"
	"github.com/ingot-agent/sdk/session"
	"github.com/ingot-agent/sdk/tool"
	"github.com/ingot-agent/sdk/workspace"
)

// These external-package compile assertions protect the public contracts that
// Builder-generated code and third-party components consume.

type httpClient struct{}

func (httpClient) Do(context.Context, *http.Request) (*http.Response, error) {
	return nil, nil
}

var _ httpx.Client = httpClient{}

type assetStore struct{}

func (assetStore) Stat(context.Context, asset.Reference) (asset.Info, error) {
	return asset.Info{}, nil
}
func (assetStore) Open(context.Context, asset.Reference) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(nil)), nil
}
func (assetStore) Put(context.Context, asset.PutRequest) (asset.Reference, asset.Info, error) {
	return asset.Reference{}, asset.Info{}, nil
}

var (
	_ asset.Resolver = assetStore{}
	_ asset.Store    = assetStore{}
)

type toolImplementation struct{}

func (toolImplementation) Definition() tool.Definition { return tool.Definition{} }
func (toolImplementation) Invoke(context.Context, tool.Invocation) (tool.Result, error) {
	return tool.Result{}, nil
}

var _ tool.Tool = toolImplementation{}

type toolRuntime struct{}

func (toolRuntime) Definitions() []tool.Definition { return nil }
func (toolRuntime) Call(context.Context, tool.Invocation) (tool.Result, error) {
	return tool.Result{}, nil
}

var _ tool.Runtime = toolRuntime{}

type modelProvider struct{}

func (modelProvider) Complete(context.Context, model.Request) (model.Response, error) {
	return model.Response{}, nil
}

type modelProviderSource struct{}

func (modelProviderSource) Snapshot(context.Context) ([]model.ProviderEntry, error) {
	provider := streamingProvider{}
	return []model.ProviderEntry{{Name: "provider", Complete: provider.Complete, Stream: provider.Stream}}, nil
}

var _ model.ProviderSource = modelProviderSource{}

type streamingProvider struct{ modelProvider }

func (streamingProvider) Stream(
	context.Context,
	model.Request,
	model.StreamHandler,
) (model.Response, error) {
	return model.Response{}, nil
}

type modelRuntime struct{ streamingProvider }

var (
	_ model.Runtime          = modelRuntime{}
	_ model.StreamingRuntime = modelRuntime{}
)

type store struct{}

func (store) Create(context.Context, session.CreateRequest) (session.Metadata, error) {
	return session.Metadata{}, nil
}
func (store) Append(context.Context, session.ID, session.Entry) error   { return nil }
func (store) Load(context.Context, session.ID) ([]session.Entry, error) { return nil, nil }
func (store) Get(context.Context, session.ID) (session.Metadata, error) {
	return session.Metadata{}, nil
}
func (store) Rename(context.Context, session.ID, string) (session.Metadata, error) {
	return session.Metadata{}, nil
}
func (store) Archive(context.Context, session.ID) (session.Metadata, error) {
	return session.Metadata{}, nil
}
func (store) Restore(context.Context, session.ID) (session.Metadata, error) {
	return session.Metadata{}, nil
}
func (store) Delete(context.Context, session.ID) error { return nil }
func (store) Fork(context.Context, session.ID, session.ForkRequest) (session.Metadata, error) {
	return session.Metadata{}, nil
}
func (store) List(context.Context) ([]session.Metadata, error) {
	return nil, nil
}

var (
	_ session.Store   = store{}
	_ session.Manager = store{}
	_ session.Query   = store{}
)

type contributor struct{}

func (contributor) Contribute(context.Context, prompt.Request) ([]prompt.Block, error) {
	return nil, nil
}

var _ prompt.Contributor = contributor{}

type renderer struct{}

func (renderer) Render(context.Context, prompt.Request) ([]model.Message, error) {
	return nil, nil
}

var _ prompt.Renderer = renderer{}

type compactor struct{}

func (compactor) Compact(
	context.Context,
	contextwindow.CompactionRequest,
) (contextwindow.CompactionResult, error) {
	return contextwindow.CompactionResult{}, nil
}

var _ contextwindow.Compactor = compactor{}

var _ = contextwindow.CompactionRequest{
	SessionID: "session-1",
	Invocation: model.Request{
		Provider: "provider",
		Model:    "model",
	},
}

var _ = contextwindow.CompactionResult{
	Messages: []model.Message{{Role: model.RoleUser, Content: content.FromText("hello")}},
	Changed:  true,
}

type channel struct{}

func (channel) Request(context.Context, interaction.Request) (interaction.Response, error) {
	return interaction.Response{}, nil
}
func (channel) Emit(context.Context, interaction.Event) error { return nil }
func (channel) Set(context.Context, interaction.State) error  { return nil }
func (channel) Clear(context.Context, string) error           { return nil }

var _ interaction.Channel = channel{}

type executionBinder struct{}

func (executionBinder) Bind(scope execution.Scope) (interaction.Channel, error) {
	_ = scope.SessionID
	return channel{}, nil
}

var _ interaction.ExecutionBinder = executionBinder{}

var _ = interaction.Request{
	Name:        "continue",
	Description: "Continue?",
	Fields: []interaction.Field{
		{
			Name:     "decision",
			Label:    "Decision",
			Kind:     interaction.FieldChoice,
			Required: true,
			Options: []interaction.Option{
				{Value: "yes", Label: "Yes", Description: "Continue the operation."},
				{Value: "no", Label: "No", Description: "Stop the operation."},
			},
		},
	},
}

var (
	_ = interaction.Event{Name: "connection_lost", Level: interaction.LevelWarning}
	_ = interaction.State{Name: "connection", Values: []interaction.Entry{{Name: "connected", Value: interaction.BooleanValue(false)}}}
	_ = interaction.Response{Values: []interaction.Answer{{Name: "decision", Value: interaction.StringValue("yes")}}}
)

type operationImplementation struct{}

func (operationImplementation) Definition() operation.Definition {
	return operation.Definition{
		Name:         "session.inspect",
		Description:  "Inspect one session.",
		InputSchema:  json.RawMessage(`{"type":"object"}`),
		OutputSchema: json.RawMessage(`{"type":"object"}`),
	}
}

func (operationImplementation) Invoke(context.Context, operation.Request) (operation.Result, error) {
	return operation.Result{Output: json.RawMessage(`{}`)}, nil
}

var _ operation.Operation = operationImplementation{}

var _ = operation.Request{
	SessionID:   "session-1",
	Input:       json.RawMessage(`{}`),
	Interaction: interaction.Unavailable(),
}

type agentRuntime struct{}

func (agentRuntime) Run(context.Context, agent.Turn) (agent.Execution, error) {
	return agent.Execution{}, nil
}

var _ agent.Runtime = agentRuntime{}

type observationConsumer struct{}

func (observationConsumer) Emit(context.Context, observation.Detail) {}

type executionObserver struct{}

func (executionObserver) Observe(observation.Event) {}

var (
	_ observation.Consumer = observationConsumer{}
	_ observation.Observer = executionObserver{}
)

type toolInterceptor struct{}

func (toolInterceptor) Invoke(
	ctx context.Context,
	invocation tool.Invocation,
	next pipeline.Next[tool.Invocation, tool.Result],
) (tool.Result, error) {
	return next(ctx, invocation)
}

var _ tool.Interceptor = toolInterceptor{}

type modelInterceptor struct{}

func (modelInterceptor) Invoke(
	ctx context.Context,
	request model.Request,
	next pipeline.Next[model.Request, model.Response],
) (model.Response, error) {
	return next(ctx, request)
}

var _ model.Interceptor = modelInterceptor{}

type streamInterceptor struct{}

func (streamInterceptor) InvokeStream(
	ctx context.Context,
	request model.Request,
	handler model.StreamHandler,
	next model.StreamNext,
) (model.Response, error) {
	return next(ctx, request, handler)
}

var _ model.StreamInterceptor = streamInterceptor{}

type agentInterceptor struct{}

func (agentInterceptor) Invoke(
	ctx context.Context,
	turn agent.Turn,
	next pipeline.Next[agent.Turn, agent.Result],
) (agent.Result, error) {
	return next(ctx, turn)
}

var _ agent.Interceptor = agentInterceptor{}

type workspaceResolver struct{}

func (workspaceResolver) Resolve(context.Context, execution.Scope) (workspace.Binding, error) {
	return workspace.Binding{}, nil
}

var _ workspace.Resolver = workspaceResolver{}

type workspaceManager struct{}

func (workspaceManager) Assign(context.Context, session.ID, workspace.Binding) error {
	return nil
}

var _ workspace.Manager = workspaceManager{}

// toolEnvelope tests that a Tool, Runtime, and Interceptor implementation can
// read the execution scope and the durable call payload through the public
// invocation envelope without any hidden context requirement.
type toolEnvelope struct{}

func (toolEnvelope) Definition() tool.Definition { return tool.Definition{} }
func (toolEnvelope) Invoke(_ context.Context, invocation tool.Invocation) (tool.Result, error) {
	_ = invocation.Scope.SessionID
	_ = invocation.Call.Name
	return tool.Result{}, nil
}

var _ tool.Tool = toolEnvelope{}
