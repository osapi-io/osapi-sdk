# Orchestrator Phase 2 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Add hooks, structured DAG access, full error strategy support,
operation documentation, and a comprehensive example to the orchestrator
package.

**Architecture:** All changes are in `pkg/orchestrator/` inside osapi-sdk. Hooks
provide consumer-controlled callbacks for logging and progress. `Levels()`
exposes the levelized DAG for programmatic access. Error strategies (`Continue`,
`Retry`, per-task `OnError`) get wired into the runner. Operation docs catalog
available OSAPI operations for orchestration authors.

**Tech Stack:** Go 1.25, osapi-sdk, testify/suite

**Repo:** `osapi-io/osapi-sdk` (orchestrator worktree)

---

## Task 1: Hooks Type and WithHooks Option

**Files:**

- Modify: `pkg/orchestrator/options.go`
- Modify: `pkg/orchestrator/options_public_test.go`

### Step 1: Write the failing test

Add to `pkg/orchestrator/options_public_test.go`:

```go
func (s *OptionsSuite) TestWithHooks() {
	called := false
	hooks := orchestrator.Hooks{
		BeforeTask: func(_ string) {
			called = true
		},
	}

	cfg := orchestrator.PlanConfig{}
	opt := orchestrator.WithHooks(hooks)
	opt(&cfg)

	s.NotNil(cfg.Hooks)
	s.NotNil(cfg.Hooks.BeforeTask)
	cfg.Hooks.BeforeTask("test")
	s.True(called)
}

func (s *OptionsSuite) TestHooksDefaults() {
	h := orchestrator.Hooks{}

	// Nil callbacks should be safe — no panic.
	s.Nil(h.BeforePlan)
	s.Nil(h.BeforeLevel)
	s.Nil(h.BeforeTask)
	s.Nil(h.AfterTask)
	s.Nil(h.AfterPlan)
}
```

### Step 2: Run test to verify it fails

```bash
go test -run TestOptionsSuite/TestWithHooks -v ./pkg/orchestrator/...
```

Expected: FAIL — `Hooks` type not defined.

### Step 3: Implement Hooks type and WithHooks option

Add to `pkg/orchestrator/options.go`:

```go
// Hooks provides consumer-controlled callbacks for plan execution
// events. All fields are optional — nil callbacks are skipped.
type Hooks struct {
	BeforePlan  func(explain string)
	BeforeLevel func(level int, names []string, parallel bool)
	BeforeTask  func(name string)
	AfterTask   func(result TaskResult)
	AfterPlan   func(report *Report)
}
```

Add `Hooks` field to `PlanConfig`:

```go
type PlanConfig struct {
	OnErrorStrategy ErrorStrategy
	Verbose         bool
	Output          io.Writer
	Hooks           *Hooks
}
```

Add `WithHooks` option function:

```go
func WithHooks(
	hooks Hooks,
) PlanOption {
	return func(cfg *PlanConfig) {
		cfg.Hooks = &hooks
	}
}
```

### Step 4: Run test to verify it passes

```bash
go test -run TestOptionsSuite -v ./pkg/orchestrator/...
```

Expected: PASS

### Step 5: Commit

```bash
git add pkg/orchestrator/options.go pkg/orchestrator/options_public_test.go
git commit -m "feat(orchestrator): add Hooks type and WithHooks option"
```

---

## Task 2: Levels() Method

**Files:**

- Modify: `pkg/orchestrator/plan.go`
- Modify: `pkg/orchestrator/plan_public_test.go`

### Step 1: Write the failing test

Add to `pkg/orchestrator/plan_public_test.go`:

```go
func (s *PlanSuite) TestLevels() {
	plan := orchestrator.NewPlan(nil)

	a := plan.Task("a", &orchestrator.Op{Operation: "node.hostname.get", Target: "_any"})
	b := plan.Task("b", &orchestrator.Op{Operation: "node.disk.get", Target: "_any"})
	c := plan.Task("c", &orchestrator.Op{Operation: "node.load.get", Target: "_any"})

	b.DependsOn(a)
	c.DependsOn(a)

	levels, err := plan.Levels()
	s.NoError(err)
	s.Len(levels, 2)

	// Level 0: a
	s.Len(levels[0], 1)
	s.Equal("a", levels[0][0].Name())

	// Level 1: b and c (parallel)
	s.Len(levels[1], 2)
	names := []string{levels[1][0].Name(), levels[1][1].Name()}
	s.ElementsMatch([]string{"b", "c"}, names)
}

func (s *PlanSuite) TestLevelsValidationError() {
	plan := orchestrator.NewPlan(nil)
	a := plan.Task("a", &orchestrator.Op{Operation: "test", Target: "_any"})
	b := plan.Task("b", &orchestrator.Op{Operation: "test", Target: "_any"})
	a.DependsOn(b)
	b.DependsOn(a)

	_, err := plan.Levels()
	s.Error(err)
}
```

### Step 2: Run test to verify it fails

```bash
go test -run TestPlanSuite/TestLevels -v ./pkg/orchestrator/...
```

Expected: FAIL — `Levels()` method not defined.

### Step 3: Implement Levels()

Add to `pkg/orchestrator/plan.go`:

```go
// Levels returns the levelized DAG — tasks grouped into execution
// levels where all tasks in a level can run concurrently.
// Returns an error if the plan fails validation.
func (p *Plan) Levels() ([][]*Task, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	return levelize(p.tasks), nil
}
```

### Step 4: Refactor Explain() to use Levels()

Replace the direct `levelize()` call in `Explain()`:

```go
func (p *Plan) Explain() string {
	levels, err := p.Levels()
	if err != nil {
		return fmt.Sprintf("invalid plan: %v", err)
	}

	var b strings.Builder

	fmt.Fprintf(&b, "Plan: %d tasks, %d levels\n", len(p.tasks), len(levels))

	// ... rest unchanged ...
```

### Step 5: Run tests to verify everything passes

```bash
go test -v ./pkg/orchestrator/...
```

Expected: ALL PASS (including existing Explain tests).

### Step 6: Commit

```bash
git add pkg/orchestrator/plan.go pkg/orchestrator/plan_public_test.go
git commit -m "feat(orchestrator): add Levels() for structured DAG access"
```

---

## Task 3: Wire Hooks into Runner and Refactor WithVerbose

**Files:**

- Modify: `pkg/orchestrator/runner.go`
- Modify: `pkg/orchestrator/options.go`
- Modify: `pkg/orchestrator/plan_integration_test.go`

### Step 1: Write the failing test

Add to `pkg/orchestrator/plan_integration_test.go`:

```go
func (s *PlanIntegrationSuite) TestHooksCalledDuringRun() {
	var events []string
	hooks := orchestrator.Hooks{
		BeforePlan: func(_ string) {
			events = append(events, "before-plan")
		},
		BeforeLevel: func(level int, _ []string, _ bool) {
			events = append(events, fmt.Sprintf("before-level-%d", level))
		},
		BeforeTask: func(name string) {
			events = append(events, "before-"+name)
		},
		AfterTask: func(r orchestrator.TaskResult) {
			events = append(events, "after-"+r.Name)
		},
		AfterPlan: func(_ *orchestrator.Report) {
			events = append(events, "after-plan")
		},
	}

	plan := orchestrator.NewPlan(nil, orchestrator.WithHooks(hooks))

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})

	plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	}).DependsOn(a)

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.NotNil(report)

	s.Equal([]string{
		"before-plan",
		"before-level-0",
		"before-a",
		"after-a",
		"before-level-1",
		"before-b",
		"after-b",
		"after-plan",
	}, events)
}
```

### Step 2: Run test to verify it fails

```bash
go test -run TestPlanIntegrationSuite/TestHooksCalledDuringRun -v ./pkg/orchestrator/...
```

Expected: FAIL — hooks not called in runner.

### Step 3: Add hook helper methods to runner

Add to `pkg/orchestrator/runner.go`:

```go
func (r *runner) hook() *Hooks {
	return r.plan.config.Hooks
}

func (r *runner) callBeforePlan(
	explain string,
) {
	if h := r.hook(); h != nil && h.BeforePlan != nil {
		h.BeforePlan(explain)
	}
}

func (r *runner) callBeforeLevel(
	level int,
	names []string,
	parallel bool,
) {
	if h := r.hook(); h != nil && h.BeforeLevel != nil {
		h.BeforeLevel(level, names, parallel)
	}
}

func (r *runner) callBeforeTask(
	name string,
) {
	if h := r.hook(); h != nil && h.BeforeTask != nil {
		h.BeforeTask(name)
	}
}

func (r *runner) callAfterTask(
	result TaskResult,
) {
	if h := r.hook(); h != nil && h.AfterTask != nil {
		h.AfterTask(result)
	}
}

func (r *runner) callAfterPlan(
	report *Report,
) {
	if h := r.hook(); h != nil && h.AfterPlan != nil {
		h.AfterPlan(report)
	}
}
```

### Step 4: Wire hooks into runner.run()

Replace the verbose logging calls in `run()`, `runLevel()`, and `runTask()` with
hook calls:

**In `run()`:**

```go
func (r *runner) run(
	ctx context.Context,
) (*Report, error) {
	start := time.Now()
	levels := levelize(r.plan.tasks)

	// Hooks + verbose
	r.callBeforePlan(r.plan.Explain())
	if r.plan.config.Verbose {
		r.log("%s", r.plan.Explain())
	}

	var taskResults []TaskResult

	for i, level := range levels {
		names := make([]string, len(level))
		for j, t := range level {
			names[j] = t.name
		}
		parallel := len(level) > 1

		r.callBeforeLevel(i, names, parallel)

		if r.plan.config.Verbose {
			if parallel {
				r.log("--- Level %d: %s (parallel)\n", i, strings.Join(names, ", "))
			} else {
				r.log("--- Level %d: %s\n", i, names[0])
			}
		}

		results, err := r.runLevel(ctx, level)
		taskResults = append(taskResults, results...)

		if err != nil {
			report := &Report{Tasks: taskResults, Duration: time.Since(start)}
			r.callAfterPlan(report)
			return report, err
		}
	}

	report := &Report{Tasks: taskResults, Duration: time.Since(start)}
	r.callAfterPlan(report)
	return report, nil
}
```

**In `runTask()`** — add `callBeforeTask` before execution and `callAfterTask`
before returning each `TaskResult`. Every return path must call `callAfterTask`:

```go
// Before execution (after guards pass):
r.callBeforeTask(t.name)

// Before every return of a TaskResult:
tr := TaskResult{...}
r.callAfterTask(tr)
return tr
```

Note: `callAfterTask` is called for ALL outcomes (changed, unchanged, skipped,
failed) so consumers can track every task.

### Step 5: Refactor WithVerbose to use hooks internally

Add a `defaultVerboseHooks` factory to `pkg/orchestrator/options.go`:

```go
func defaultVerboseHooks(
	w io.Writer,
) Hooks {
	return Hooks{
		BeforePlan: func(explain string) {
			fmt.Fprint(w, explain)
		},
		BeforeLevel: func(level int, names []string, parallel bool) {
			if parallel {
				fmt.Fprintf(w, "--- Level %d: %s (parallel)\n", level, strings.Join(names, ", "))
			} else {
				fmt.Fprintf(w, "--- Level %d: %s\n", level, names[0])
			}
		},
		BeforeTask: func(name string) {
			fmt.Fprintf(w, "    %-20s running...\n", name)
		},
		AfterTask: func(result TaskResult) {
			switch result.Status {
			case StatusSkipped:
				fmt.Fprintf(w, "    %-20s skipped (%s)\n", result.Name, result.Duration)
			case StatusFailed:
				fmt.Fprintf(w, "    %-20s FAILED (%s)\n", result.Name, result.Duration)
			default:
				fmt.Fprintf(w, "    %-20s %s (%s)\n", result.Name, result.Status, result.Duration)
			}
		},
	}
}
```

Then update `WithVerbose()` to set hooks:

```go
func WithVerbose() PlanOption {
	return func(cfg *PlanConfig) {
		cfg.Verbose = true
		if cfg.Output == nil {
			cfg.Output = os.Stdout
		}
		cfg.Hooks = ptr(defaultVerboseHooks(cfg.Output))
	}
}

func ptr[T any](v T) *T {
	return &v
}
```

**Precedence rule:** If the consumer sets both `WithVerbose()` and
`WithHooks()`, the last option wins (standard functional options). The consumer
should use `WithHooks()` when they want full control.

### Step 6: Remove old `r.log()` calls from runner

After hooks are wired, remove the inline `if r.plan.config.Verbose` blocks and
the `r.log()` calls from `run()` and `runTask()`. The verbose hooks now produce
the same output via callbacks. Keep the `log()` method for any future use but
remove all call sites.

### Step 7: Run full test suite

```bash
go test -v ./pkg/orchestrator/...
```

Expected: ALL PASS — existing verbose tests still produce the same output via
hooks.

### Step 8: Commit

```bash
git add pkg/orchestrator/runner.go pkg/orchestrator/options.go \
       pkg/orchestrator/plan_integration_test.go
git commit -m "feat(orchestrator): wire hooks into runner, refactor verbose"
```

---

## Task 4: Continue Error Strategy

**Files:**

- Modify: `pkg/orchestrator/runner.go`
- Modify: `pkg/orchestrator/plan_integration_test.go`

### Step 1: Write the failing test

Add to `pkg/orchestrator/plan_integration_test.go`:

```go
func (s *PlanIntegrationSuite) TestContinueStrategy() {
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Continue),
	)

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("a failed")
	})

	plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	}).DependsOn(a)

	// c is independent of a
	plan.TaskFunc("c", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err) // Continue doesn't return error

	s.Len(report.Tasks, 3)

	results := make(map[string]orchestrator.Status)
	for _, r := range report.Tasks {
		results[r.Name] = r.Status
	}

	s.Equal(orchestrator.StatusFailed, results["a"])
	s.Equal(orchestrator.StatusSkipped, results["b"]) // dependent skipped
	s.Equal(orchestrator.StatusChanged, results["c"]) // independent runs
}
```

### Step 2: Run test to verify it fails

```bash
go test -run TestPlanIntegrationSuite/TestContinueStrategy -v ./pkg/orchestrator/...
```

Expected: FAIL — `b` still runs instead of being skipped.

### Step 3: Add failed tracking to runner

Add `failed` map to `runner` struct:

```go
type runner struct {
	plan    *Plan
	results Results
	failed  map[string]bool
	mu      sync.Mutex
}

func newRunner(
	plan *Plan,
) *runner {
	return &runner{
		plan:    plan,
		results: make(Results),
		failed:  make(map[string]bool),
	}
}
```

### Step 4: Check for failed dependencies in runTask()

Add at the top of `runTask()`, before the `requiresChange` check:

```go
// Skip if any dependency failed (for Continue strategy).
r.mu.Lock()
for _, dep := range t.deps {
	if r.failed[dep.name] {
		r.mu.Unlock()

		if verbose {
			r.log("    %-20s skipped (dependency failed)\n", t.name)
		}

		tr := TaskResult{
			Name:     t.name,
			Status:   StatusSkipped,
			Duration: time.Since(start),
		}
		r.callAfterTask(tr)

		return tr
	}
}
r.mu.Unlock()
```

### Step 5: Mark failed tasks in runTask()

On task failure, add to `failed` map:

```go
if err != nil {
	r.mu.Lock()
	r.failed[t.name] = true
	r.mu.Unlock()

	// ... existing failure handling ...
}
```

### Step 6: Run tests

```bash
go test -v ./pkg/orchestrator/...
```

Expected: ALL PASS

### Step 7: Commit

```bash
git add pkg/orchestrator/runner.go pkg/orchestrator/plan_integration_test.go
git commit -m "feat(orchestrator): implement Continue error strategy"
```

---

## Task 5: Retry Error Strategy

**Files:**

- Modify: `pkg/orchestrator/runner.go`
- Modify: `pkg/orchestrator/plan_integration_test.go`

### Step 1: Write the failing test

Add to `pkg/orchestrator/plan_integration_test.go`:

```go
func (s *PlanIntegrationSuite) TestRetryStrategy() {
	attempts := 0

	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Retry(2)),
	)

	plan.TaskFunc("flaky", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		attempts++
		if attempts < 3 {
			return nil, fmt.Errorf("attempt %d failed", attempts)
		}

		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.Equal(3, attempts) // 1 initial + 2 retries
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestRetryExhausted() {
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Retry(1)),
	)

	plan.TaskFunc("always-fail", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("permanent failure")
	})

	report, err := plan.Run(context.Background())
	s.Error(err) // StopAll is implicit after retries exhausted
	s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
}
```

### Step 2: Run test to verify it fails

```bash
go test -run TestPlanIntegrationSuite/TestRetryStrategy -v ./pkg/orchestrator/...
```

Expected: FAIL — no retry loop.

### Step 3: Add effectiveStrategy helper

Add to `pkg/orchestrator/runner.go`:

```go
func (r *runner) effectiveStrategy(
	t *Task,
) ErrorStrategy {
	if t.errorStrategy != nil {
		return *t.errorStrategy
	}

	return r.plan.config.OnErrorStrategy
}
```

### Step 4: Add retry loop in runTask()

Wrap the execution block in `runTask()`:

```go
strategy := r.effectiveStrategy(t)
maxAttempts := 1
if strategy.kind == "retry" {
	maxAttempts = strategy.retryCount + 1
}

var result *Result
var err error

for attempt := range maxAttempts {
	if t.fn != nil {
		result, err = t.fn(ctx, client)
	} else {
		result, err = r.executeOp(ctx, t.op)
	}

	if err == nil {
		break
	}

	if verbose && attempt < maxAttempts-1 {
		r.log("    %-20s retry %d/%d\n", t.name, attempt+1, strategy.retryCount)
	}
}
```

### Step 5: Run tests

```bash
go test -v ./pkg/orchestrator/...
```

Expected: ALL PASS

### Step 6: Commit

```bash
git add pkg/orchestrator/runner.go pkg/orchestrator/plan_integration_test.go
git commit -m "feat(orchestrator): implement Retry error strategy"
```

---

## Task 6: Per-Task OnError

**Files:**

- Modify: `pkg/orchestrator/runner.go`
- Modify: `pkg/orchestrator/plan_integration_test.go`

### Step 1: Write the failing test

Add to `pkg/orchestrator/plan_integration_test.go`:

```go
func (s *PlanIntegrationSuite) TestPerTaskOnError() {
	plan := orchestrator.NewPlan(nil) // default StopAll

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("a failed")
	})
	a.OnError(orchestrator.Continue) // override: keep going

	plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err) // Continue on a, so no error returned

	results := make(map[string]orchestrator.Status)
	for _, r := range report.Tasks {
		results[r.Name] = r.Status
	}

	s.Equal(orchestrator.StatusFailed, results["a"])
	s.Equal(orchestrator.StatusChanged, results["b"])
}
```

### Step 2: Run test to verify it fails

```bash
go test -run TestPlanIntegrationSuite/TestPerTaskOnError -v ./pkg/orchestrator/...
```

Expected: FAIL — `a` fails with StopAll, `b` never runs.

### Step 3: Use effectiveStrategy in runLevel()

Update `runLevel()` to check per-task strategy:

```go
for _, tr := range results {
	if tr.Status == StatusFailed {
		// Find the task to check its strategy.
		strategy := r.plan.config.OnErrorStrategy
		for _, t := range tasks {
			if t.name == tr.Name && t.errorStrategy != nil {
				strategy = *t.errorStrategy

				break
			}
		}

		if strategy.kind == "stop_all" {
			return results, tr.Error
		}
	}
}
```

### Step 4: Run tests

```bash
go test -v ./pkg/orchestrator/...
```

Expected: ALL PASS

### Step 5: Commit

```bash
git add pkg/orchestrator/runner.go pkg/orchestrator/plan_integration_test.go
git commit -m "feat(orchestrator): implement per-task OnError override"
```

---

## Task 7: Operation Documentation

**Files:**

- Create: `docs/orchestration/README.md`
- Create: `docs/orchestration/operations/command-exec.md`
- Create: `docs/orchestration/operations/command-shell.md`
- Create: `docs/orchestration/operations/network-dns-get.md`
- Create: `docs/orchestration/operations/network-dns-update.md`
- Create: `docs/orchestration/operations/network-ping.md`
- Create: `docs/orchestration/operations/node-hostname.md`
- Create: `docs/orchestration/operations/node-status.md`
- Create: `docs/orchestration/operations/node-disk.md`
- Create: `docs/orchestration/operations/node-memory.md`
- Create: `docs/orchestration/operations/node-load.md`

### Step 1: Create the orchestration README

Create `docs/orchestration/README.md`:

````markdown
# Orchestration

The `orchestrator` package provides DAG-based task orchestration on top of the
OSAPI SDK client. Define tasks with dependencies and the library handles
execution order, parallelism, conditional logic, and reporting.

## Operations

Operations are the building blocks of orchestration plans. Each operation maps
to an OSAPI job type that agents execute.

| Operation                                                | Description                    | Idempotent | Category |
| -------------------------------------------------------- | ------------------------------ | ---------- | -------- |
| [`command.exec.execute`](operations/command-exec.md)     | Execute a command directly     | No         | Command  |
| [`command.shell.execute`](operations/command-shell.md)   | Execute a shell command string | No         | Command  |
| [`network.dns.get`](operations/network-dns-get.md)       | Get DNS configuration          | Read-only  | Network  |
| [`network.dns.update`](operations/network-dns-update.md) | Update DNS servers             | Yes        | Network  |
| [`network.ping.do`](operations/network-ping.md)          | Ping a host                    | Read-only  | Network  |
| [`node.hostname.get`](operations/node-hostname.md)       | Get system hostname            | Read-only  | Node     |
| [`node.status.get`](operations/node-status.md)           | Get comprehensive node status  | Read-only  | Node     |
| [`node.disk.get`](operations/node-disk.md)               | Get disk usage                 | Read-only  | Node     |
| [`node.memory.get`](operations/node-memory.md)           | Get memory statistics          | Read-only  | Node     |
| [`node.load.get`](operations/node-load.md)               | Get load averages              | Read-only  | Node     |

### Idempotency

- **Read-only** operations never modify state and always return
  `Changed: false`.
- **Idempotent** write operations check current state before mutating and return
  `Changed: true` only if something actually changed.
- **Non-idempotent** operations (command exec/shell) always return
  `Changed: true`. Use guards (`When`, `OnlyIfChanged`) to control when they
  run.

## Hooks

Register callbacks to control logging and progress reporting:

​```go hooks := orchestrator.Hooks{ BeforePlan: func(explain string) {
fmt.Print(explain) }, BeforeLevel: func(level int, names []string, parallel
bool) { ... }, BeforeTask: func(name string) { log.Printf("starting %s", name)
}, AfterTask: func(r orchestrator.TaskResult) { log.Printf("%s: %s", r.Name,
r.Status) }, AfterPlan: func(report \*orchestrator.Report) {
fmt.Println(report.Summary()) }, }

plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks)) ​```

## Error Strategies

| Strategy            | Behavior                                        |
| ------------------- | ----------------------------------------------- |
| `StopAll` (default) | Fail fast, cancel everything                    |
| `Continue`          | Skip dependents, keep running independent tasks |
| `Retry(n)`          | Retry n times before failing                    |

Strategies can be set at plan level or overridden per-task:

​`go plan := orchestrator.NewPlan(client, orchestrator.OnError(orchestrator.Continue)) task.OnError(orchestrator.Retry(3)) // override for this task ​`

## Adding a New Operation

When a new operation is added to OSAPI:

1. Create `docs/orchestration/operations/{name}.md` following the template below
2. Add a row to the operations table in this README
3. Add the operation to `examples/all/main.go`
4. Update `CLAUDE.md` package structure if new files were added
````

### Step 2: Create operation doc template

Each operation doc follows this template. Create all 10 files. Example for
`docs/orchestration/operations/command-exec.md`:

```markdown
# command.exec.execute

Execute a command directly on the target node.

## Usage

​`go task := plan.Task("install-nginx", &orchestrator.Op{     Operation: "command.exec.execute",     Target:    "_all",     Params: map[string]any{         "command": "apt",         "args":    []string{"install", "-y", "nginx"},     }, }) ​`

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
```

Follow this pattern for all operations. For read-only operations, the
idempotency section says "Read-only. Returns `Changed: false`." For
`network.dns.update`, note that it checks current state and returns
`Changed: false` if servers already match.

### Step 3: Commit

```bash
git add docs/orchestration/
git commit -m "docs(orchestrator): add operation catalog and documentation"
```

---

## Task 8: Comprehensive Example

**Files:**

- Create: `examples/all/main.go`
- Create: `examples/all/go.mod`
- Modify: `README.md`

### Step 1: Create `examples/all/main.go`

This example demonstrates every current feature: Op tasks, TaskFunc tasks,
dependencies, guards, hooks, error strategies, Levels(), Explain(), and Report.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
	"github.com/osapi-io/osapi-sdk/pkg/osapi"
)

func main() {
	token := os.Getenv("OSAPI_TOKEN")
	if token == "" {
		log.Fatal("OSAPI_TOKEN is required")
	}

	url := os.Getenv("OSAPI_URL")
	if url == "" {
		url = "http://localhost:8080"
	}

	client := osapi.New(
		osapi.WithBaseURL(url),
		osapi.WithToken(token),
	)

	// --- Hooks: consumer-controlled logging ---

	hooks := orchestrator.Hooks{
		BeforePlan: func(explain string) {
			fmt.Println("=== Execution Plan ===")
			fmt.Print(explain)
			fmt.Println()
		},
		BeforeLevel: func(level int, names []string, parallel bool) {
			label := ""
			if parallel {
				label = " (parallel)"
			}

			fmt.Printf(
				">>> Level %d%s: %s\n",
				level,
				label,
				strings.Join(names, ", "),
			)
		},
		BeforeTask: func(name string) {
			fmt.Printf("  [start] %s\n", name)
		},
		AfterTask: func(r orchestrator.TaskResult) {
			fmt.Printf(
				"  [%s] %s (%s)\n",
				r.Status,
				r.Name,
				r.Duration,
			)
		},
		AfterPlan: func(report *orchestrator.Report) {
			fmt.Printf(
				"\n=== Done: %s in %s ===\n",
				report.Summary(),
				report.Duration,
			)
		},
	}

	plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))

	// --- Task definitions ---

	// Level 0: health check (no deps)
	checkHealth := plan.TaskFunc(
		"check-health",
		func(
			ctx context.Context,
			c *osapi.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Health.Liveness(ctx)
			if err != nil {
				return nil, fmt.Errorf("health check: %w", err)
			}

			if resp.StatusCode() != 200 {
				return nil, fmt.Errorf(
					"API not healthy: %d",
					resp.StatusCode(),
				)
			}

			return &orchestrator.Result{Changed: false}, nil
		},
	)

	// Level 1: parallel queries (all depend on health)
	getHostname := plan.Task("get-hostname", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})
	getHostname.DependsOn(checkHealth)

	getDisk := plan.Task("get-disk", &orchestrator.Op{
		Operation: "node.disk.get",
		Target:    "_any",
	})
	getDisk.DependsOn(checkHealth)

	getMemory := plan.Task("get-memory", &orchestrator.Op{
		Operation: "node.memory.get",
		Target:    "_any",
	})
	getMemory.DependsOn(checkHealth)

	getLoad := plan.Task("get-load", &orchestrator.Op{
		Operation: "node.load.get",
		Target:    "_any",
	})
	getLoad.DependsOn(checkHealth)

	// Level 2: summary (depends on all queries, conditional)
	summary := plan.TaskFunc(
		"summary",
		func(
			_ context.Context,
			_ *osapi.Client,
		) (*orchestrator.Result, error) {
			fmt.Println("\n  --- Fleet Summary ---")
			fmt.Println("  All node queries completed successfully.")

			return &orchestrator.Result{Changed: false}, nil
		},
	)
	summary.DependsOn(getHostname, getDisk, getMemory, getLoad)

	// Guard: only print summary if hostname query succeeded.
	summary.When(func(results orchestrator.Results) bool {
		r := results.Get("get-hostname")

		return r != nil
	})

	// --- Structured DAG access ---

	levels, err := plan.Levels()
	if err != nil {
		log.Fatalf("invalid plan: %v", err)
	}

	fmt.Printf("DAG has %d levels:\n", len(levels))

	for i, level := range levels {
		names := make([]string, len(level))
		for j, t := range level {
			names[j] = t.Name()
		}

		fmt.Printf(
			"  Level %d: %s\n",
			i,
			strings.Join(names, ", "),
		)
	}

	fmt.Println()

	// --- Run ---

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatalf("plan failed: %v", err)
	}

	// --- Detailed results ---

	fmt.Println("\nDetailed results:")

	for _, r := range report.Tasks {
		fmt.Printf(
			"  %-20s status=%-10s changed=%-5v duration=%s\n",
			r.Name,
			r.Status,
			r.Changed,
			r.Duration,
		)
	}
}
```

### Step 2: Create `examples/all/go.mod`

```bash
cd examples/all
go mod init github.com/osapi-io/osapi-sdk/examples/all
```

Add replace directive for local development:

```
replace github.com/osapi-io/osapi-sdk => ../..
```

Then:

```bash
go mod tidy
```

### Step 3: Verify it compiles

```bash
go build ./...
```

Expected: compiles.

### Step 4: Update README.md examples table

Add the new example to the table in `README.md`:

```markdown
| [all](examples/all/main.go) | Every feature: Op tasks, TaskFunc, dependencies,
guards, hooks, Levels(), and reporting |
```

### Step 5: Commit

```bash
git add examples/all/ README.md
git commit -m "docs(orchestrator): add comprehensive example with hooks and DAG access"
```

---

## Task 9: CLAUDE.md Update and Final Verification

**Files:**

- Modify: `CLAUDE.md`

### Step 1: Update CLAUDE.md package structure

Update the orchestrator section to reflect new files:

```markdown
- **`pkg/orchestrator/`** - DAG-based task orchestration
  - `options.go` - ErrorStrategy, PlanOption, OnError, Hooks, WithHooks
  - `plan.go` - Plan, NewPlan, Validate, Run, Explain, Levels
  - `task.go` - Task, Op, TaskFn, DependsOn, When, OnError
  - `runner.go` - DAG resolution, parallel execution, job polling, hooks
  - `result.go` - Result, TaskResult, Report, Summary
```

### Step 2: Add orchestration docs section

Add after the package structure:

```markdown
## Orchestration Docs

Operation documentation lives in `docs/orchestration/`. When adding a new OSAPI
operation to the orchestrator:

1. Create `docs/orchestration/operations/{name}.md` following the existing
   template (params, target, idempotency, permissions)
2. Add a row to the operations table in `docs/orchestration/README.md`
3. Add the operation to `examples/all/main.go`
```

### Step 3: Run full test suite

```bash
just test
```

Expected: all tests pass, lint clean.

### Step 4: Verify all examples compile

```bash
for d in examples/*/; do (cd "$d" && go build ./...); done
```

Expected: all compile.

### Step 5: Commit

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with orchestrator phase 2 changes"
```
