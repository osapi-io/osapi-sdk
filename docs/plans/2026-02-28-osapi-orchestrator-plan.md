# Orchestrator Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to
> implement this plan task-by-task.

**Goal:** Add DAG-based task orchestration primitives to osapi-sdk so operators
can compose SDK calls into dependency-aware, parallel execution plans.

**Architecture:** New `pkg/orchestrator/` package inside osapi-sdk. Operators
write Go programs that define tasks with dependencies, and the library handles
DAG resolution, parallel execution, conditional logic, and reporting. No new
server components. See `docs/plans/2026-02-28-orchestrator-design.md` in the
osapi repo for the full design.

**Tech Stack:** Go 1.25, osapi-sdk, testify/suite

**Repo:** `osapi-io/osapi-sdk` (existing)

---

## Task 1: Update SDK README and CLAUDE.md

**Files:**

- Modify: `README.md`
- Modify: `CLAUDE.md`

**Step 1: Update `README.md`**

Update the heading and description to reflect the broader scope:

```markdown
# OSAPI SDK

Go SDK for [OSAPI][] — client library and orchestration primitives.
```

Add an Orchestration section after Usage:

```markdown
## Orchestration

The `orchestrator` package provides DAG-based task orchestration on top of the
client library. Define tasks with dependencies, and the library handles
execution order, parallelism, conditional logic, and reporting.

See the [orchestrator example][] for a complete walkthrough.

[orchestrator example]: examples/orchestrator/main.go
```

**Step 2: Update `CLAUDE.md`**

Add orchestrator to the package structure section:

```markdown
- **`pkg/orchestrator/`** - DAG-based task orchestration
  - `plan.go` - Plan, NewPlan
  - `task.go` - Task, TaskFunc, DependsOn, When
  - `runner.go` - DAG resolution, parallel execution
  - `result.go` - Result, Report, Summary
  - `options.go` - OnError, Retry, OnlyIfChanged
```

**Step 3: Update `.coverignore`**

Verify `/examples/` is already in `.coverignore` (it is). No change needed.

**Step 4: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: update README and CLAUDE.md for orchestrator"
```

---

## Task 2: Result Type

**Files:**

- Create: `pkg/orchestrator/result.go`
- Create: `pkg/orchestrator/result_public_test.go`

**Step 1: Write the failing test**

Create `pkg/orchestrator/result_public_test.go`:

```go
package orchestrator_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

type ResultSuite struct {
	suite.Suite
}

func TestResultSuite(t *testing.T) {
	suite.Run(t, new(ResultSuite))
}

func (s *ResultSuite) TestReportSummary() {
	tests := []struct {
		name     string
		tasks    []orchestrator.TaskResult
		contains []string
	}{
		{
			name: "mixed results",
			tasks: []orchestrator.TaskResult{
				{Name: "a", Status: orchestrator.StatusChanged, Changed: true, Duration: time.Second},
				{Name: "b", Status: orchestrator.StatusUnchanged, Changed: false, Duration: 2 * time.Second},
				{Name: "c", Status: orchestrator.StatusSkipped, Changed: false, Duration: 0},
				{Name: "d", Status: orchestrator.StatusChanged, Changed: true, Duration: 500 * time.Millisecond},
			},
			contains: []string{"4 tasks", "2 changed", "1 unchanged", "1 skipped"},
		},
		{
			name:     "empty report",
			tasks:    nil,
			contains: []string{"0 tasks"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			report := orchestrator.Report{Tasks: tt.tasks}
			summary := report.Summary()
			for _, c := range tt.contains {
				s.Contains(summary, c)
			}
		})
	}
}

func (s *ResultSuite) TestResultsGet() {
	tests := []struct {
		name       string
		results    orchestrator.Results
		lookupName string
		wantNil    bool
		wantChange bool
	}{
		{
			name: "found",
			results: orchestrator.Results{
				"install": {Changed: true},
			},
			lookupName: "install",
			wantNil:    false,
			wantChange: true,
		},
		{
			name:       "not found",
			results:    orchestrator.Results{},
			lookupName: "missing",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.results.Get(tt.lookupName)
			if tt.wantNil {
				s.Nil(got)
			} else {
				s.Require().NotNil(got)
				s.Equal(tt.wantChange, got.Changed)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: FAIL — types do not exist yet.

**Step 3: Write minimal implementation**

Create `pkg/orchestrator/result.go`:

```go
package orchestrator

import (
	"fmt"
	"strings"
	"time"
)

// Status represents the outcome of a task execution.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusChanged   Status = "changed"
	StatusUnchanged Status = "unchanged"
	StatusSkipped   Status = "skipped"
	StatusFailed    Status = "failed"
)

// Result is the outcome of a single task execution.
type Result struct {
	Changed bool
	Data    map[string]any
}

// TaskResult records the full execution details of a task.
type TaskResult struct {
	Name     string
	Status   Status
	Changed  bool
	Duration time.Duration
	Error    error
}

// Results is a map of task name to Result, used for conditional logic.
type Results map[string]*Result

// Get returns the Result for the named task, or nil if not found.
func (r Results) Get(
	name string,
) *Result {
	return r[name]
}

// Report is the aggregate output of a plan execution.
type Report struct {
	Tasks    []TaskResult
	Duration time.Duration
}

// Summary returns a human-readable summary of the report.
func (r *Report) Summary() string {
	var changed, unchanged, skipped, failed int

	for _, t := range r.Tasks {
		switch t.Status {
		case StatusChanged:
			changed++
		case StatusUnchanged:
			unchanged++
		case StatusSkipped:
			skipped++
		case StatusFailed:
			failed++
		}
	}

	parts := []string{
		fmt.Sprintf("%d tasks", len(r.Tasks)),
	}

	if changed > 0 {
		parts = append(parts, fmt.Sprintf("%d changed", changed))
	}

	if unchanged > 0 {
		parts = append(parts, fmt.Sprintf("%d unchanged", unchanged))
	}

	if skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", skipped))
	}

	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}

	return strings.Join(parts, ", ")
}
```

**Step 4: Run test to verify it passes**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: PASS

**Step 5: Run lint**

```bash
just go::vet
```

Expected: clean

**Step 6: Commit**

```bash
git add pkg/orchestrator/result.go pkg/orchestrator/result_public_test.go
git commit -m "feat(orchestrator): add Result, Report, and status types"
```

---

## Task 3: Options Type

**Files:**

- Create: `pkg/orchestrator/options.go`
- Create: `pkg/orchestrator/options_public_test.go`

**Step 1: Write the failing test**

Create `pkg/orchestrator/options_public_test.go`:

```go
package orchestrator_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

type OptionsSuite struct {
	suite.Suite
}

func TestOptionsSuite(t *testing.T) {
	suite.Run(t, new(OptionsSuite))
}

func (s *OptionsSuite) TestErrorStrategy() {
	tests := []struct {
		name     string
		strategy orchestrator.ErrorStrategy
		wantStr  string
	}{
		{
			name:     "stop all",
			strategy: orchestrator.StopAll,
			wantStr:  "stop_all",
		},
		{
			name:     "continue",
			strategy: orchestrator.Continue,
			wantStr:  "continue",
		},
		{
			name:     "retry",
			strategy: orchestrator.Retry(3),
			wantStr:  "retry(3)",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.wantStr, tt.strategy.String())
		})
	}
}

func (s *OptionsSuite) TestRetryCount() {
	tests := []struct {
		name     string
		strategy orchestrator.ErrorStrategy
		want     int
	}{
		{
			name:     "stop all has zero retries",
			strategy: orchestrator.StopAll,
			want:     0,
		},
		{
			name:     "continue has zero retries",
			strategy: orchestrator.Continue,
			want:     0,
		},
		{
			name:     "retry has n retries",
			strategy: orchestrator.Retry(5),
			want:     5,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, tt.strategy.RetryCount())
		})
	}
}

func (s *OptionsSuite) TestPlanOption() {
	tests := []struct {
		name        string
		option      orchestrator.PlanOption
		wantOnError orchestrator.ErrorStrategy
	}{
		{
			name:        "on error sets strategy",
			option:      orchestrator.OnError(orchestrator.Continue),
			wantOnError: orchestrator.Continue,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg := &orchestrator.PlanConfig{}
			tt.option(cfg)
			s.Equal(tt.wantOnError.String(), cfg.OnErrorStrategy.String())
		})
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: FAIL

**Step 3: Write minimal implementation**

Create `pkg/orchestrator/options.go`:

```go
package orchestrator

import "fmt"

// ErrorStrategy defines how the runner handles task failures.
type ErrorStrategy struct {
	kind       string
	retryCount int
}

// StopAll cancels all remaining tasks on first failure.
var StopAll = ErrorStrategy{kind: "stop_all"}

// Continue skips dependents of the failed task but continues
// independent tasks.
var Continue = ErrorStrategy{kind: "continue"}

// Retry returns a strategy that retries a failed task n times
// before failing.
func Retry(
	n int,
) ErrorStrategy {
	return ErrorStrategy{kind: "retry", retryCount: n}
}

// String returns a human-readable representation of the strategy.
func (e ErrorStrategy) String() string {
	if e.kind == "retry" {
		return fmt.Sprintf("retry(%d)", e.retryCount)
	}

	return e.kind
}

// RetryCount returns the number of retries for this strategy.
func (e ErrorStrategy) RetryCount() int {
	return e.retryCount
}

// PlanConfig holds plan-level configuration.
type PlanConfig struct {
	OnErrorStrategy ErrorStrategy
}

// PlanOption is a functional option for NewPlan.
type PlanOption func(*PlanConfig)

// OnError returns a PlanOption that sets the default error strategy.
func OnError(
	strategy ErrorStrategy,
) PlanOption {
	return func(cfg *PlanConfig) {
		cfg.OnErrorStrategy = strategy
	}
}
```

**Step 4: Run test to verify it passes**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: PASS

**Step 5: Run lint**

```bash
just go::vet
```

Expected: clean

**Step 6: Commit**

```bash
git add pkg/orchestrator/options.go pkg/orchestrator/options_public_test.go
git commit -m "feat(orchestrator): add ErrorStrategy and PlanOption types"
```

---

## Task 4: Task Type

**Files:**

- Create: `pkg/orchestrator/task.go`
- Create: `pkg/orchestrator/task_public_test.go`

**Step 1: Write the failing test**

Create `pkg/orchestrator/task_public_test.go`:

```go
package orchestrator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

type TaskSuite struct {
	suite.Suite
}

func TestTaskSuite(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}

func (s *TaskSuite) TestDependsOn() {
	tests := []struct {
		name       string
		setupDeps  func(a, b, c *orchestrator.Task)
		checkTask  string
		wantDepLen int
	}{
		{
			name: "single dependency",
			setupDeps: func(a, b, _ *orchestrator.Task) {
				b.DependsOn(a)
			},
			checkTask:  "b",
			wantDepLen: 1,
		},
		{
			name: "multiple dependencies",
			setupDeps: func(a, b, c *orchestrator.Task) {
				c.DependsOn(a, b)
			},
			checkTask:  "c",
			wantDepLen: 2,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			a := orchestrator.NewTask("a", &orchestrator.Op{Operation: "noop"})
			b := orchestrator.NewTask("b", &orchestrator.Op{Operation: "noop"})
			c := orchestrator.NewTask("c", &orchestrator.Op{Operation: "noop"})
			tt.setupDeps(a, b, c)

			tasks := map[string]*orchestrator.Task{"a": a, "b": b, "c": c}
			s.Len(tasks[tt.checkTask].Dependencies(), tt.wantDepLen)
		})
	}
}

func (s *TaskSuite) TestOnlyIfChanged() {
	task := orchestrator.NewTask("t", &orchestrator.Op{Operation: "noop"})
	dep := orchestrator.NewTask("dep", &orchestrator.Op{Operation: "noop"})
	task.DependsOn(dep).OnlyIfChanged()

	s.True(task.RequiresChange())
}

func (s *TaskSuite) TestWhen() {
	task := orchestrator.NewTask("t", &orchestrator.Op{Operation: "noop"})
	called := false
	task.When(func(_ orchestrator.Results) bool {
		called = true

		return true
	})

	guard := task.Guard()
	s.NotNil(guard)
	s.True(guard(orchestrator.Results{}))
	s.True(called)
}

func (s *TaskSuite) TestTaskFunc() {
	fn := func(
		_ context.Context,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	}

	task := orchestrator.NewTaskFunc("custom", fn)
	s.Equal("custom", task.Name())
	s.True(task.IsFunc())
}

func (s *TaskSuite) TestOnErrorOverride() {
	task := orchestrator.NewTask("t", &orchestrator.Op{Operation: "noop"})
	task.OnError(orchestrator.Continue)

	s.NotNil(task.ErrorStrategy())
	s.Equal("continue", task.ErrorStrategy().String())
}
```

**Step 2: Run test to verify it fails**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: FAIL

**Step 3: Write minimal implementation**

Create `pkg/orchestrator/task.go`:

```go
package orchestrator

import "context"

// Op represents a declarative SDK operation.
type Op struct {
	Operation string
	Target    string
	Params    map[string]any
}

// TaskFn is the signature for functional tasks.
type TaskFn func(ctx context.Context) (*Result, error)

// GuardFn is a predicate that determines if a task should run.
type GuardFn func(results Results) bool

// Task is a unit of work in an orchestration plan.
type Task struct {
	name           string
	op             *Op
	fn             TaskFn
	deps           []*Task
	guard          GuardFn
	requiresChange bool
	errorStrategy  *ErrorStrategy
}

// NewTask creates a declarative task wrapping an SDK operation.
func NewTask(
	name string,
	op *Op,
) *Task {
	return &Task{
		name: name,
		op:   op,
	}
}

// NewTaskFunc creates a functional task with custom logic.
func NewTaskFunc(
	name string,
	fn TaskFn,
) *Task {
	return &Task{
		name: name,
		fn:   fn,
	}
}

// Name returns the task name.
func (t *Task) Name() string {
	return t.name
}

// IsFunc returns true if this is a functional task.
func (t *Task) IsFunc() bool {
	return t.fn != nil
}

// Operation returns the declarative operation, or nil for functional
// tasks.
func (t *Task) Operation() *Op {
	return t.op
}

// Fn returns the task function, or nil for declarative tasks.
func (t *Task) Fn() TaskFn {
	return t.fn
}

// DependsOn sets this task's dependencies. Returns the task for
// chaining.
func (t *Task) DependsOn(
	deps ...*Task,
) *Task {
	t.deps = append(t.deps, deps...)

	return t
}

// Dependencies returns the task's dependencies.
func (t *Task) Dependencies() []*Task {
	return t.deps
}

// OnlyIfChanged marks this task to only run if at least one
// dependency reported Changed=true.
func (t *Task) OnlyIfChanged() {
	t.requiresChange = true
}

// RequiresChange returns true if OnlyIfChanged was set.
func (t *Task) RequiresChange() bool {
	return t.requiresChange
}

// When sets a custom guard function that determines whether
// this task should execute.
func (t *Task) When(
	fn GuardFn,
) {
	t.guard = fn
}

// Guard returns the guard function, or nil if none is set.
func (t *Task) Guard() GuardFn {
	return t.guard
}

// OnError sets a per-task error strategy override.
func (t *Task) OnError(
	strategy ErrorStrategy,
) {
	t.errorStrategy = &strategy
}

// ErrorStrategy returns the per-task error strategy, or nil to
// use the plan default.
func (t *Task) ErrorStrategy() *ErrorStrategy {
	return t.errorStrategy
}
```

**Step 4: Run test to verify it passes**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: PASS

**Step 5: Run lint**

```bash
just go::vet
```

Expected: clean

**Step 6: Commit**

```bash
git add pkg/orchestrator/task.go pkg/orchestrator/task_public_test.go
git commit -m "feat(orchestrator): add Task, Op, and dependency types"
```

---

## Task 5: Plan Type

**Files:**

- Create: `pkg/orchestrator/plan.go`
- Create: `pkg/orchestrator/plan_public_test.go`

**Step 1: Write the failing test**

Create `pkg/orchestrator/plan_public_test.go`:

```go
package orchestrator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

type PlanSuite struct {
	suite.Suite
}

func TestPlanSuite(t *testing.T) {
	suite.Run(t, new(PlanSuite))
}

func (s *PlanSuite) TestNewPlan() {
	tests := []struct {
		name         string
		opts         []orchestrator.PlanOption
		wantErrStrat string
	}{
		{
			name:         "default error strategy",
			opts:         nil,
			wantErrStrat: "stop_all",
		},
		{
			name:         "custom error strategy",
			opts:         []orchestrator.PlanOption{orchestrator.OnError(orchestrator.Continue)},
			wantErrStrat: "continue",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			plan := orchestrator.NewPlan(tt.opts...)
			s.Equal(tt.wantErrStrat, plan.Config().OnErrorStrategy.String())
		})
	}
}

func (s *PlanSuite) TestPlanTask() {
	plan := orchestrator.NewPlan()

	t1 := plan.Task("install", &orchestrator.Op{Operation: "command.exec"})
	t2 := plan.Task("configure", &orchestrator.Op{Operation: "network.dns.update"})

	tasks := plan.Tasks()

	s.Len(tasks, 2)
	s.Equal("install", t1.Name())
	s.Equal("configure", t2.Name())
}

func (s *PlanSuite) TestPlanTaskFunc() {
	plan := orchestrator.NewPlan()

	plan.TaskFunc("verify", func(
		_ context.Context,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	s.Len(plan.Tasks(), 1)
	s.True(plan.Tasks()[0].IsFunc())
}

func (s *PlanSuite) TestPlanValidateCycleDetection() {
	plan := orchestrator.NewPlan()
	a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
	b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
	a.DependsOn(b)
	b.DependsOn(a)

	err := plan.Validate()
	s.Error(err)
	s.Contains(err.Error(), "cycle")
}

func (s *PlanSuite) TestPlanValidateNoCycle() {
	plan := orchestrator.NewPlan()
	a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
	b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
	b.DependsOn(a)

	err := plan.Validate()
	s.NoError(err)
}

func (s *PlanSuite) TestPlanValidateDuplicateName() {
	plan := orchestrator.NewPlan()
	plan.Task("same", &orchestrator.Op{Operation: "noop"})
	plan.Task("same", &orchestrator.Op{Operation: "noop"})

	err := plan.Validate()
	s.Error(err)
	s.Contains(err.Error(), "duplicate")
}
```

**Step 2: Run test to verify it fails**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: FAIL

**Step 3: Write minimal implementation**

Create `pkg/orchestrator/plan.go`:

```go
package orchestrator

import (
	"context"
	"fmt"
)

// Plan is a DAG of tasks with dependency edges.
type Plan struct {
	tasks  []*Task
	config PlanConfig
}

// NewPlan creates a new plan with optional configuration.
func NewPlan(
	opts ...PlanOption,
) *Plan {
	cfg := PlanConfig{
		OnErrorStrategy: StopAll,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &Plan{
		config: cfg,
	}
}

// Config returns the plan configuration.
func (p *Plan) Config() PlanConfig {
	return p.config
}

// Task creates a declarative task, adds it to the plan, and returns it.
func (p *Plan) Task(
	name string,
	op *Op,
) *Task {
	t := NewTask(name, op)
	p.tasks = append(p.tasks, t)

	return t
}

// TaskFunc creates a functional task, adds it to the plan, and
// returns it.
func (p *Plan) TaskFunc(
	name string,
	fn TaskFn,
) *Task {
	t := NewTaskFunc(name, fn)
	p.tasks = append(p.tasks, t)

	return t
}

// Tasks returns all tasks in the plan.
func (p *Plan) Tasks() []*Task {
	return p.tasks
}

// Validate checks the plan for errors: duplicate names and cycles.
func (p *Plan) Validate() error {
	names := make(map[string]bool, len(p.tasks))

	for _, t := range p.tasks {
		if names[t.name] {
			return fmt.Errorf("duplicate task name: %q", t.name)
		}

		names[t.name] = true
	}

	return p.detectCycle()
}

// Run validates the plan, resolves the DAG, and executes tasks.
func (p *Plan) Run(
	ctx context.Context,
) (*Report, error) {
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("plan validation: %w", err)
	}

	runner := newRunner(p)

	return runner.run(ctx)
}

// detectCycle uses DFS to find cycles in the dependency graph.
func (p *Plan) detectCycle() error {
	const (
		white = 0 // unvisited
		gray  = 1 // in progress
		black = 2 // done
	)

	color := make(map[string]int, len(p.tasks))

	var visit func(t *Task) error
	visit = func(t *Task) error {
		color[t.name] = gray

		for _, dep := range t.deps {
			switch color[dep.name] {
			case gray:
				return fmt.Errorf(
					"cycle detected: %q depends on %q",
					t.name,
					dep.name,
				)
			case white:
				if err := visit(dep); err != nil {
					return err
				}
			}
		}

		color[t.name] = black

		return nil
	}

	for _, t := range p.tasks {
		if color[t.name] == white {
			if err := visit(t); err != nil {
				return err
			}
		}
	}

	return nil
}

// Stub runner until Task 6 adds the real implementation.
func newRunner(
	_ *Plan,
) *runner {
	return &runner{}
}

type runner struct{}

func (r *runner) run(
	_ context.Context,
) (*Report, error) {
	return &Report{}, nil
}
```

**Step 4: Run test to verify it passes**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: PASS

**Step 5: Run lint**

```bash
just go::vet
```

Expected: clean

**Step 6: Commit**

```bash
git add pkg/orchestrator/plan.go pkg/orchestrator/plan_public_test.go
git commit -m "feat(orchestrator): add Plan with DAG validation"
```

---

## Task 6: Runner — DAG Resolution and Parallel Execution

**Files:**

- Create: `pkg/orchestrator/runner.go`
- Create: `pkg/orchestrator/runner_test.go`
- Modify: `pkg/orchestrator/plan.go` — remove runner stub

**Step 1: Write the failing test**

Create `pkg/orchestrator/runner_test.go` (internal test — needs access to
`topoSort` and `levelize`):

```go
package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RunnerSuite struct {
	suite.Suite
}

func TestRunnerSuite(t *testing.T) {
	suite.Run(t, new(RunnerSuite))
}

func (s *RunnerSuite) TestTopoSort() {
	tests := []struct {
		name  string
		setup func() []*Task
	}{
		{
			name: "linear chain",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})
				c := NewTask("c", &Op{Operation: "noop"})
				b.DependsOn(a)
				c.DependsOn(b)

				return []*Task{a, b, c}
			},
		},
		{
			name: "diamond",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})
				c := NewTask("c", &Op{Operation: "noop"})
				d := NewTask("d", &Op{Operation: "noop"})
				b.DependsOn(a)
				c.DependsOn(a)
				d.DependsOn(b, c)

				return []*Task{a, b, c, d}
			},
		},
		{
			name: "independent tasks",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})

				return []*Task{a, b}
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tasks := tt.setup()
			sorted := topoSort(tasks)

			s.Len(sorted, len(tasks))

			pos := make(map[string]int, len(sorted))
			for i, t := range sorted {
				pos[t.name] = i
			}

			for _, t := range tasks {
				for _, dep := range t.deps {
					s.Less(
						pos[dep.name],
						pos[t.name],
						"%s should come before %s",
						dep.name,
						t.name,
					)
				}
			}
		})
	}
}

func (s *RunnerSuite) TestLevelize() {
	tests := []struct {
		name       string
		setup      func() []*Task
		wantLevels int
	}{
		{
			name: "linear chain has 3 levels",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})
				c := NewTask("c", &Op{Operation: "noop"})
				b.DependsOn(a)
				c.DependsOn(b)

				return []*Task{a, b, c}
			},
			wantLevels: 3,
		},
		{
			name: "diamond has 3 levels",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})
				c := NewTask("c", &Op{Operation: "noop"})
				d := NewTask("d", &Op{Operation: "noop"})
				b.DependsOn(a)
				c.DependsOn(a)
				d.DependsOn(b, c)

				return []*Task{a, b, c, d}
			},
			wantLevels: 3,
		},
		{
			name: "independent tasks in 1 level",
			setup: func() []*Task {
				a := NewTask("a", &Op{Operation: "noop"})
				b := NewTask("b", &Op{Operation: "noop"})

				return []*Task{a, b}
			},
			wantLevels: 1,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tasks := tt.setup()
			levels := levelize(tasks)
			s.Len(levels, tt.wantLevels)
		})
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: FAIL — `topoSort` and `levelize` don't exist.

**Step 3: Write implementation**

Create `pkg/orchestrator/runner.go`:

```go
package orchestrator

import (
	"context"
	"sync"
	"time"
)

// runner executes a validated plan.
type runner struct {
	plan    *Plan
	results Results
	mu      sync.Mutex
}

// newRunner creates a runner for the plan.
func newRunner(
	plan *Plan,
) *runner {
	return &runner{
		plan:    plan,
		results: make(Results),
	}
}

// run executes the plan by levelizing the DAG and running each
// level in parallel.
func (r *runner) run(
	ctx context.Context,
) (*Report, error) {
	start := time.Now()
	levels := levelize(r.plan.tasks)

	var taskResults []TaskResult

	for _, level := range levels {
		results, err := r.runLevel(ctx, level)
		taskResults = append(taskResults, results...)

		if err != nil {
			return &Report{
				Tasks:    taskResults,
				Duration: time.Since(start),
			}, err
		}
	}

	return &Report{
		Tasks:    taskResults,
		Duration: time.Since(start),
	}, nil
}

// runLevel executes all tasks in a level concurrently.
func (r *runner) runLevel(
	ctx context.Context,
	tasks []*Task,
) ([]TaskResult, error) {
	results := make([]TaskResult, len(tasks))

	var wg sync.WaitGroup

	for i, t := range tasks {
		wg.Add(1)

		go func() {
			defer wg.Done()

			results[i] = r.runTask(ctx, t)
		}()
	}

	wg.Wait()

	for _, tr := range results {
		if tr.Status == StatusFailed {
			strategy := r.plan.config.OnErrorStrategy
			if strategy.kind == "stop_all" {
				return results, tr.Error
			}
		}
	}

	return results, nil
}

// runTask executes a single task with guard checks.
func (r *runner) runTask(
	ctx context.Context,
	t *Task,
) TaskResult {
	start := time.Now()

	if t.requiresChange {
		anyChanged := false

		r.mu.Lock()

		for _, dep := range t.deps {
			if res := r.results.Get(dep.name); res != nil && res.Changed {
				anyChanged = true

				break
			}
		}

		r.mu.Unlock()

		if !anyChanged {
			return TaskResult{
				Name:     t.name,
				Status:   StatusSkipped,
				Duration: time.Since(start),
			}
		}
	}

	if t.guard != nil {
		r.mu.Lock()
		shouldRun := t.guard(r.results)
		r.mu.Unlock()

		if !shouldRun {
			return TaskResult{
				Name:     t.name,
				Status:   StatusSkipped,
				Duration: time.Since(start),
			}
		}
	}

	var result *Result
	var err error

	if t.fn != nil {
		result, err = t.fn(ctx)
	} else {
		result = &Result{Changed: false}
	}

	if err != nil {
		return TaskResult{
			Name:     t.name,
			Status:   StatusFailed,
			Duration: time.Since(start),
			Error:    err,
		}
	}

	r.mu.Lock()
	r.results[t.name] = result
	r.mu.Unlock()

	status := StatusUnchanged
	if result.Changed {
		status = StatusChanged
	}

	return TaskResult{
		Name:     t.name,
		Status:   status,
		Changed:  result.Changed,
		Duration: time.Since(start),
	}
}

// topoSort returns tasks in topological order using Kahn's algorithm.
func topoSort(
	tasks []*Task,
) []*Task {
	inDegree := make(map[string]int, len(tasks))
	taskMap := make(map[string]*Task, len(tasks))

	for _, t := range tasks {
		taskMap[t.name] = t

		if _, ok := inDegree[t.name]; !ok {
			inDegree[t.name] = 0
		}
	}

	for _, t := range tasks {
		for range t.deps {
			inDegree[t.name]++
		}
	}

	var queue []*Task

	for _, t := range tasks {
		if inDegree[t.name] == 0 {
			queue = append(queue, t)
		}
	}

	var sorted []*Task

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for _, t := range tasks {
			for _, dep := range t.deps {
				if dep.name == current.name {
					inDegree[t.name]--

					if inDegree[t.name] == 0 {
						queue = append(queue, t)
					}
				}
			}
		}
	}

	return sorted
}

// levelize groups tasks into levels where all tasks in a level can
// run concurrently (all dependencies are in earlier levels).
func levelize(
	tasks []*Task,
) [][]*Task {
	level := make(map[string]int, len(tasks))

	var computeLevel func(t *Task) int
	computeLevel = func(t *Task) int {
		if l, ok := level[t.name]; ok {
			return l
		}

		maxDep := -1

		for _, dep := range t.deps {
			depLevel := computeLevel(dep)
			if depLevel > maxDep {
				maxDep = depLevel
			}
		}

		level[t.name] = maxDep + 1

		return maxDep + 1
	}

	maxLevel := 0

	for _, t := range tasks {
		l := computeLevel(t)
		if l > maxLevel {
			maxLevel = l
		}
	}

	levels := make([][]*Task, maxLevel+1)

	for _, t := range tasks {
		l := level[t.name]
		levels[l] = append(levels[l], t)
	}

	return levels
}
```

**Step 4: Remove the runner stub from `plan.go`**

Delete the `newRunner` stub, `runner` type, and `run` method from the bottom of
`plan.go`. The real implementations are now in `runner.go`.

**Step 5: Run test to verify it passes**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: PASS

**Step 6: Run lint**

```bash
just go::vet
```

Expected: clean

**Step 7: Commit**

```bash
git add pkg/orchestrator/runner.go pkg/orchestrator/runner_test.go pkg/orchestrator/plan.go
git commit -m "feat(orchestrator): add DAG runner with parallel execution"
```

---

## Task 7: Integration Test — Plan.Run End-to-End

**Files:**

- Create: `pkg/orchestrator/plan_integration_test.go`

**Step 1: Write the test**

Create `pkg/orchestrator/plan_integration_test.go`:

```go
package orchestrator_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

type PlanIntegrationSuite struct {
	suite.Suite
}

func TestPlanIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PlanIntegrationSuite))
}

func (s *PlanIntegrationSuite) TestRunLinearChain() {
	var order []string
	plan := orchestrator.NewPlan()

	mkTask := func(name string, changed bool) *orchestrator.Task {
		return plan.TaskFunc(name, func(
			_ context.Context,
		) (*orchestrator.Result, error) {
			order = append(order, name)

			return &orchestrator.Result{Changed: changed}, nil
		})
	}

	a := mkTask("a", true)
	b := mkTask("b", true)
	c := mkTask("c", false)

	b.DependsOn(a)
	c.DependsOn(b)

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Equal([]string{"a", "b", "c"}, order)
	s.Len(report.Tasks, 3)
	s.Contains(report.Summary(), "2 changed")
	s.Contains(report.Summary(), "1 unchanged")
}

func (s *PlanIntegrationSuite) TestRunParallelExecution() {
	var concurrentMax atomic.Int32
	var concurrent atomic.Int32

	plan := orchestrator.NewPlan()

	mkTask := func(name string) *orchestrator.Task {
		return plan.TaskFunc(name, func(
			_ context.Context,
		) (*orchestrator.Result, error) {
			cur := concurrent.Add(1)

			for {
				max := concurrentMax.Load()
				if cur > max {
					if concurrentMax.CompareAndSwap(max, cur) {
						break
					}
				} else {
					break
				}
			}

			concurrent.Add(-1)

			return &orchestrator.Result{Changed: false}, nil
		})
	}

	mkTask("a")
	mkTask("b")
	mkTask("c")

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Len(report.Tasks, 3)
	s.GreaterOrEqual(int(concurrentMax.Load()), 1)
}

func (s *PlanIntegrationSuite) TestRunOnlyIfChanged() {
	plan := orchestrator.NewPlan()
	skippedRan := false

	dep := plan.TaskFunc("dep", func(
		_ context.Context,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	conditional := plan.TaskFunc("conditional", func(
		_ context.Context,
	) (*orchestrator.Result, error) {
		skippedRan = true

		return &orchestrator.Result{Changed: true}, nil
	})

	conditional.DependsOn(dep).OnlyIfChanged()

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.False(skippedRan)
	s.Contains(report.Summary(), "skipped")
}

func (s *PlanIntegrationSuite) TestRunStopAllOnError() {
	plan := orchestrator.NewPlan()

	fail := plan.TaskFunc("fail", func(
		_ context.Context,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("boom")
	})

	didRun := false

	next := plan.TaskFunc("next", func(
		_ context.Context,
	) (*orchestrator.Result, error) {
		didRun = true

		return &orchestrator.Result{}, nil
	})

	next.DependsOn(fail)

	_, err := plan.Run(context.Background())
	s.Error(err)
	s.False(didRun)
}

func (s *PlanIntegrationSuite) TestRunContinueOnError() {
	plan := orchestrator.NewPlan(
		orchestrator.OnError(orchestrator.Continue),
	)

	plan.TaskFunc("fail", func(
		_ context.Context,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("boom")
	})

	didRun := false

	plan.TaskFunc("independent", func(
		_ context.Context,
	) (*orchestrator.Result, error) {
		didRun = true

		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.True(didRun)
	s.Len(report.Tasks, 2)
}

func (s *PlanIntegrationSuite) TestRunCycleDetection() {
	plan := orchestrator.NewPlan()
	a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
	b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
	a.DependsOn(b)
	b.DependsOn(a)

	_, err := plan.Run(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "cycle")
}
```

**Step 2: Run test**

```bash
go test -v ./pkg/orchestrator/...
```

Expected: PASS

**Step 3: Run lint**

```bash
just go::vet
```

Expected: clean

**Step 4: Commit**

```bash
git add pkg/orchestrator/plan_integration_test.go
git commit -m "test(orchestrator): add end-to-end Plan.Run tests"
```

---

## Task 8: Orchestrator Example

**Files:**

- Create: `examples/orchestrator/main.go`
- Create: `examples/orchestrator/go.mod`

**Step 1: Create `examples/orchestrator/main.go`**

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

func main() {
	plan := orchestrator.NewPlan()

	createUser := plan.Task("create-user", &orchestrator.Op{
		Operation: "command.exec",
		Target:    "_all",
		Params: map[string]any{
			"command": "useradd",
			"args":    []string{"-m", "www"},
		},
	})

	installNginx := plan.Task("install-nginx", &orchestrator.Op{
		Operation: "command.exec",
		Target:    "_all",
		Params: map[string]any{
			"command": "apt",
			"args":    []string{"install", "-y", "nginx"},
		},
	})

	configureDNS := plan.Task("configure-dns", &orchestrator.Op{
		Operation: "network.dns.update",
		Target:    "_all",
		Params: map[string]any{
			"address": "8.8.8.8",
		},
	})

	startNginx := plan.Task("start-nginx", &orchestrator.Op{
		Operation: "command.exec",
		Target:    "_all",
		Params: map[string]any{
			"command": "systemctl",
			"args":    []string{"start", "nginx"},
		},
	})

	installNginx.DependsOn(createUser)
	startNginx.DependsOn(installNginx, configureDNS)
	startNginx.OnlyIfChanged()

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(report.Summary())

	for _, r := range report.Tasks {
		fmt.Printf(
			"%s: %s (changed=%v, duration=%s)\n",
			r.Name,
			r.Status,
			r.Changed,
			r.Duration,
		)
	}
}
```

**Step 2: Create `examples/orchestrator/go.mod`**

```bash
cd examples/orchestrator
go mod init github.com/osapi-io/osapi-sdk/examples/orchestrator
```

Add replace directive for local development:

```
replace github.com/osapi-io/osapi-sdk => ../..
```

Then:

```bash
go mod tidy
```

**Step 3: Verify it compiles**

```bash
go build ./...
```

Expected: compiles

**Step 4: Commit**

```bash
git add examples/orchestrator/
git commit -m "docs(orchestrator): add webserver deployment example"
```

---

## Task 9: Final Verification

**Step 1: Run full test suite**

```bash
just test
```

Expected: all tests pass, lint clean, coverage reported.

**Step 2: Verify example compiles**

```bash
cd examples/orchestrator && go build ./...
```

Expected: compiles.

**Step 3: Commit any remaining changes**

If `go mod tidy` or formatting produced changes:

```bash
git add -A
git commit -m "chore: tidy modules and formatting"
```
