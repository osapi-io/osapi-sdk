package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RunnerBroadcastTestSuite struct {
	suite.Suite
}

func TestRunnerBroadcastTestSuite(t *testing.T) {
	suite.Run(t, new(RunnerBroadcastTestSuite))
}

func (s *RunnerBroadcastTestSuite) TestIsBroadcastTarget() {
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{
			name:   "all agents is broadcast",
			target: "_all",
			want:   true,
		},
		{
			name:   "label selector is broadcast",
			target: "role:web",
			want:   true,
		},
		{
			name:   "single agent is not broadcast",
			target: "agent-001",
			want:   false,
		},
		{
			name:   "empty string is not broadcast",
			target: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := IsBroadcastTarget(tt.target)
			s.Equal(tt.want, got)
		})
	}
}

func (s *RunnerBroadcastTestSuite) TestExtractHostResults() {
	tests := []struct {
		name string
		data map[string]any
		want []HostResult
	}{
		{
			name: "extracts host results from results array",
			data: map[string]any{
				"results": []any{
					map[string]any{
						"hostname": "host-1",
						"changed":  true,
						"data":     "something",
					},
					map[string]any{
						"hostname": "host-2",
						"changed":  false,
						"error":    "connection refused",
					},
				},
			},
			want: []HostResult{
				{
					Hostname: "host-1",
					Changed:  true,
					Data: map[string]any{
						"hostname": "host-1",
						"changed":  true,
						"data":     "something",
					},
				},
				{
					Hostname: "host-2",
					Changed:  false,
					Error:    "connection refused",
					Data: map[string]any{
						"hostname": "host-2",
						"changed":  false,
						"error":    "connection refused",
					},
				},
			},
		},
		{
			name: "no results key returns nil",
			data: map[string]any{
				"other": "value",
			},
			want: nil,
		},
		{
			name: "results not an array returns nil",
			data: map[string]any{
				"results": "not-an-array",
			},
			want: nil,
		},
		{
			name: "empty results array returns empty slice",
			data: map[string]any{
				"results": []any{},
			},
			want: []HostResult{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := extractHostResults(tt.data)
			s.Equal(tt.want, got)
		})
	}
}

func (s *RunnerBroadcastTestSuite) TestIsCommandOp() {
	tests := []struct {
		name      string
		operation string
		want      bool
	}{
		{
			name:      "command.exec.execute is a command op",
			operation: "command.exec.execute",
			want:      true,
		},
		{
			name:      "command.shell.execute is a command op",
			operation: "command.shell.execute",
			want:      true,
		},
		{
			name:      "node.hostname.get is not a command op",
			operation: "node.hostname.get",
			want:      false,
		},
		{
			name:      "empty string is not a command op",
			operation: "",
			want:      false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := isCommandOp(tt.operation)
			s.Equal(tt.want, got)
		})
	}
}
