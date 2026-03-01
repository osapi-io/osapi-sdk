package orchestrator

import (
	"fmt"
	"io"
	"os"
)

// ErrorStrategy defines how the runner handles task failures.
type ErrorStrategy struct {
	kind       string
	retryCount int
}

// StopAll cancels all remaining tasks on first failure.
var StopAll = ErrorStrategy{kind: "stop_all"}

// Continue skips dependents of the failed task but continues
// independent tasks.
var Continue = ErrorStrategy{kind: "continue"}

// Retry returns a strategy that retries a failed task n times
// before failing.
func Retry(
	n int,
) ErrorStrategy {
	return ErrorStrategy{kind: "retry", retryCount: n}
}

// String returns a human-readable representation of the strategy.
func (e ErrorStrategy) String() string {
	if e.kind == "retry" {
		return fmt.Sprintf("retry(%d)", e.retryCount)
	}

	return e.kind
}

// RetryCount returns the number of retries for this strategy.
func (e ErrorStrategy) RetryCount() int {
	return e.retryCount
}

// Hooks provides consumer-controlled callbacks for plan execution
// events. All fields are optional — nil callbacks are skipped.
type Hooks struct {
	BeforePlan  func(explain string)
	BeforeLevel func(level int, names []string, parallel bool)
	BeforeTask  func(name string)
	AfterTask   func(result TaskResult)
	AfterPlan   func(report *Report)
}

// PlanConfig holds plan-level configuration.
type PlanConfig struct {
	OnErrorStrategy ErrorStrategy
	Verbose         bool
	Output          io.Writer
	Hooks           *Hooks
}

// PlanOption is a functional option for NewPlan.
type PlanOption func(*PlanConfig)

// OnError returns a PlanOption that sets the default error strategy.
func OnError(
	strategy ErrorStrategy,
) PlanOption {
	return func(cfg *PlanConfig) {
		cfg.OnErrorStrategy = strategy
	}
}

// WithVerbose enables execution logging. Output goes to stdout
// unless overridden with WithOutput.
func WithVerbose() PlanOption {
	return func(cfg *PlanConfig) {
		cfg.Verbose = true

		if cfg.Output == nil {
			cfg.Output = os.Stdout
		}
	}
}

// WithOutput sets the writer for verbose output.
func WithOutput(
	w io.Writer,
) PlanOption {
	return func(cfg *PlanConfig) {
		cfg.Output = w
	}
}

// WithHooks attaches lifecycle callbacks to plan execution.
func WithHooks(
	hooks Hooks,
) PlanOption {
	return func(cfg *PlanConfig) {
		cfg.Hooks = &hooks
	}
}
