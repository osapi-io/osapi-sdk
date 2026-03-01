# SDK Client

The `osapi` package provides a typed Go client for the OSAPI REST API.
Create a client with `New()` and use domain-specific services to
interact with the API.

## Quick Start

```go
client, err := osapi.New("http://localhost:8080", "your-jwt-token")
if err != nil {
    log.Fatal(err)
}

resp, err := client.Node.Hostname(ctx, "_any")
```

## Services

| Service | Description | Source |
| ------- | ----------- | ------ |
| [`Agent`](agent.md) | Agent discovery and details | `agent.go` |
| [`Audit`](audit.md) | Audit log operations | `audit.go` |
| [`Health`](health.md) | Health check operations | `health.go` |
| [`Job`](job.md) | Async job queue operations | `job.go` |
| [`Metrics`](metrics.md) | Prometheus metrics access | `metrics.go` |
| [`Node`](node.md) | Node management, network, commands | `node.go` |

## Client Options

| Option | Description |
| ------ | ----------- |
| `WithLogger(logger)` | Set custom `slog.Logger` (defaults to `slog.Default()`) |
| `WithHTTPTransport(transport)` | Set custom `http.RoundTripper` base transport |

## Targeting

Most operations accept a `target` parameter:

| Target | Behavior |
| ------ | -------- |
| `_any` | Send to any available agent (load balanced) |
| `_all` | Broadcast to every agent |
| `hostname` | Send to a specific host |
| `key:value` | Send to agents matching a label |
