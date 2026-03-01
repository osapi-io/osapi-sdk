package orchestrator

import "fmt"

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

// PlanConfig holds plan-level configuration.
type PlanConfig struct {
	OnErrorStrategy ErrorStrategy
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
