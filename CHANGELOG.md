# Changelog

This record describes the source changes at the repository's actual `v0.x.y`
tags. It was reconstructed from tagged source and diffs on 2026-09-22; the 19
tagged commits were also checked against GitHub's remote tag refs. This is not
a claim that every tag is available from every Go proxy. Check the exact version
with `GOWORK=off go mod download -json github.com/ingot-agent/sdk@v0.2.11`
before depending on it in a release. Local branch names and README design
milestones are not module versions.

The SDK is still `v0`: incompatible changes have occurred in patch releases.
Read all intervening entries and the [migration guide](docs/MIGRATIONS.md), pin
the version in each consuming module, and test the complete selected module
graph. The Go language version declared by these tags is Go 1.24.0.

## Unreleased

## v0.2.13 — fork target metadata

- Added `session.ForkRequest.Meta` so session implementations can persist
  target-owned metadata atomically when creating a fork. It is not inherited
  from the source session.
- Added this version-based history and a migration guide.
- Clarified direct imports of domain contract modules and documentation
  ownership. These documentation changes do not change public SDK
  contracts.

## v0.2.11 — single-Turn child agents

Source commit: `1952ec6`.

- Added `agent.Children`, child definitions, state/snapshot types, pagination,
  cancellation results, diagnostics, and child-specific sentinel errors.
- Added `agent.ChildSessionRepository` for atomic parent/child registration,
  conditional metadata updates, branch transitions, and process recovery.
- Added namespaced `session.Meta`, `Metadata.Meta`, and `CreateRequest.Depth`
  and `CreateRequest.Meta`. Session implementations must preserve namespaces
  they do not own and atomically persist creation metadata.
- Child management is an optional capability; importing the SDK does not supply
  a scheduler, persistence provider, workspace manager, or tool implementation.
  Existing Session providers need a semantic review even if they still compile.

## v0.2.10 — live model provider sources (breaking)

Source commit: `fb6d5ef`.

- Removed `model.Provider` and `model.StreamingProvider`.
- Added `model.ProviderSource.Snapshot` and `model.ProviderEntry` with required
  `Complete` and optional `Stream` functions. `model.Runtime` and
  `model.StreamingRuntime` remain the invocation boundaries.
- Each snapshot has deterministic order and caller-owned storage. Previously
  returned functions retain their immutable configuration and remain usable
  for the source's lifetime after entries are replaced or removed.
- Consumers must reject duplicate provider names across sources and retain one
  snapshot per invocation. See the [provider migration](docs/MIGRATIONS.md#live-provider-sources-v0210).

## v0.2.9 — nested interactions and operation grouping

Source commit: `ba7963d`.

- Added object/list request fields, recursive `interaction.Value` values, and
  `ObjectValue` / `ListValue` helpers. Hosts must handle nested values and fields
  recursively when supporting them.
- Added `operation.Definition.Group` as a presentation hint. It does not change
  the operation's identity, which remains `Name`; an empty Group is valid.

## v0.2.8 — explicit execution scope and workspace (breaking)

Source commit: `1941f51`.

- Added `execution.Scope` with explicit Session identity and `tool.Invocation`
  containing `Scope` plus the durable `tool.Call` payload.
- Changed `tool.Tool.Invoke`, `tool.Runtime.Call`, and `tool.Interceptor` to
  receive `Invocation`. Implementations must preserve its scope unchanged.
- Added `interaction.ExecutionBinder`, `UnavailableBinder`, and
  `ErrInvalidExecutionScope` for immutable execution-bound host interaction.
- Added session-scoped `workspace.Resolver` and `workspace.Manager` contracts.
  A workspace binding does not imply sandboxing or filesystem authorization.

## v0.2.7 — externally invocable operations

Source commit: `f6ab072`.

- Added `operation` definitions, invocations, results, and error contracts.
  Inputs/outputs are JSON objects, schemas use Draft 2020-12, and each
  invocation has an explicit host interaction channel.
- The host owns transport syntax, authentication, authorization, and routing.
  The SDK does not supply an operation registry or service locator.

## v0.2.6 — Session lifecycle management (breaking)

Source commit: `48ef49f`.

- Changed `session.Store.Create` from `(Metadata) (ID, error)` to
  `(CreateRequest) (Metadata, error)`; storage owns IDs and timestamps.
- Removed `MutableStore`, `Summary`, the old pagination `Query` struct, and
  `Store.List`. Added separate `Manager` and `Query` interfaces.
- Added get, rename, archive, restore, delete, and fork contracts, authoritative
  lifecycle metadata, and `ErrArchived`. Archived sessions remain readable;
  appending to them fails.

## v0.2.5 — execution outcome accounting (breaking)

Source commit: `3722317`.

- Changed `agent.Runtime.Run` and `agent.StreamingRuntime.Stream` to return
  `agent.Execution`, with an optional canonical `Result` and an `Outcome` that
  remains valid on failure after Turn lifecycle establishment.
- Added failure stages, duration, invocation accounting, usage coverage, and
  per-provider/model accounting.
- Added `model.Usage.Reported` and authoritative response Provider/Model fields.
  Observation terminal events now carry execution outcomes.
- `agent.Interceptor` still wraps `Turn` → `Result`; it did not change to
  `Execution`. The runtime owns settlement/accounting outside that chain.

## v0.2.4 — stabilized execution semantics (breaking)

Source commit: `c6b414b`. This is the former README **v0.3 design milestone**, not
a `v0.3.0` module release.

- Removed model capability metadata contracts (`CapabilityRequest`,
  `ContentCapability`, `Capabilities`, `CapabilityResolver`,
  `CapabilityProvider`, and `ErrCapabilitiesUnavailable`).
- Removed `agent.ErrStreamingUnsupported` and made `agent.StreamingRuntime`
  independent of `agent.Runtime`. Agent implementations may complete through a
  non-streaming model path and may succeed with no stream events.
- Clarified authoritative results, side-effect uncertainty, pre-dispatch tool
  errors, and Session append settlement. An error does not establish retry
  safety; completed side effects are not rolled back.

## v0.2.3 — passive execution observations

Source commit: `f82e655`.

- Added the `observation` package for correlated Turn/Round/Model/Tool lifecycle
  facts, event ordering, context-carried tracing, and passive consumers.
- Added transient `tool.Progress`. Progress is not a partial authoritative
  `tool.Result` and need not appear in the final result.

## v0.2.2 — agent round execution

Source commit: `fbe9ba7`.

- Added `agent.Round`, `RoundResult`, and `RoundInterceptor` plus round
  validation/limit sentinels. Round interception observes the complete model
  decision before tool effects, separates immutable execution facts from the
  mutable proposed decision, and carries persisted tool results.

## v0.2.1 — agent output streaming

Source commit: `830a7c8`.

- Added agent reasoning/output deltas, stream handlers, streaming runtime, and
  nil-handler/unsupported-streaming errors. Subsequent changes in v0.2.4 and
  v0.2.5 modify these initial contracts.
- Added independent model content/reasoning stream semantics. Reasoning is
  transient and excluded from the canonical response content.

## v0.2.0 — multimodal content (breaking)

Source commit: `61e4b74`.

- Added ordered `content.Content`, binary `asset` contracts, URI/inline/asset
  sources, agent attachments, validation, cloning, and text conversion helpers.
- Replaced text-only model/tool/prompt/agent content with ordered content parts.
  Model streaming changed to part start/delta/end events.
- Changed `session.Entry.Payload` from a string to opaque `[]byte`.
- Introduced model capability metadata contracts later removed in v0.2.4.

## v0.1.6 — separate runtime ABI (breaking)

Source commit: `ba18973`.

- Removed the root `sdk` package, `application`, and `config`. Component
  primitives and host contracts moved to
  [`github.com/ingot-agent/ingot-abi`](https://github.com/ingot-agent/ingot-abi).
- SDK capabilities became explicitly optional and independent of the Builder.
  Runtime configuration decoding and context-based process/state lookup are
  no longer SDK responsibilities. See the
  [ABI migration](docs/MIGRATIONS.md#runtime-contracts-and-the-abi-v016).

## v0.1.5 — structured host effects (breaking)

Source commit: `38a426c`.

- Replaced `interaction.Channel.Ask` / `Render` and their terminal-oriented
  request/event forms with structured `Request`, `Emit`, `Set`, and `Clear`
  effects. Added typed fields/values and an unavailable-host implementation.

## v0.1.4 — usage and request resolution

Source commit: `bf4f673` (annotated tag).

- Added model-aware input counting in `usage`, with explicit accuracy, and
  `model.RequestResolver` for materializing provider/model defaults without
  calling the provider or executing model interceptors.

## v0.1.3 — mutable Session titles

Source commit: `39885cb` (annotated tag).

- Added `session.MutableStore.Rename` with ordering and timestamp semantics.
  This interface was replaced by `session.Manager` in v0.2.6.

## v0.1.2 — process controls and agent history

Source commit: `b86a423`.

- Added `application.Process` and context helpers, later removed in v0.1.6.
- Added `agent.History` for validated persisted model messages.

## v0.1.1 — interaction input cleanup (breaking)

Source commit: `abb0e0d`.

- Removed `interaction.Channel.ReadLine`. At this version, free-form input used
  `Ask` with no options. That API was itself replaced in v0.1.5.

## v0.1.0 — first tagged contract set

Source commit: `4f78610`.

- Included component primitives, runtime configuration helpers, and the initial
  pipeline, HTTP, model, tool, Session, prompt, agent, and interaction contracts.
- Added context compaction contracts in `contextwindow`.
- Later `v0.1.x` and `v0.2.x` entries above supersede several of these contracts.

## Maintaining this record

For a release, compare actual tag trees (`git diff OLD_TAG NEW_TAG --`), inspect
the public declarations and their comments, and record source commits. Some
existing tags refer to pre-squash development commits; the tagged source tree,
not branch ancestry or a pull-request title, determines the version's API.
Do not assign a module version to an untagged design milestone. Move an
Unreleased entry into a tagged entry only after verifying that exact tag.
