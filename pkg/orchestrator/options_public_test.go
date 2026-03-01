package orchestrator_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

type OptionsSuite struct {
	suite.Suite
}

func TestOptionsSuite(t *testing.T) {
	suite.Run(t, new(OptionsSuite))
}

func (s *OptionsSuite) TestErrorStrategy() {
	tests := []struct {
		name     string
		strategy orchestrator.ErrorStrategy
		wantStr  string
	}{
		{
			name:     "stop all",
			strategy: orchestrator.StopAll,
			wantStr:  "stop_all",
		},
		{
			name:     "continue",
			strategy: orchestrator.Continue,
			wantStr:  "continue",
		},
		{
			name:     "retry",
			strategy: orchestrator.Retry(3),
			wantStr:  "retry(3)",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.wantStr, tt.strategy.String())
		})
	}
}

func (s *OptionsSuite) TestRetryCount() {
	tests := []struct {
		name     string
		strategy orchestrator.ErrorStrategy
		want     int
	}{
		{
			name:     "stop all has zero retries",
			strategy: orchestrator.StopAll,
			want:     0,
		},
		{
			name:     "continue has zero retries",
			strategy: orchestrator.Continue,
			want:     0,
		},
		{
			name:     "retry has n retries",
			strategy: orchestrator.Retry(5),
			want:     5,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, tt.strategy.RetryCount())
		})
	}
}

func (s *OptionsSuite) TestWithVerbose() {
	plan := orchestrator.NewPlan(nil, orchestrator.WithVerbose())
	s.True(plan.Config().Verbose)
	s.NotNil(plan.Config().Output)
}

func (s *OptionsSuite) TestWithOutput() {
	var buf bytes.Buffer
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.WithOutput(&buf),
		orchestrator.WithVerbose(),
	)
	s.Equal(&buf, plan.Config().Output)
}

func (s *OptionsSuite) TestWithHooks() {
	called := false
	hooks := orchestrator.Hooks{
		BeforeTask: func(_ *orchestrator.Task) {
			called = true
		},
	}

	cfg := orchestrator.PlanConfig{}
	opt := orchestrator.WithHooks(hooks)
	opt(&cfg)

	s.NotNil(cfg.Hooks)
	s.NotNil(cfg.Hooks.BeforeTask)

	// Create a task to pass to the callback.
	t := orchestrator.NewTask(
		"test",
		&orchestrator.Op{Operation: "node.hostname.get", Target: "_any"},
	)
	cfg.Hooks.BeforeTask(t)
	s.True(called)
}

func (s *OptionsSuite) TestHooksDefaults() {
	h := orchestrator.Hooks{}

	// Nil callbacks should be safe — no panic.
	s.Nil(h.BeforePlan)
	s.Nil(h.AfterPlan)
	s.Nil(h.BeforeLevel)
	s.Nil(h.AfterLevel)
	s.Nil(h.BeforeTask)
	s.Nil(h.AfterTask)
	s.Nil(h.OnRetry)
	s.Nil(h.OnSkip)
}

func (s *OptionsSuite) TestPlanOption() {
	tests := []struct {
		name        string
		option      orchestrator.PlanOption
		wantOnError orchestrator.ErrorStrategy
	}{
		{
			name:        "on error sets strategy",
			option:      orchestrator.OnError(orchestrator.Continue),
			wantOnError: orchestrator.Continue,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg := &orchestrator.PlanConfig{}
			tt.option(cfg)
			s.Equal(tt.wantOnError.String(), cfg.OnErrorStrategy.String())
		})
	}
}
