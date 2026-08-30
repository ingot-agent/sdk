# Repository Instructions

## Project role

`github.com/ingot-agent/sdk` is an optional public contract library for the
ingot plugin ecosystem. It provides reusable, domain-neutral agent capability
contracts. It is not the ingot runtime, a plugin framework, or a mandatory
dependency for plugins.

Encourage plugins to use these contracts when they fit because shared types
improve interoperability. Never assume that every plugin must import this
module: plugins may use private contracts, another SDK, or a purpose-built SDK.
Contract modules participate in the Component Graph through ordinary Go type
identity only; no separate module configuration exists for them.

## Scope is domain-neutral

Only add capabilities that remain useful and meaningful across materially
different agent domains.

- Do not add a contract solely because one plugin or official implementation
  needs it.
- Do not add UI widgets, rendering or layout models, terminal behavior,
  frontend-framework state, or other presentation-specific protocols.
- Do not add coding, customer-service, commerce, healthcare, provider, or
  product-specific schemas and workflows.
- Presentation-neutral semantic effects are acceptable only when their
  contract does not prescribe how a host displays or collects them.
- Keep private concepts in the owning plugin. Put shared domain concepts in a
  separately versioned domain SDK that Plugins import directly.

A proposal belongs here only when it has precise, implementation-independent
semantics and more than one plausible interchangeable implementation.

## Public contract discipline

- Keep the module small and independent of ingot core and concrete plugins.
- Treat exported types, interfaces, sentinel errors, and documented semantics
  as compatibility commitments.
- All blocking operations accept `context.Context` and preserve cancellation
  and deadline errors.
- State concurrency, ordering, ownership, and lifecycle behavior explicitly.
- Aggregate inputs are immutable by contract; return caller-owned aggregate
  outputs unless documented otherwise.
- Preserve ordinary Go error chains for `errors.Is` and `errors.As`.
- Document every exported identifier accurately.
- Prefer small capability packages over broad abstractions or frameworks.
- Add external-package tests that model use by independent plugin authors.
- Treat incompatible API or semantic changes as breaking changes.

## Change discipline

- Treat current code, package documentation, and tests as the source of truth.
- Update the README whenever a change affects SDK scope, package inventory, or
  documented behavior.
- Format only changed Go files with `gofmt` and avoid unrelated rewrites.
- Do not add dependencies unless a public contract genuinely requires them.
- Preserve existing work and keep changes focused on the requested task.

## Git workflow

- Never edit, commit, or push directly on `main`. Use a focused working branch
  and submit changes through a pull request.
- Inspect the working tree before switching branches and preserve all existing
  uncommitted work.
- Do not create commits, push branches, force-push, rewrite shared history, or
  open pull requests unless the user explicitly requests that action.

## Required validation

Before every commit, run the complete race-enabled test suite:

```sh
go test -race ./...
git diff --check
```

Targeted tests are useful while iterating, but they do not replace the complete
suite. If required validation cannot run or does not pass, report the blocker
and do not describe the change as ready to commit or merge.
