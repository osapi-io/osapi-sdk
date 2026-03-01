# Orchestration

The `orchestrator` package provides DAG-based task orchestration on top of the
OSAPI SDK client. Define tasks with dependencies and the library handles
execution order, parallelism, conditional logic, and reporting.

## Operations

Operations are the building blocks of orchestration plans. Each operation maps
to an OSAPI job type that agents execute.

| Operation                                     | Description            | Idempotent | Category |
| --------------------------------------------- | ---------------------- | ---------- | -------- |
| [`command.exec.execute`](command-exec.md)     | Execute a command      | No         | Command  |
| [`command.shell.execute`](command-shell.md)   | Execute a shell string | No         | Command  |
| [`network.dns.get`](network-dns-get.md)       | Get DNS configuration  | Read-only  | Network  |
| [`network.dns.update`](network-dns-update.md) | Update DNS servers     | Yes        | Network  |
| [`network.ping.do`](network-ping.md)          | Ping a host            | Read-only  | Network  |
| [`node.hostname.get`](node-hostname.md)       | Get system hostname    | Read-only  | Node     |
| [`node.status.get`](node-status.md)           | Get node status        | Read-only  | Node     |
| [`node.disk.get`](node-disk.md)               | Get disk usage         | Read-only  | Node     |
| [`node.memory.get`](node-memory.md)           | Get memory stats       | Read-only  | Node     |
| [`node.load.get`](node-load.md)               | Get load averages      | Read-only  | Node     |

### Idempotency

- **Read-only** operations never modify state and always return
  `Changed: false`.
- **Idempotent** write operations check current state before mutating and return
  `Changed: true` only if something actually changed.
- **Non-idempotent** operations (command exec/shell) always return
  `Changed: true`. Use guards (`When`, `OnlyIfChanged`) to control when they
  run.

## Hooks

Register callbacks to control logging and progress at every stage:

```go
hooks := orchestrator.Hooks{
    BeforePlan:  func(explain string) { fmt.Print(explain) },
    AfterPlan:   func(report *orchestrator.Report) { fmt.Println(report.Summary()) },
    BeforeLevel: func(level int, tasks []*orchestrator.Task, parallel bool) { ... },
    AfterLevel:  func(level int, results []orchestrator.TaskResult) { ... },
    BeforeTask:  func(task *orchestrator.Task) { ... },
    AfterTask:   func(task *orchestrator.Task, result orchestrator.TaskResult) { ... },
    OnRetry:     func(task *orchestrator.Task, attempt int, err error) { ... },
    OnSkip:      func(task *orchestrator.Task, reason string) { ... },
}

plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))
```

The SDK performs no logging -- hooks are the only output mechanism. Consumers
bring their own formatting.

## Error Strategies

| Strategy            | Behavior                                        |
| ------------------- | ----------------------------------------------- |
| `StopAll` (default) | Fail fast, cancel everything                    |
| `Continue`          | Skip dependents, keep running independent tasks |
| `Retry(n)`          | Retry n times before failing                    |

Strategies can be set at plan level or overridden per-task:

```go
plan := orchestrator.NewPlan(client, orchestrator.OnError(orchestrator.Continue))
task.OnError(orchestrator.Retry(3)) // override for this task
```

## Adding a New Operation

When a new operation is added to OSAPI:

1. Create `docs/orchestration/{name}.md` following the template of existing
   operation docs
2. Add a row to the operations table in this README
3. Add the operation to `examples/all/main.go`
4. Update `CLAUDE.md` package structure if new files were added
