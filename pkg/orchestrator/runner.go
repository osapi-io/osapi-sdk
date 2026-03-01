package orchestrator

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

	if r.plan.config.Verbose {
		r.log("%s", r.plan.Explain())
	}

	var taskResults []TaskResult

	for i, level := range levels {
		if r.plan.config.Verbose {
			names := make([]string, len(level))
			for j, t := range level {
				names[j] = t.name
			}

			if len(level) > 1 {
				r.log(
					"--- Level %d: %s (parallel)\n",
					i,
					strings.Join(names, ", "),
				)
			} else {
				r.log("--- Level %d: %s\n", i, names[0])
			}
		}

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

// log writes verbose output if configured.
func (r *runner) log(
	format string,
	args ...any,
) {
	if r.plan.config.Output != nil {
		_, _ = fmt.Fprintf(r.plan.config.Output, format, args...)
	}
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

	verbose := r.plan.config.Verbose

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
			if verbose {
				r.log("    %-20s skipped (no deps changed)\n", t.name)
			}

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
			if verbose {
				r.log("    %-20s skipped (guard)\n", t.name)
			}

			return TaskResult{
				Name:     t.name,
				Status:   StatusSkipped,
				Duration: time.Since(start),
			}
		}
	}

	if verbose {
		r.log("    %-20s running...\n", t.name)
	}

	var result *Result
	var err error

	client := r.plan.client

	if t.fn != nil {
		result, err = t.fn(ctx, client)
	} else {
		result, err = r.executeOp(ctx, t.op)
	}

	elapsed := time.Since(start)

	if err != nil {
		if verbose {
			r.log("    %-20s FAILED (%s)\n", t.name, elapsed)
		}

		return TaskResult{
			Name:     t.name,
			Status:   StatusFailed,
			Duration: elapsed,
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

	if verbose {
		r.log("    %-20s %s (%s)\n", t.name, status, elapsed)
	}

	return TaskResult{
		Name:     t.name,
		Status:   status,
		Changed:  result.Changed,
		Duration: elapsed,
	}
}

// defaultPollInterval is the default interval between job status polls.
var defaultPollInterval = 500 * time.Millisecond

// executeOp submits a declarative Op as a job via the SDK and polls
// for completion.
func (r *runner) executeOp(
	ctx context.Context,
	op *Op,
) (*Result, error) {
	client := r.plan.client
	if client == nil {
		return nil, fmt.Errorf(
			"op task %q requires an OSAPI client",
			op.Operation,
		)
	}

	operation := make(map[string]interface{}, len(op.Params)+1)
	operation["operation"] = op.Operation

	for k, v := range op.Params {
		operation[k] = v
	}

	createResp, err := client.Job.Create(ctx, operation, op.Target)
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	if createResp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf(
			"create job: unexpected status %d",
			createResp.StatusCode(),
		)
	}

	jobID := createResp.JSON201.JobId.String()

	return r.pollJob(ctx, jobID)
}

// pollJob polls a job until it reaches a terminal state.
func (r *runner) pollJob(
	ctx context.Context,
	jobID string,
) (*Result, error) {
	ticker := time.NewTicker(defaultPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			resp, err := r.plan.client.Job.Get(ctx, jobID)
			if err != nil {
				return nil, fmt.Errorf("poll job %s: %w", jobID, err)
			}

			if resp.StatusCode() != http.StatusOK {
				return nil, fmt.Errorf(
					"poll job %s: unexpected status %d",
					jobID,
					resp.StatusCode(),
				)
			}

			status := ""
			if resp.JSON200.Status != nil {
				status = *resp.JSON200.Status
			}

			switch status {
			case "completed":
				data := make(map[string]any)
				if resp.JSON200.Result != nil {
					if m, ok := resp.JSON200.Result.(map[string]any); ok {
						data = m
					}
				}

				return &Result{Changed: true, Data: data}, nil
			case "failed":
				errMsg := "job failed"
				if resp.JSON200.Error != nil {
					errMsg = *resp.JSON200.Error
				}

				return nil, fmt.Errorf("job %s: %s", jobID, errMsg)
			}
		}
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
