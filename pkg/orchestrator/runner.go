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
