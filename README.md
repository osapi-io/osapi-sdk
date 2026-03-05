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

## 📦 Install

```bash
go get github.com/osapi-io/osapi-sdk
```

## 🔌 SDK Client

Typed Go client for every OSAPI endpoint. See the
[client docs](docs/osapi/README.md) for targeting, options, and quick
start.

| Service | Operations                                            | Docs                          | Source                               |
| ------- | ----------------------------------------------------- | ----------------------------- | ------------------------------------ |
| Node    | Hostname, disk, memory, load, uptime, OS info, status | [docs](docs/osapi/node.md)    | [`node.go`](pkg/osapi/node.go)       |
| Network | DNS get/update, ping                                  | [docs](docs/osapi/node.md)    | [`node.go`](pkg/osapi/node.go)       |
| Command | exec, shell                                           | [docs](docs/osapi/node.md)    | [`node.go`](pkg/osapi/node.go)       |
| Job     | Create, get, list, delete, retry, stats               | [docs](docs/osapi/job.md)     | [`job.go`](pkg/osapi/job.go)         |
| Agent   | List, get (discovery + heartbeat data)                | [docs](docs/osapi/agent.md)   | [`agent.go`](pkg/osapi/agent.go)     |
| Health  | Liveness, readiness, status                           | [docs](docs/osapi/health.md)  | [`health.go`](pkg/osapi/health.go)   |
| Audit   | List, get, export                                     | [docs](docs/osapi/audit.md)   | [`audit.go`](pkg/osapi/audit.go)     |
| Metrics | Prometheus text                                       | [docs](docs/osapi/metrics.md) | [`metrics.go`](pkg/osapi/metrics.go) |

## 🔀 Orchestration

DAG-based task execution on top of the client. See the
[orchestration docs](docs/orchestration/README.md) for hooks, error
strategies, and adding new operations.

| Feature             | Description                                                               | Source                                      |
| ------------------- | ------------------------------------------------------------------------- | ------------------------------------------- |
| DAG execution       | Dependency-based ordering with automatic parallelism                      | [`plan.go`](pkg/orchestrator/plan.go)       |
| Op tasks            | Declarative OSAPI operations with target routing and params               | [`task.go`](pkg/orchestrator/task.go)       |
| TaskFunc            | Custom Go functions with SDK client access                                | [`task.go`](pkg/orchestrator/task.go)       |
| TaskFuncWithResults | Custom functions that receive completed results from prior tasks          | [`task.go`](pkg/orchestrator/task.go)       |
| Hooks               | 8 lifecycle callbacks — consumer-controlled logging                       | [`options.go`](pkg/orchestrator/options.go) |
| Error strategies    | StopAll, Continue (skip dependents), Retry(n)                             | [`options.go`](pkg/orchestrator/options.go) |
| Guards              | `When` predicates, `OnlyIfChanged` conditional execution                  | [`task.go`](pkg/orchestrator/task.go)       |
| Changed detection   | Read-only ops return false, mutators reflect actual state                 | [`runner.go`](pkg/orchestrator/runner.go)   |
| Result access       | `Results.Get()` for cross-task data flow, `Status` for outcome inspection | [`result.go`](pkg/orchestrator/result.go)   |
| Broadcast results   | Per-host `HostResult` data for multi-target operations                    | [`result.go`](pkg/orchestrator/result.go)   |

### Operations

| Operation               | Description            | Idempotent | Docs                                             |
| ----------------------- | ---------------------- | ---------- | ------------------------------------------------ |
| `node.hostname.get`     | Get system hostname    | Read-only  | [docs](docs/orchestration/node-hostname.md)      |
| `node.status.get`       | Get node status        | Read-only  | [docs](docs/orchestration/node-status.md)        |
| `node.disk.get`         | Get disk usage         | Read-only  | [docs](docs/orchestration/node-disk.md)          |
| `node.memory.get`       | Get memory stats       | Read-only  | [docs](docs/orchestration/node-memory.md)        |
| `node.uptime.get`       | Get system uptime      | Read-only  | [docs](docs/orchestration/node-uptime.md)        |
| `node.load.get`         | Get load averages      | Read-only  | [docs](docs/orchestration/node-load.md)          |
| `network.dns.get`       | Get DNS configuration  | Read-only  | [docs](docs/orchestration/network-dns-get.md)    |
| `network.dns.update`    | Update DNS servers     | Yes        | [docs](docs/orchestration/network-dns-update.md) |
| `network.ping.do`       | Ping a host            | Read-only  | [docs](docs/orchestration/network-ping.md)       |
| `command.exec.execute`  | Execute a command      | No         | [docs](docs/orchestration/command-exec.md)       |
| `command.shell.execute` | Execute a shell string | No         | [docs](docs/orchestration/command-shell.md)      |

## 📋 Examples

Each example is a standalone Go program you can read and run.

| Example                                 | What it shows                                                                                         |
| --------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| [discovery](examples/discovery/main.go) | Runnable DAG that discovers fleet info: health check, agent listing, and status in parallel           |
| [all](examples/all/main.go)             | Every feature: hooks, Op tasks, TaskFunc, dependencies, guards, Levels(), error strategies, reporting |

```bash
cd examples/discovery
OSAPI_TOKEN="<jwt>" go run main.go
```

## 📖 Documentation

See the [generated documentation][] for package-level API details.

## 🤝 Contributing

See the [Development](docs/development.md) guide for prerequisites, setup,
and conventions. See the [Contributing](docs/contributing.md) guide before
submitting a PR.

## 📄 License

The [MIT][] License.

[OSAPI]: https://github.com/osapi-io/osapi
[generated documentation]: docs/gen/
[MIT]: LICENSE
