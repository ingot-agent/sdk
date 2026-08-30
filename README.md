# ingot SDK

> Optional, domain-neutral Go contracts for interoperable agent plugins.

`github.com/ingot-agent/sdk` is the shared public protocol library maintained
for the ingot plugin ecosystem. It provides general-purpose agent capability
contracts. It does not discover plugins, resolve component graphs, generate
wiring, load plugins at runtime, or provide concrete agent implementations.

Most importantly, **this SDK is useful, but it is not mandatory**. ingot is a
build-time composition system, not a framework tied to one protocol package.
Plugins are encouraged to use these standard contracts when they fit because
shared capabilities make independently developed plugins interoperable. A
plugin may also use its own types, use another SDK, or define a new SDK when the
standard contracts are not appropriate.

## Optional by design

The official SDK is one possible contract vocabulary for a component graph. It
is not a privileged plugin API and importing it is not what makes a module an
ingot plugin.

A plugin author can choose among three approaches:

1. **Use the standard SDK** for common capabilities such as models, tools,
   sessions, prompts, and agent execution.
2. **Mix standard and domain contracts** by using this SDK for general
   capabilities and a separate SDK for domain-specific ones.
3. **Use a different SDK entirely** when the plugin ecosystem needs different
   foundational or capability contracts.

Agent capabilities are ordinary Go types. The Builder participates in the
Component Graph purely through Go type identity: no SDK module path needs to
be configured for this module, or for any other contract module, to be used
by a plugin.

## Scope: general agent capabilities only

This repository is for capabilities that remain meaningful across unrelated
agent domains. “General” does not mean every agent must use every package. It
means a contract can be understood and implemented without depending on one
product, industry, interface, workflow, or provider.

A contract belongs here when it is:

- useful across materially different kinds of agents;
- independent of a specific UI, transport product, model vendor, or business
  domain;
- small enough to have precise ownership, concurrency, ordering, and error
  semantics;
- implementable by multiple interchangeable plugins; and
- stable enough to become a shared ecosystem dependency.

The following do **not** belong in this SDK:

- UI component models, rendering or layout protocols, terminal key bindings,
  browser events, and framework-specific frontend state;
- coding-agent-only editor operations or repository workflows;
- customer-service ticket schemas, commerce order models, healthcare records,
  or other industry-specific data;
- one provider's request fields, one product's policy language, or one
  application's orchestration state;
- runtime-owned host protocols: Component ABI primitives, process control,
  invocation metadata and plugin state scope (those live in the fixed
  [ingot ABI](https://github.com/ingot-agent/ingot-abi)).

Put a private contract next to the plugin that owns it. If several plugins in a
domain need to interoperate, publish a separate domain SDK and let plugins
import it directly. Presentation-neutral semantic effects may be shared; the UI
protocol that renders those effects should remain outside this module.

## Packages

Every package is independently consumable; a plugin only imports the contracts
it uses.

| Package | Purpose |
|---|---|
| `pipeline` | Generic typed interceptor composition. |
| `httpx` | A shared, context-aware HTTP client capability. |
| `filesystem` | Safe workspace-relative filesystem access. |
| `tool` | Tool definitions, invocation, runtime lookup, and interception. |
| `model` | Model providers, complete and streaming runtimes, request resolution, and interception. |
| `session` | Ordered session persistence and mutable session metadata. |
| `prompt` | Prompt contribution and rendering. |
| `contextwindow` | Model-context compaction. |
| `usage` | Model-aware input counting with explicit accuracy. |
| `agent` | Agent turn execution, history access, and interception. |
| `interaction` | Presentation-neutral structured effects between plugins and a host environment. |

`interaction` describes semantic requests, events, and state. It deliberately
does not define widgets, layouts, terminal behavior, or rendering.

The Component ABI (`Cleanup`, `Optional`, `Named`) and the runtime host
contracts (invocation, lifecycle, state scope) live in the
[ingot ABI](https://github.com/ingot-agent/ingot-abi), not in this module.

## Migrating runtime contracts

Runtime-facing contracts previously published by this module have moved to
`github.com/ingot-agent/ingot-abi`:

- replace imports of the root SDK package for `Cleanup`, `Optional`, and
  `Named` with the root `ingot-abi` package;
- split `application.Process` into explicit `ingot-abi/invocation.Invocation`
  and `ingot-abi/lifecycle.Controller` dependencies;
- replace `config.StateDir` context lookup with an explicit
  `ingot-abi/state.Scope` dependency; and
- remove calls to `config.ResolveTables` and `config.Decode`; the Builder and
  generated Runtime Image own strict decoding and pass the component's typed
  configuration value to its constructor.

The ABI repository is fixed infrastructure shared by the Builder and generated
Runtime Images. This SDK remains an optional collection of agent capability
contracts.

## Using the standard SDK

Requires Go 1.24 or newer.

```sh
go get github.com/ingot-agent/sdk@latest
go get github.com/ingot-agent/ingot-abi@v0.1.0
```

Capabilities are ordinary Go interfaces and values. A component declares only
what it consumes and provides:

```go
import (
	"context"

	ingotabi "github.com/ingot-agent/ingot-abi"
	"github.com/ingot-agent/sdk/agent"
	"github.com/ingot-agent/sdk/model"
	"github.com/ingot-agent/sdk/session"
	"github.com/ingot-agent/sdk/tool"
)

type Config struct{}

type Dependencies struct {
	Model model.Runtime
	Tools tool.Runtime
	Store session.Store
}

type Exports struct {
	Agent agent.Runtime
}

func New(
	ctx context.Context,
	cfg Config,
	deps Dependencies,
) (Exports, ingotabi.Cleanup, error) {
	return Exports{}, nil, nil
}
```

There is no service locator or global registration API. The ingot Builder reads
the component's types, resolves the static graph, and generates the constructor
calls.

## Contract conventions

All packages follow the same baseline rules unless a more specific contract
says otherwise:

- Blocking operations accept `context.Context`; that argument owns
  cancellation and deadline authority.
- Capability implementations are safe for concurrent use unless the contract
  defines a narrower ordering rule.
- Aggregate inputs are immutable by contract. A callee copies mutable data it
  retains, and ownership of aggregate outputs passes to the caller on return.
- Errors use ordinary Go error chains and preserve context and package sentinel
  errors through `errors.Is` and `errors.As`.
- Ordering is explicit and deterministic wherever it affects observable
  behavior.
- Contracts describe semantics, not a particular implementation, provider, or
  presentation.

Package documentation is the source of truth for capability-specific behavior.

## Evolving the SDK

The SDK is a shared dependency, so additions should be conservative:

1. Demonstrate that the capability is useful outside one plugin and one domain.
2. Define its lifecycle, ownership, concurrency, ordering, cancellation, and
   error semantics.
3. Prefer a small independent package or interface over a broad framework.
4. Add external-package contract tests that model third-party use.
5. Treat incompatible public API or semantic changes as breaking changes.

When a proposal fails the domain-neutrality test, create a separate SDK instead
of expanding this one.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for scope checks, development workflow,
testing requirements, and pull request expectations.

## Development

Run the complete conformance suite with the race detector:

```sh
go test -race ./...
```

## License

[MIT](./LICENSE)
