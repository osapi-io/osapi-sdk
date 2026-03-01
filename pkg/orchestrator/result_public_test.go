package orchestrator_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

type ResultPublicTestSuite struct {
	suite.Suite
}

func TestResultPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ResultPublicTestSuite))
}

func (s *ResultPublicTestSuite) TestReportSummary() {
	tests := []struct {
		name     string
		tasks    []orchestrator.TaskResult
		contains []string
	}{
		{
			name: "mixed results",
			tasks: []orchestrator.TaskResult{
				{
					Name:     "a",
					Status:   orchestrator.StatusChanged,
					Changed:  true,
					Duration: time.Second,
				},
				{
					Name:     "b",
					Status:   orchestrator.StatusUnchanged,
					Changed:  false,
					Duration: 2 * time.Second,
				},
				{Name: "c", Status: orchestrator.StatusSkipped, Changed: false, Duration: 0},
				{
					Name:     "d",
					Status:   orchestrator.StatusChanged,
					Changed:  true,
					Duration: 500 * time.Millisecond,
				},
			},
			contains: []string{"4 tasks", "2 changed", "1 unchanged", "1 skipped"},
		},
		{
			name: "all statuses including failed",
			tasks: []orchestrator.TaskResult{
				{Name: "a", Status: orchestrator.StatusChanged, Changed: true},
				{Name: "b", Status: orchestrator.StatusUnchanged},
				{Name: "c", Status: orchestrator.StatusSkipped},
				{Name: "d", Status: orchestrator.StatusFailed},
			},
			contains: []string{"4 tasks", "1 changed", "1 unchanged", "1 skipped", "1 failed"},
		},
		{
			name:     "empty report",
			tasks:    nil,
			contains: []string{"0 tasks"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			report := orchestrator.Report{Tasks: tt.tasks}
			summary := report.Summary()
			for _, c := range tt.contains {
				s.Contains(summary, c)
			}
		})
	}
}

func (s *ResultPublicTestSuite) TestResultsGet() {
	tests := []struct {
		name       string
		results    orchestrator.Results
		lookupName string
		wantNil    bool
		wantChange bool
	}{
		{
			name: "found",
			results: orchestrator.Results{
				"install": {Changed: true},
			},
			lookupName: "install",
			wantNil:    false,
			wantChange: true,
		},
		{
			name:       "not found",
			results:    orchestrator.Results{},
			lookupName: "missing",
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := tt.results.Get(tt.lookupName)
			if tt.wantNil {
				s.Nil(got)
			} else {
				s.Require().NotNil(got)
				s.Equal(tt.wantChange, got.Changed)
			}
		})
	}
}
