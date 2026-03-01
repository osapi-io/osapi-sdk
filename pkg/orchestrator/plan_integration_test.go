package orchestrator_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
	"github.com/osapi-io/osapi-sdk/pkg/osapi"
)

type PlanIntegrationSuite struct {
	suite.Suite
}

func TestPlanIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PlanIntegrationSuite))
}

func (s *PlanIntegrationSuite) TestRunLinearChain() {
	var order []string
	plan := orchestrator.NewPlan(nil)

	mkTask := func(name string, changed bool) *orchestrator.Task {
		return plan.TaskFunc(name, func(
			_ context.Context,
			_ *osapi.Client,
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

	plan := orchestrator.NewPlan(nil)

	mkTask := func(name string) *orchestrator.Task {
		return plan.TaskFunc(name, func(
			_ context.Context,
			_ *osapi.Client,
		) (*orchestrator.Result, error) {
			cur := concurrent.Add(1)

			for {
				prev := concurrentMax.Load()
				if cur > prev {
					if concurrentMax.CompareAndSwap(prev, cur) {
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
	plan := orchestrator.NewPlan(nil)
	skippedRan := false

	dep := plan.TaskFunc("dep", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	conditional := plan.TaskFunc("conditional", func(
		_ context.Context,
		_ *osapi.Client,
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
	plan := orchestrator.NewPlan(nil)

	fail := plan.TaskFunc("fail", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("boom")
	})

	didRun := false

	next := plan.TaskFunc("next", func(
		_ context.Context,
		_ *osapi.Client,
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
		nil,
		orchestrator.OnError(orchestrator.Continue),
	)

	plan.TaskFunc("fail", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("boom")
	})

	didRun := false

	plan.TaskFunc("independent", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		didRun = true

		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.True(didRun)
	s.Len(report.Tasks, 2)
}

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

	// c is independent of a — should still run
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
	s.Equal(orchestrator.StatusSkipped, results["b"]) // dependent of failed task
	s.Equal(orchestrator.StatusChanged, results["c"]) // independent, should run
}

func (s *PlanIntegrationSuite) TestContinueStrategyTransitive() {
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

	b := plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})
	b.DependsOn(a)

	c := plan.TaskFunc("c", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})
	c.DependsOn(b)

	report, err := plan.Run(context.Background())
	s.NoError(err)

	results := make(map[string]orchestrator.Status)
	for _, r := range report.Tasks {
		results[r.Name] = r.Status
	}

	s.Equal(orchestrator.StatusFailed, results["a"])
	s.Equal(orchestrator.StatusSkipped, results["b"])
	s.Equal(orchestrator.StatusSkipped, results["c"]) // transitive skip
}

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
	s.Error(err)
	s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestRunCycleDetection() {
	plan := orchestrator.NewPlan(nil)
	a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
	b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
	a.DependsOn(b)
	b.DependsOn(a)

	_, err := plan.Run(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "cycle")
}

func (s *PlanIntegrationSuite) TestRunOpTaskRequiresClient() {
	plan := orchestrator.NewPlan(nil)
	plan.Task("install", &orchestrator.Op{
		Operation: "command.exec",
		Target:    "_any",
		Params:    map[string]any{"command": "uptime"},
	})

	report, err := plan.Run(context.Background())
	s.Error(err)
	s.NotNil(report)
	s.Contains(report.Tasks[0].Error.Error(), "requires an OSAPI client")
}

func (s *PlanIntegrationSuite) TestHooksCalledDuringRun() {
	var events []string
	hooks := orchestrator.Hooks{
		BeforePlan: func(_ string) {
			events = append(events, "before-plan")
		},
		AfterPlan: func(_ *orchestrator.Report) {
			events = append(events, "after-plan")
		},
		BeforeLevel: func(level int, _ []*orchestrator.Task, _ bool) {
			events = append(events, fmt.Sprintf("before-level-%d", level))
		},
		AfterLevel: func(level int, _ []orchestrator.TaskResult) {
			events = append(events, fmt.Sprintf("after-level-%d", level))
		},
		BeforeTask: func(task *orchestrator.Task) {
			events = append(events, "before-"+task.Name())
		},
		AfterTask: func(_ *orchestrator.Task, r orchestrator.TaskResult) {
			events = append(events, "after-"+r.Name)
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
		"after-level-0",
		"before-level-1",
		"before-b",
		"after-b",
		"after-level-1",
		"after-plan",
	}, events)
}
