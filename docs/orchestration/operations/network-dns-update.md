# network.dns.update

Update DNS servers for a network interface.

## Usage

```go
task := plan.Task("update-dns", &orchestrator.Op{
    Operation: "network.dns.update",
    Target:    "_all",
    Params: map[string]any{
        "interface": "eth0",
        "servers":   []string{"8.8.8.8", "8.8.4.4"},
    },
})
```

## Parameters

| Param | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `interface` | string | Yes | Network interface name |
| `servers` | []string | Yes | DNS server addresses |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label
selector (`key:value`).

## Idempotency

**Idempotent.** Checks current DNS servers before mutating. Returns
`Changed: true` only if the servers were actually updated. Returns
`Changed: false` if the servers already match the desired state.

## Permissions

Requires `network:write` permission.
