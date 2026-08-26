package sdk_test

import (
	"context"
	"io/fs"
	"net/http"

	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/contextwindow"
	"github.com/ingot-agent/sdk/filesystem"
	"github.com/ingot-agent/sdk/httpx"
	"github.com/ingot-agent/sdk/interaction"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/pipeline"
	"github.com/ingot-agent/sdk/prompt"
	"github.com/ingot-agent/sdk/session"
	"github.com/ingot-agent/sdk/tool"
)

// These external-package compile assertions protect the public contracts that
// Builder-generated code and third-party components consume.

type httpClient struct{}

func (httpClient) Do(context.Context, *http.Request) (*http.Response, error) {
	return nil, nil
}

var _ httpx.Client = httpClient{}

type workspaceFS struct{}

func (workspaceFS) ReadFile(context.Context, string) ([]byte, error) { return nil, nil }
func (workspaceFS) WriteFile(context.Context, string, []byte, fs.FileMode) error {
	return nil
}
func (workspaceFS) ReadDir(context.Context, string) ([]fs.DirEntry, error) { return nil, nil }
func (workspaceFS) Stat(context.Context, string) (fs.FileInfo, error)      { return nil, nil }
func (workspaceFS) MkdirAll(context.Context, string, fs.FileMode) error    { return nil }
func (workspaceFS) Remove(context.Context, string) error                   { return nil }
func (workspaceFS) Rename(context.Context, string, string) error           { return nil }

var _ filesystem.FS = workspaceFS{}

type toolImplementation struct{}

func (toolImplementation) Definition() tool.Definition { return tool.Definition{} }
func (toolImplementation) Invoke(context.Context, tool.Call) (tool.Result, error) {
	return tool.Result{}, nil
}

var _ tool.Tool = toolImplementation{}

type toolRuntime struct{}

func (toolRuntime) Definitions() []tool.Definition { return nil }
func (toolRuntime) Call(context.Context, tool.Call) (tool.Result, error) {
	return tool.Result{}, nil
}

var _ tool.Runtime = toolRuntime{}

type modelProvider struct{}

func (modelProvider) Complete(context.Context, model.Request) (model.Response, error) {
	return model.Response{}, nil
}

var _ model.Provider = modelProvider{}

type streamingProvider struct{ modelProvider }

func (streamingProvider) Stream(
	context.Context,
	model.Request,
	model.StreamHandler,
) (model.Response, error) {
	return model.Response{}, nil
}

var _ model.StreamingProvider = streamingProvider{}

type modelRuntime struct{ streamingProvider }

var (
	_ model.Runtime          = modelRuntime{}
	_ model.StreamingRuntime = modelRuntime{}
)

type store struct{}

func (store) Create(context.Context, session.Metadata) (session.ID, error) { return "", nil }
func (store) Append(context.Context, session.ID, session.Entry) error      { return nil }
func (store) Load(context.Context, session.ID) ([]session.Entry, error)    { return nil, nil }
func (store) List(context.Context, session.Query) ([]session.Summary, error) {
	return nil, nil
}

var _ session.Store = store{}

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
	Messages: []model.Message{{Role: model.RoleUser, Content: "hello"}},
	Changed:  true,
}

type channel struct{}

func (channel) Ask(context.Context, interaction.AskRequest) (interaction.AskResponse, error) {
	return interaction.AskResponse{}, nil
}
func (channel) Render(context.Context, interaction.Event) error { return nil }

var _ interaction.Channel = channel{}

var _ = interaction.AskRequest{
	Prompt: "Continue?",
	Options: []interaction.AskOption{
		{Label: "Yes", Description: "Continue the operation."},
		{Label: "No", Description: "Stop the operation."},
	},
	AllowTextInput: true,
}

var (
	_ interaction.Event = interaction.TextEvent{}
	_ interaction.Event = interaction.StatusEvent{}
	_ interaction.Event = interaction.ErrorEvent{}
	_ interaction.Event = interaction.ToolCallEvent{}
	_ interaction.Event = interaction.ToolResultEvent{}
)

type agentRuntime struct{}

func (agentRuntime) Run(context.Context, agent.Turn) (agent.Result, error) {
	return agent.Result{}, nil
}

var _ agent.Runtime = agentRuntime{}

type toolInterceptor struct{}

func (toolInterceptor) Invoke(
	ctx context.Context,
	call tool.Call,
	next pipeline.Next[tool.Call, tool.Result],
) (tool.Result, error) {
	return next(ctx, call)
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
