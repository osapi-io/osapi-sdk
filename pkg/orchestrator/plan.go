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
