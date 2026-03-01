package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RunnerTestSuite struct {
	suite.Suite
}

func TestRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerTestSuite))
}

func (s *RunnerTestSuite) TestLevelize() {
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
