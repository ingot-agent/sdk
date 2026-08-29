# Contributing to the ingot SDK

Thank you for helping improve the shared contracts used by the ingot plugin
ecosystem. Contributions of code, tests, documentation, bug reports, and design
feedback are welcome.

## Before you start

- Search existing issues and pull requests before opening a new one.
- Small fixes and documentation improvements can go directly to a pull
  request.
- Discuss new capabilities, public API changes, and breaking semantic changes
  with the maintainers before investing in an implementation.
- Do not report security vulnerabilities in a public issue. Use the
  repository's private security reporting channel, or contact a maintainer
  privately if that channel is unavailable.

## Decide whether the contract belongs here

This SDK is optional and domain-neutral. A proposal belongs in this repository
only when it is useful across materially different agent domains and can be
implemented without depending on one plugin, product, UI, provider, or
industry.

Before proposing a new capability, answer these questions:

- Do multiple unrelated agent domains need the same semantics?
- Are there at least two plausible interchangeable implementations?
- Can cancellation, concurrency, ordering, ownership, lifecycle, and errors be
  specified precisely?
- Will the contract remain independent of a concrete UI or workflow?
- Is a shared public dependency preferable to a plugin-local type?

If the answer is no, keep the contract in the owning plugin. When multiple
plugins in one domain need to interoperate, create a separately versioned
domain SDK that ingot can configure alongside this one.

Do not add UI widgets, rendering or layout models, terminal behavior,
frontend-framework state, coding workflows, customer-service schemas, or other
domain-specific protocols to this SDK.

## Development workflow

### 1. Create a branch

Never commit or push directly to `main`. Start from an up-to-date `main` and
create a focused branch:

```sh
git switch main
git pull --ff-only
git switch -c feat/short-description
```

### 2. Make a focused change

- Keep each pull request limited to one coherent purpose.
- Prefer small capability packages and interfaces over broad abstractions.
- Keep the SDK independent of ingot core and concrete plugin implementations.
- Avoid new dependencies unless a public contract genuinely requires them.
- Update package documentation and the README when public behavior or SDK scope
  changes.
- Do not include secrets, local configuration, IDE metadata, build artifacts,
  or unrelated formatting changes.

Format every changed Go file with `gofmt`:

```sh
gofmt -w path/to/changed_file.go
```

### 3. Test the complete module

Targeted tests are useful while developing, but they do not replace the full
race-enabled suite required before every commit:

```sh
go test -race ./...
git diff --check
```

If required validation cannot run or does not pass, document the blocker and
do not present the change as ready to merge.

### 4. Commit clearly

Use concise, imperative commit subjects consistent with the repository's
existing style:

```text
feat(model): add a general request contract
fix(session): preserve append ordering
docs: clarify SDK scope
```

Keep commits reviewable and do not rewrite shared branch history without
coordinating with other contributors.

### 5. Open a pull request

Open the pull request against `main` and include:

- what changed and why;
- why any new capability is domain-neutral;
- relevant issue links;
- tests and validation performed;
- compatibility or migration impact;
- known limitations or follow-up work.

All required CI checks must pass. Avoid force-pushing after review has started
unless it has been coordinated with the reviewers.

## Public contract expectations

- Document every exported identifier and its observable semantics.
- Preserve `context.Context` cancellation and deadlines across blocking calls.
- Define concurrency and ordering guarantees explicitly.
- Treat aggregate inputs as immutable; return caller-owned aggregates unless
  documented otherwise.
- Preserve ordinary error chains for `errors.Is` and `errors.As`.
- Add positive, negative, concurrency, and ordering tests where relevant.
- Prefer external-package tests that model independent plugin authors.
- Treat incompatible exported API or semantic changes as breaking changes.

## License

By submitting a contribution, you agree that it may be distributed under the
repository's [MIT License](./LICENSE).
