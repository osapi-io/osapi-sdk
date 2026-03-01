# node.disk.get

Get disk usage statistics for all mounted filesystems.

## Usage

```go
task := plan.Task("get-disk", &orchestrator.Op{
    Operation: "node.disk.get",
    Target:    "_any",
})
```

## Parameters

None.

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Read-only.** Never modifies state. Always returns `Changed: false`.

## Permissions

Requires `node:read` permission.
