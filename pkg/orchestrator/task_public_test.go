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
