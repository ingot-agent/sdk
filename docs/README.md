# SDK documentation

The SDK provides optional, domain-neutral capability contracts. It has no
runtime, plugin loader, provider implementation, or persistence implementation.

| Start here | What it covers |
| --- | --- |
| [Repository README](../README.md) | Scope, package inventory, dependency use, and contract conventions |
| [Changelog](../CHANGELOG.md) | Changes mapped to actual module tags through v0.2.11 |
| [Migration guide](MIGRATIONS.md) | Module selection and coordinated API, semantic, and persistence changes |
| [Contribution guide](../CONTRIBUTING.md) | Admission criteria, implementation expectations, and validation |
| [Security policy](../SECURITY.md) | Vulnerability reporting and current channel availability |
| [Design history](design-history/README.md) | Archived design context; not current API or release guarantees |

For capability-level details, use the Go package comments in this checkout,
for example `go doc ./tool`, `go doc ./model.ProviderSource`, and
`go doc ./agent.Children`. The comments and tests define ownership, concurrency,
ordering, cancellation, and error semantics. Consult the
[plugins repository](https://github.com/ingot-agent/plugins) for concrete
implementations, configuration, recipes, and plugin development. Runtime host
contracts belong to [ingot-abi](https://github.com/ingot-agent/ingot-abi).
