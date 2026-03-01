# MetricsService

Prometheus metrics access.

## Methods

| Method     | Description                       |
| ---------- | --------------------------------- |
| `Get(ctx)` | Fetch raw Prometheus metrics text |

## Usage

```go
text, err := client.Metrics.Get(ctx)
fmt.Print(text)
```

## Permissions

Unauthenticated. The `/metrics` endpoint is open.
