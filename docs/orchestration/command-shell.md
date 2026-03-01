# command.shell.execute

Execute a shell command string on the target node. The command is
passed to `/bin/sh -c`.

## Usage

```go
task := plan.Task("check-disk-space", &orchestrator.Op{
    Operation: "command.shell.execute",
    Target:    "_any",
    Params: map[string]any{
        "command": "df -h / | tail -1 | awk '{print $5}'",
    },
})
```

## Parameters

| Param | Type | Required | Description |
| ----- | ---- | -------- | ----------- |
| `command` | string | Yes | The full shell command string |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label
selector (`key:value`).

## Idempotency

**Not idempotent.** Always returns `Changed: true`. Use guards
(`OnlyIfChanged`, `When`) to control execution.

## Permissions

Requires `command:execute` permission.
