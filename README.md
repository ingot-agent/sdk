# ingot SDK

`github.com/ingot-agent/sdk` is the public Go contract layer for ingot component
graphs. It contains composition primitives and stable capability types; it does
not perform plugin discovery, graph resolution, static wiring, or provide the
official component implementations.

The packages follow the SDK v0.1 design:

- `sdk`: `Cleanup`, `Optional[T]`, and `Named[T]`;
- `config`: strict plugin TOML decoding and plugin-scoped state directories;
- `pipeline`: generic typed interceptor composition;
- `httpx` and `filesystem`: shared infrastructure capabilities;
- `tool`, `model`, `session`, `prompt`, `contextwindow`, `interaction`, and
  `agent`: domain contracts and runtime chokepoints;
- `application`: immutable process invocation metadata and controlled,
  cleanup-preserving shutdown requests for interactive frontends.

## Contract rules

- Blocking operations take `context.Context`; that argument owns cancellation
  and deadline authority.
- Capability implementations are concurrent-safe unless a domain contract
  defines a narrower ordering rule.
- Aggregate inputs are immutable-by-contract. A callee that retains mutable
  input data copies it first; ownership of aggregate outputs passes to the
  caller when the call returns.
- Errors use ordinary Go chains, preserving context and package sentinel errors
  through `errors.Is`.
- `pipeline.Compose` treats the first interceptor as the outermost interceptor.
- `interaction.Channel` exposes `Ask` for synchronous user input and `Render`
  for output. `Ask` with empty `Options` is a plain free-form text question;
  with `Options` it presents ordered choices, and `AllowTextInput` adds a
  free-form choice. `AskResponse.Text` contains either the selected label or
  the original custom text.
- `contextwindow.Compactor` receives an immutable model invocation and returns
  a complete caller-owned message sequence. Implementations may keep
  session-scoped compaction state, but the SDK does not prescribe a summary,
  checkpoint, or token-counting strategy.
- `agent.History` exposes a validated, caller-owned snapshot of persisted model
  messages without making frontends understand an Agent implementation's
  private session payload format.
- `application.Process` is supplied through Context by the generated runtime.
  Frontends use it to inspect arguments/check mode and request normal or fatal
  shutdown; the generated runtime remains responsible for cancellation and
  reverse cleanup.

The module requires Go 1.24 or newer because the public API uses generic type
aliases as specified by the design.

Run the conformance tests with:

```sh
go test -race ./...
```
