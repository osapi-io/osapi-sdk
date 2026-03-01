# node.hostname.get

Get the system hostname and agent labels.

## Usage

```go
task := plan.Task("get-hostname", &orchestrator.Op{
    Operation: "node.hostname.get",
    Target:    "_any",
})
```

## Parameters

None.

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label
selector (`key:value`).

## Idempotency

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `node:read` permission.
