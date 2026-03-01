# HealthService

Health check operations.

## Methods

| Method | Description |
| ------ | ----------- |
| `Liveness(ctx)` | Check if API server process is alive |
| `Ready(ctx)` | Check if server and dependencies are ready |
| `Status(ctx)` | Detailed system status (components, NATS, jobs) |

## Usage

```go
// Simple liveness check (unauthenticated)
resp, err := client.Health.Liveness(ctx)

// Readiness check (unauthenticated)
resp, err := client.Health.Ready(ctx)

// Detailed status (requires auth)
resp, err := client.Health.Status(ctx)
```

## Permissions

`Liveness` and `Ready` are unauthenticated. `Status` requires
`health:read` permission.
