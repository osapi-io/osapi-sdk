[![release](https://img.shields.io/github/release/osapi-io/osapi-sdk.svg?style=for-the-badge)](https://github.com/osapi-io/osapi-sdk/releases/latest)
[![codecov](https://img.shields.io/codecov/c/github/osapi-io/osapi-sdk?style=for-the-badge)](https://codecov.io/gh/osapi-io/osapi-sdk)
[![go report card](https://goreportcard.com/badge/github.com/osapi-io/osapi-sdk?style=for-the-badge)](https://goreportcard.com/report/github.com/osapi-io/osapi-sdk)
[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE)
[![build](https://img.shields.io/github/actions/workflow/status/osapi-io/osapi-sdk/go.yml?style=for-the-badge)](https://github.com/osapi-io/osapi-sdk/actions/workflows/go.yml)
[![powered by](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=for-the-badge)](https://github.com/goreleaser)
[![conventional commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![built with just](https://img.shields.io/badge/Built_with-Just-black?style=for-the-badge&logo=just&logoColor=white)](https://just.systems)
![gitHub commit activity](https://img.shields.io/github/commit-activity/m/osapi-io/osapi-sdk?style=for-the-badge)

# OSAPI SDK

Go SDK for [OSAPI][] — client library and orchestration primitives.

## Install

```bash
go get github.com/osapi-io/osapi-sdk
```

## What's Inside

**Client library** (`pkg/osapi`) — typed Go client for every OSAPI
endpoint: node management, network config, command execution, job
control, health checks, audit logs, and metrics. Connect, authenticate,
and call any operation in a few lines.

**Orchestration** (`pkg/orchestrator`) — DAG-based task execution on top
of the client. Define tasks with dependencies and the library handles
execution order, parallelism, conditional logic (`OnlyIfChanged`,
`When`), error strategies (`StopAll`, `Continue`, `Retry`), and
reporting.

## Examples

Each example is a standalone Go program you can read and run.

| Example | What it shows |
| ------- | ------------- |
| [basic](examples/basic/main.go) | Connect to OSAPI, query a hostname, run a command, list audit entries, check health |
| [discovery](examples/discovery/main.go) | Runnable DAG that discovers fleet info: health check, agent listing, and status in parallel |
| [orchestrator](examples/orchestrator/main.go) | Declarative deployment DAG with dependencies and conditional execution |
| [all](examples/all/main.go) | Every feature: hooks, Op tasks, TaskFunc, dependencies, guards, Levels(), error strategies, reporting |

```bash
cd examples/discovery
OSAPI_TOKEN="<jwt>" go run main.go
```

## Documentation

See the [generated documentation][] for package-level API details.

## Contributing

See the [Development](docs/development.md) guide for prerequisites, setup,
and conventions. See the [Contributing](docs/contributing.md) guide before
submitting a PR.

## License

The [MIT][] License.

[OSAPI]: https://github.com/osapi-io/osapi
[generated documentation]: docs/gen/
[MIT]: LICENSE
