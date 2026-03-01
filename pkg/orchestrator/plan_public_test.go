package orchestrator_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
	"github.com/osapi-io/osapi-sdk/pkg/osapi"
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
			plan := orchestrator.NewPlan(nil, tt.opts...)
			s.Equal(tt.wantErrStrat, plan.Config().OnErrorStrategy.String())
		})
	}
}

func (s *PlanSuite) TestNewPlanClient() {
	plan := orchestrator.NewPlan(nil)
	s.Nil(plan.Client())
}

func (s *PlanSuite) TestPlanTask() {
	plan := orchestrator.NewPlan(nil)

	t1 := plan.Task("install", &orchestrator.Op{Operation: "command.exec"})
	t2 := plan.Task("configure", &orchestrator.Op{Operation: "network.dns.update"})

	tasks := plan.Tasks()

	s.Len(tasks, 2)
	s.Equal("install", t1.Name())
	s.Equal("configure", t2.Name())
}

func (s *PlanSuite) TestPlanTaskFunc() {
	plan := orchestrator.NewPlan(nil)

	plan.TaskFunc("verify", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	s.Len(plan.Tasks(), 1)
	s.True(plan.Tasks()[0].IsFunc())
}

func (s *PlanSuite) TestPlanValidateCycleDetection() {
	plan := orchestrator.NewPlan(nil)
	a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
	b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
	a.DependsOn(b)
	b.DependsOn(a)

	err := plan.Validate()
	s.Error(err)
	s.Contains(err.Error(), "cycle")
}

func (s *PlanSuite) TestPlanValidateNoCycle() {
	plan := orchestrator.NewPlan(nil)
	a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
	b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
	b.DependsOn(a)

	err := plan.Validate()
	s.NoError(err)
}

func (s *PlanSuite) TestPlanExplain() {
	plan := orchestrator.NewPlan(nil)
	a := plan.Task("install", &orchestrator.Op{Operation: "command.exec"})
	b := plan.TaskFunc("verify", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, nil
	})
	b.DependsOn(a)
	b.When(func(_ orchestrator.Results) bool { return true })

	out := plan.Explain()
	s.Contains(out, "2 tasks, 2 levels")
	s.Contains(out, "install [op]")
	s.Contains(out, "verify [fn] <- install (when)")
}

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

func (s *PlanSuite) TestPlanValidateDuplicateName() {
	plan := orchestrator.NewPlan(nil)
	plan.Task("same", &orchestrator.Op{Operation: "noop"})
	plan.Task("same", &orchestrator.Op{Operation: "noop"})

	err := plan.Validate()
	s.Error(err)
	s.Contains(err.Error(), "duplicate")
}
