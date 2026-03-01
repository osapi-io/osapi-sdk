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
