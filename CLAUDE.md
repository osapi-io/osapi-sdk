# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go SDK for OSAPI — client library and orchestration primitives. Used by osapi-io projects (linked via `replace` in consuming project's `go.mod`).

## Development Reference

For setup, prerequisites, and contributing guidelines:

- @docs/development.md - Prerequisites, setup, code style, testing, commits
- @docs/contributing.md - PR workflow and contribution guidelines

## Documentation

- @docs/osapi/README.md - SDK client overview, services, targeting
- @docs/orchestration/README.md - Orchestration overview, operations, hooks, error strategies
- @docs/gen/ - Auto-generated API reference (gomarkdoc)

## Quick Reference

```bash
just fetch / just deps / just test / just go::unit / just go::vet / just go::fmt
```

## Package Structure

- **`pkg/osapi/`** - Core SDK library
  - `osapi.go` - Client struct, New() factory, Option funcs
  - `transport.go` - HTTP transport with Bearer token and OTel tracing
  - `node.go` - NodeService (hostname, status, agents)
  - `network.go` - NetworkService (DNS get/update, ping)
  - `command.go` - CommandService (exec, shell)
  - `job.go` - JobService (create, get, list, delete, retry, stats)
  - `health.go` - HealthService (liveness, readiness, status)
  - `audit.go` - AuditService (list, get, export)
  - `metrics.go` - MetricsService (Prometheus text)
  - `gen/` - Generated OpenAPI client (`*.gen.go`)
- **`pkg/orchestrator/`** - DAG-based task orchestration
  - `plan.go` - Plan, NewPlan, Validate, Run, Explain, Levels
  - `task.go` - Task, Op, TaskFn, DependsOn, When, OnError
  - `runner.go` - DAG resolution, parallel execution, job polling
  - `result.go` - Result, TaskResult, Report, Summary
  - `options.go` - ErrorStrategy, Hooks, PlanOption, OnError, WithHooks

## Code Standards (MANDATORY)

### Function Signatures

ALL function signatures MUST use multi-line format:
```go
func FunctionName(
    param1 type1,
    param2 type2,
) (returnType, error) {
}
```

### Testing

- Public tests: `*_public_test.go` in test package (`package osapi_test`) for exported functions
- Internal tests: `*_test.go` in same package (`package osapi`) for private functions
- Use `testify/suite` with table-driven patterns

### Go Patterns

- Error wrapping: `fmt.Errorf("context: %w", err)`
- Early returns over nested if-else
- Unused parameters: rename to `_`
- Import order: stdlib, third-party, local (blank-line separated)

### Linting

golangci-lint with: errcheck, errname, goimports, govet, prealloc, predeclared, revive, staticcheck. Generated files (`*.gen.go`, `*.pb.go`) are excluded from formatting.

### Branching

See @docs/development.md#branching for full conventions.

When committing changes via `/commit`, create a feature branch first if
currently on `main`. Branch names use the pattern `type/short-description`
(e.g., `feat/add-dns-retry`, `fix/memory-leak`, `docs/update-readme`).

### Commit Messages

See @docs/development.md#commit-messages for full conventions.

Follow [Conventional Commits](https://www.conventionalcommits.org/) with the
50/72 rule. Format: `type(scope): description`.

When committing via Claude Code, end with:
- `🤖 Generated with [Claude Code](https://claude.ai/code)`
- `Co-Authored-By: Claude <noreply@anthropic.com>`
