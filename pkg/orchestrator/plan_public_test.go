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
