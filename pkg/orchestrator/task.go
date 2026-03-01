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
