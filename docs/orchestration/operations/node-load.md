# node.load.get

Get load averages (1-minute, 5-minute, and 15-minute).

## Usage

```go
task := plan.Task("get-load", &orchestrator.Op{
    Operation: "node.load.get",
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
