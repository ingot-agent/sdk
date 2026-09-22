# SDK migration guide

This guide covers the tagged SDK APIs through **v0.2.11**. Use the
[changelog](../CHANGELOG.md) to identify every change between your current and
target versions. Package comments and their external-package tests are the
authoritative contract reference; this guide explains integration work across
those packages.

## Choose and verify one module graph

The SDK is optional. Plugins import the contract packages they use directly;
the Builder has no SDK selection/configuration field. A separate domain SDK
works the same way. The Core and `ingot-abi` have independent versions.

There is no blanket compatibility promise for the `v0.x` series. Historical
patch releases have changed both interfaces and semantics. A successful Go
compile catches signature changes but not missing metadata persistence, scope
routing, ownership violations, or incorrect cancellation behavior.

1. Read each consuming module's `go.mod`, including indirect requirements. Go's
   minimal version selection picks **one version per module path** in the build
   graph; different plugins do not get isolated SDK copies.
2. Select a tag containing all required contracts. For example, live provider
   sources require v0.2.10; single-Turn child contracts require v0.2.11.
3. Download that exact version, update each consumer, and inspect the resulting
   module graph. Run these in a consumer module, not in the SDK repository:

   ```sh
   GOWORK=off go mod download -json github.com/ingot-agent/sdk@v0.2.11
   GOWORK=off go get github.com/ingot-agent/sdk@v0.2.11
   GOWORK=off go mod tidy
   GOWORK=off go list -m all
   GOWORK=off go test ./...
   GOWORK=off go vet ./...
   ```

   These are POSIX shell commands. In PowerShell, set `$env:GOWORK = 'off'` in
   that shell and run the same `go ...` commands without the prefix.
4. Rebuild the consuming Runtime Image and exercise real capability wiring and
   persistence. A local `go.work` replacement can hide an unavailable release
   or an insufficient published `go.mod` requirement; run release validation
   with `GOWORK=off` and without unpublished replacement paths.
5. Back up persistent plugin data and use that plugin's migration procedure.
   The SDK does not migrate databases or decode plugin-owned entry payloads.

Repository tags were inspected for this guide on 2026-09-22, and the 19 local
tagged commits were checked against GitHub's remote tag refs. This records
source contracts, not Go proxy availability. Verify downloads in the release
environment; do not publish a consumer whose required SDK version cannot be
resolved there.

## Runtime contracts and the ABI (v0.1.6)

The root SDK package and the `application` / `config` packages were removed.
Import the fixed runtime host contracts from `github.com/ingot-agent/ingot-abi`.
Do not add an ABI dependency to a library that only uses SDK capabilities.

The v0.1.6 tag identifies the SDK/ABI split. The two-argument constructor and
plugin-owned settings described below are requirements of the **current Core
Builder**; the SDK tag alone does not imply that an older Core removed its
`Config` constructor parameter at the same time.

| Old dependency or call | Current responsibility |
| --- | --- |
| `sdk.Cleanup`, `sdk.Optional[T]`, `sdk.Named[T]`, related helpers | Root `ingot-abi` package, conventionally aliased `ingotabi` |
| `application.Process.Arguments()` | Injected `invocation.Invocation.Arguments()` |
| `application.Process.Check()` | `invocation.Invocation.Mode() == invocation.ModeCheck` |
| `application.Process.Shutdown(err)` | Injected `lifecycle.Controller.RequestShutdown(err)` |
| `application.WithProcess` / `FromContext` | Explicit `Dependencies` fields; remove context lookup |
| `config.WithStateDir` / `StateDir` | Explicit `state.Scope`; use `Dir()` |
| `config.ResolveTables`, `config.Decode`, constructor `Config` parameter | Plugin-owned loading, validation, defaults, and persistence |

This complete constructor shape illustrates the current ABI boundary:

```go
package component

import (
    "context"

    ingotabi "github.com/ingot-agent/ingot-abi"
    "github.com/ingot-agent/ingot-abi/invocation"
    "github.com/ingot-agent/ingot-abi/lifecycle"
    "github.com/ingot-agent/ingot-abi/state"
)

type Dependencies struct {
    Invocation invocation.Invocation
    Lifecycle  lifecycle.Controller
    State      state.Scope
}

type Exports struct{}

func New(ctx context.Context, deps Dependencies) (Exports, ingotabi.Cleanup, error) {
    if err := ctx.Err(); err != nil {
        return Exports{}, nil, err
    }
    // Load and validate plugin-owned settings using deps.State.Dir().
    if deps.Invocation.Mode() == invocation.ModeCheck {
        return Exports{}, nil, nil
    }
    // Construct capabilities here. An application component may use
    // deps.Invocation.Arguments() and deps.Lifecycle.RequestShutdown(err).
    return Exports{}, nil, nil
}
```

This is an ABI shape, not a complete settings implementation. Declare only the
dependencies your component actually needs. During check mode validate settings
and dependencies without starting interaction loops or occupying external
resources. `RequestShutdown` does not mean `os.Exit`: the runtime owns context
cancellation, reverse-order cleanup, aggregated failures, and process exit.
Later non-nil shutdown causes are aggregated; the old Process first-result
behavior must not be assumed.

The required ABI version is controlled by the selected Core's fixed ABI
contract. Follow the
[ABI repository](https://github.com/ingot-agent/ingot-abi) and the consuming Core
release instructions rather than setting ABI and SDK to numerically matching
versions.

## Structured host interaction (v0.1.1, v0.1.5, v0.2.9)

`ReadLine` disappeared in v0.1.1. The temporary replacement was `Ask` without
options; v0.1.5 then replaced that interface with structured effects:

- Use `Channel.Request` with named typed `Field` values for input and consume
  the returned `Response.Values`.
- Use `Emit` for an event, `Set` for current state, and `Clear` to remove state.
  These do not prescribe rendering, terminal behavior, or widgets.
- Supply `interaction.Unavailable()` for an explicit non-interactive channel;
  handle `ErrUnavailable` and preserve cancellation errors. Do not treat absence
  of interaction as implicit permission for an operation.
- Since v0.2.9, object/list fields and values can be nested. Hosts that support
  them must validate recursively. `ObjectValue` and `ListValue` copy the outer
  slice; do not assume the entire nested value graph is deep-copied.

Execution-time interaction should be bound through `ExecutionBinder` as
described in the scope migration below. An `operation.Request` already
contains its explicit call-scoped channel; operation hosts own that binding.

## Ordered content and opaque persistence (v0.2.0)

Content is now `content.Content`, an ordered slice of tagged parts. For a
text-only integration, replace string assignments explicitly:

```go
message := model.Message{
    Role:    model.RoleUser,
    Content: content.FromText("Hello"),
}
result := tool.Result{Content: content.FromText("Done")}
output := agent.Result{Output: content.FromText("Done")}
```

Use `content.TextOnly(value)` only when refusing non-text content is intentional;
check its boolean result. It never silently strips media. Use `content.Clone`
when retaining or modifying content owned by another caller. Copy nested inline
bytes and attachments as well. `content.Validate` checks tagged unions and
UTF-8; it does not validate provider modality support, media bytes, MIME, URI
permissions, size limits, or local filesystem access.

An agent Turn keeps `Input string` and adds ordered non-text `Attachments`.
`content.FromInput` normalizes them into an optional leading text part followed
by attachments. Model stream events use part start/delta/end, replacing the
earlier text-chunk API; see streaming below.

`session.Entry.Payload` is now opaque bytes. A Session provider must preserve
those exact bytes and their `Kind`, `Version`, and append order. An agent that
owns an entry schema must version and migrate it explicitly. In particular,
standard `encoding/json` encodes a `[]byte` field as base64; merely changing a
persisted string field to `[]byte` does not establish backward compatibility
for historical JSON records. The appropriate conversion belongs to the plugin
that owns the format.

## Streaming, rounds, and execution semantics (v0.2.1–v0.2.4)

Remove dependencies on the model capability metadata interfaces introduced in
v0.2.0 and removed in v0.2.4. Do not use stale capability probes as authorization
to invoke a model. Resolve the typed capability, attempt the actual operation,
and handle its documented errors.

`agent.StreamingRuntime` is independent of `agent.Runtime`; inject the capability
you consume. `agent.ErrStreamingUnsupported` no longer exists. An agent may
use a completion path internally when model streaming is unavailable, and a
successful agent stream may deliver no events.

For `model.StreamingRuntime` and `ProviderEntry.Stream`:

1. Deliver part start/delta/end synchronously in order. The `Semantic` is set on
   every event; ordinary content and provider-explicit reasoning each number
   parts contiguously from zero and may interleave. Only start events carry
   `PartKind`, `MIMEType`, and `Name`.
2. Copy `DataDelta` if retaining it after the handler returns.
3. Return a handler's error unchanged and stop subsequent delivery. The handler
   call itself is observable progress even when it returns an error.
4. Use `model.ErrStreamingUnsupported` only for pre-stream mode rejection.
   Completion fallback is permissible only under the consumer's documented
   policy and when **no handler was called**. An arbitrary error, partial output,
   or handler error must not trigger an automatic second invocation.
5. Treat the returned `model.Response` as authoritative only when `err == nil`.
   Reasoning is transient and does not enter `Response.Message.Content`.

Agent stream events contain only transient reasoning/output text from any model
round. Concatenating them does not reconstruct the final `agent.Result` or
history. Tools, lifecycle facts, and host interaction have separate contracts.
A nil agent handler returns `agent.ErrNilStreamHandler`.

`agent.RoundInterceptor` (v0.2.2) runs around control/execution after a complete
model decision and before tool effects. Preserve SessionID, Index, Invocation,
and Response. `Decision` is the permitted modification surface; honor its
validation contract, and do not rewrite a persisted `RoundResult`.

Passive observation (v0.2.3) is not a control path. Observer work must not
synchronously gate execution. Context-carried observation correlation may
enrich telemetry; it cannot replace explicit business execution identity.

v0.2.4 clarified the error boundary: known business outcomes can be a valid tool
Result with `nil` error. A runtime's `tool.ErrNotFound` and
`tool.ErrInvalidArguments` are pre-dispatch rejections; it must not emit those
sentinels after dispatch. Other errors do not prove that effects never occurred.
Session Append errors likewise leave the caller uncertain whether the entry
committed; they are not a reason to blindly retry.

## Execution settlement and accounting (v0.2.5)

Before v0.2.5, agent Run/Stream returned `agent.Result` directly. They now return
`agent.Execution`. Consumers need both success and failure branches:

```go
func execute(ctx context.Context, runtime agent.Runtime, turn agent.Turn) (content.Content, error) {
    execution, err := runtime.Run(ctx, turn)
    if execution.Outcome.Status != 0 {
        // Record authoritative termination/accounting even when err != nil.
        // Choose an application-owned recorder; the SDK does not supply one.
    }
    if err != nil {
        return nil, err // Never consume execution.Result on failure.
    }
    if execution.Result == nil {
        return nil, fmt.Errorf("agent runtime returned success without a result")
    }
    return execution.Result.Output, nil
}
```

A zero `Execution` means lifecycle establishment did not occur. A started failed
or canceled Turn still returns a valid Outcome and no Result. Preserve the
ordinary Go error chain, including `context.Canceled` and
`context.DeadlineExceeded`; failure stages are diagnostic boundaries and do not
mean rollback or retry safety.

For runtime implementers, `agent.Interceptor` still returns `agent.Result`.
Settlement wraps the interceptor chain; do not change its generic result type
to `Execution`. Count started Round, Model Runtime, and canonical Tool Runtime
attempts. Aggregate only authoritative provider-reported execution usage.
`Usage.Reported` distinguishes a reported zero from absent usage; estimates from
`usage.Counter` are not provider execution accounting. Coverage must remain
Unavailable or Partial when the available facts do not establish Complete.

## Session lifecycle capabilities (v0.2.6)

Replace old creation and discovery calls:

```go
// Before: id, err := store.Create(ctx, session.Metadata{Title: "Example"})
meta, err := store.Create(ctx, session.CreateRequest{Title: "Example"})
if err != nil {
    return err
}
id := meta.ID // Storage assigns identity and timestamps.
```

Inject `session.Manager` for Get/Rename/Archive/Restore/Delete/Fork and
`session.Query` for `List(ctx) ([]session.Metadata, error)`. `Store` now exposes
only Create/Append/Load. Remove the old `MutableStore`, `Summary`, and pagination
`Query` value. The portable Query interface has no limit/offset arguments; a
host that needs pagination must apply a documented policy or a separate
capability.

Implementations must reject Append to archived sessions with `ErrArchived` but
allow their Load. Lifecycle operations do not update the last-entry-mutation
timestamp. Archive/Restore are idempotent; Rename changes only the title. Fork
creates a new active identity with opaque copied entries and new timestamps;
prevent concurrent source mutation if a deterministic fork boundary is required.
Delete does not imply secure erase or deletion of assets referenced by opaque
entries.

## Explicit invocation scope and workspace (v0.2.8)

`tool.Call` remains the durable model/history payload. The new runtime envelope
is `tool.Invocation`; do not add hidden Session fields to Call or recover them
from context values:

```go
// Before: result, err := tools.Call(ctx, call)
result, err := tools.Call(ctx, tool.Invocation{
    Scope: execution.Scope{SessionID: turn.SessionID},
    Call:  call,
})
```

Change every Tool implementation, Runtime implementation, interceptor, mock,
and adapter in the call path. Tool interceptors now use
`pipeline.Next[tool.Invocation, tool.Result]` and must forward the original
Scope unchanged. Scope is identity, not a service container or authorization
token. The receiver validates it; zero SessionID must be rejected unless that
capability explicitly permits an unbound call. Obtain real persisted IDs from
the Session provider when interacting with persistence.

Tools that emit Session-bound host effects inject `interaction.ExecutionBinder`:

```go
channel, err := binder.Bind(invocation.Scope)
if err != nil {
    return tool.Result{}, err
}
// Use channel.Request/Emit/Set/Clear with the invocation's context.
```

Bind is non-blocking and returns a channel whose routing cannot change. A
missing SessionID must wrap `interaction.ErrInvalidExecutionScope`.
`UnavailableBinder()` preserves this validation and cancellation behavior for
non-interactive hosts. Do not let observation context values override routing.

Workspace consumers use `workspace.Resolver` with explicit scope; application
mutation uses `workspace.Manager`. One Session maps to one immutable existing
local workspace binding. This capability neither creates worktrees nor provides
a security sandbox. Filesystem confinement and approval remain implementation
responsibilities.

## Live provider sources (v0.2.10)

The removed `model.Provider` / `StreamingProvider` interfaces must no longer
appear in component exports, dependency fields, adapters, or type assertions.
Existing implementation methods can still be used as bound function values:

```go
// Old export: Providers []ingotabi.Named[model.Provider]
// Current export:
type Exports struct {
    Source model.ProviderSource
}

type source struct {
    entry model.ProviderEntry // Immutable once constructed.
}

func (s source) Snapshot(ctx context.Context) ([]model.ProviderEntry, error) {
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    return []model.ProviderEntry{s.entry}, nil // Fresh caller-owned slice.
}

// Construction from an existing immutable provider implementation:
// source{entry: model.ProviderEntry{
//     Name: "primary", Complete: provider.Complete, Stream: provider.Stream,
// }}
// Set Stream to nil for a completion-only provider.
```

Consumers inject `[]model.ProviderSource`, snapshot them for an invocation, and
select by name. `Name` must be nonempty valid UTF-8 and unique across all sources;
reject duplicates instead of silently selecting the first entry. `Complete`
must be non-nil. A snapshot error makes its entire returned slice unusable; an
empty successful snapshot means no configured providers.

For configuration reloads, publish new immutable provider implementations and
their bound functions. Do not let functions already returned by Snapshot read
mutable shared credentials/endpoints on invocation. Old functions must remain
usable for the source's lifetime, including requests already in flight. Keep
each invocation on one selected snapshot; do not switch providers between
request resolution and completion/streaming because a reload happened.

The Component Graph is still static. Live sources permit data changes within
already-wired capabilities; they do not dynamically add graph dependencies.

## Child-agent and namespaced metadata contracts (v0.2.11)

Child support is optional. Inject `agent.Children` for child management and
implement `agent.ChildSessionRepository` only when the Session provider can
honor its persistence semantics. Adding SDK v0.2.11 alone does not enable child
tools or supply a scheduler.

- `CreateRequest.Depth` and `Meta` must be respected even when a provider's old
  Create method still compiles. Store creation metadata atomically, reject
  unrepresentable depths, and preserve all unowned top-level Meta namespaces.
  Copy retained and returned raw JSON bytes according to the ownership rules.
- Child creation atomically registers the parent relationship and new child
  record. Conditional updates must distinguish omitted values, zero values,
  and clearing nullable fields (`UpdateValue` / `UpdateNullable`). A false
  `updated` result means the condition did not match, not a storage failure.
- Each child represents exactly one Turn. Repeated conversational continuation
  is not this contract. Definition and task snapshots are immutable for that
  child; future configuration changes apply to future children.
- Every management operation receives explicit `execution.Scope` and must
  authorize the caller from it. Session identity alone is not an authorization
  grant. Honor `ErrChildUnauthorized`, capacity, unsupported, invalid-state,
  and result-submission errors without inventing a global dispatcher.
- Check never waits. Wait only waits for an execution object owned by the
  current process; without one it reads persisted state once and returns.
  A caller's canceled wait is distinct from an explicit Cancel request.
- Persisted State, StateConfirmed, nullable ExecutionStopped, and Result are
  separate facts. Unknown state is not failure, Cancel may return while work
  responds to cancellation, and a confirmed Completed result is the only
  authoritative child result.
- Recover process-local states left by the previous runtime using the repository
  contract; do not claim external writers have stopped merely because the new
  process has no in-memory execution object.
- An optional workspace binding references an existing directory. Neither these
  contracts nor cancellation create/remove worktrees or roll back their effects.

Preserve plugin-owned persistence formats and follow the concrete Session
provider's migration/backup procedure. Core image rollback alone cannot reverse
a plugin's data migration.

## Verify the migration

Run the SDK race-enabled suite and each affected consumer's own tests.
Consumer checks should exercise
actual constructor exports/dependencies, scope propagation, concurrent calls,
retained input/output ownership, handler failure, context cancellation, and
persisted old/new data as applicable. For a release, repeat consumer checks on
the exact downloadable tagged module graph with `GOWORK=off`; local replacement
success is useful development evidence but is not publication evidence.
