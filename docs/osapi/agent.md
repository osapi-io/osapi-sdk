# AgentService

Agent discovery and details.

## Methods

| Method | Description |
| ------ | ----------- |
| `List(ctx)` | Retrieve all active agents |
| `Get(ctx, hostname)` | Get detailed agent info by hostname |

## Usage

```go
// List all agents
resp, err := client.Agent.List(ctx)

// Get specific agent details
resp, err := client.Agent.Get(ctx, "web-01")
```

## Permissions

Requires `agent:read` permission.
