# JobService

Async job queue operations.

## Methods

| Method | Description |
| ------ | ----------- |
| `Create(ctx, operation, target)` | Create a new job |
| `Get(ctx, id)` | Retrieve a job by UUID |
| `List(ctx, params)` | List jobs with optional filters |
| `Delete(ctx, id)` | Delete a job by UUID |
| `Retry(ctx, id, target)` | Retry a failed job |
| `QueueStats(ctx)` | Retrieve queue statistics |

## Usage

```go
// Create a job
resp, err := client.Job.Create(ctx, map[string]any{
    "type":   "node.hostname.get",
    "params": map[string]any{},
}, "_any")

// List completed jobs
resp, err := client.Job.List(ctx, osapi.ListParams{
    Status: "completed",
    Limit:  20,
})

// Retry a failed job
resp, err := client.Job.Retry(ctx, "uuid-string", "_any")
```

## Permissions

Read operations require `job:read`. Write operations (create, delete,
retry) require `job:write`.
