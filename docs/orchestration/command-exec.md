# command.exec.execute

Execute a command directly on the target node.

## Usage

```go
task := plan.Task("install-nginx", &orchestrator.Op{
    Operation: "command.exec.execute",
    Target:    "_all",
    Params: map[string]any{
        "command": "apt",
        "args":    []string{"install", "-y", "nginx"},
    },
})
```

## Parameters

| Param     | Type     | Required | Description            |
| --------- | -------- | -------- | ---------------------- |
| `command` | string   | Yes      | The command to execute |
| `args`    | []string | No       | Command arguments      |

## Target

Accepts any valid target: `_any`, `_all`, a hostname, or a label selector
(`key:value`).

## Idempotency

**Not idempotent.** Always returns `Changed: true`. Use guards (`OnlyIfChanged`,
`When`) to control execution.

## Permissions

Requires `command:execute` permission.
