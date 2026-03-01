package orchestrator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
	"github.com/osapi-io/osapi-sdk/pkg/osapi"
)

type PlanIntegrationSuite struct {
	suite.Suite
}

func TestPlanIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PlanIntegrationSuite))
}

func (s *PlanIntegrationSuite) TestRunLinearChain() {
	var order []string
	plan := orchestrator.NewPlan(nil)

	mkTask := func(name string, changed bool) *orchestrator.Task {
		return plan.TaskFunc(name, func(
			_ context.Context,
			_ *osapi.Client,
		) (*orchestrator.Result, error) {
			order = append(order, name)

			return &orchestrator.Result{Changed: changed}, nil
		})
	}

	a := mkTask("a", true)
	b := mkTask("b", true)
	c := mkTask("c", false)

	b.DependsOn(a)
	c.DependsOn(b)

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Equal([]string{"a", "b", "c"}, order)
	s.Len(report.Tasks, 3)
	s.Contains(report.Summary(), "2 changed")
	s.Contains(report.Summary(), "1 unchanged")
}

func (s *PlanIntegrationSuite) TestRunParallelExecution() {
	var concurrentMax atomic.Int32
	var concurrent atomic.Int32

	plan := orchestrator.NewPlan(nil)

	mkTask := func(name string) *orchestrator.Task {
		return plan.TaskFunc(name, func(
			_ context.Context,
			_ *osapi.Client,
		) (*orchestrator.Result, error) {
			cur := concurrent.Add(1)

			for {
				prev := concurrentMax.Load()
				if cur > prev {
					if concurrentMax.CompareAndSwap(prev, cur) {
						break
					}
				} else {
					break
				}
			}

			concurrent.Add(-1)

			return &orchestrator.Result{Changed: false}, nil
		})
	}

	mkTask("a")
	mkTask("b")
	mkTask("c")

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Len(report.Tasks, 3)
	s.GreaterOrEqual(int(concurrentMax.Load()), 1)
}

func (s *PlanIntegrationSuite) TestRunOnlyIfChanged() {
	plan := orchestrator.NewPlan(nil)
	skippedRan := false

	dep := plan.TaskFunc("dep", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	conditional := plan.TaskFunc("conditional", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		skippedRan = true

		return &orchestrator.Result{Changed: true}, nil
	})

	conditional.DependsOn(dep).OnlyIfChanged()

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.False(skippedRan)
	s.Contains(report.Summary(), "skipped")
}

func (s *PlanIntegrationSuite) TestRunStopAllOnError() {
	plan := orchestrator.NewPlan(nil)

	fail := plan.TaskFunc("fail", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("boom")
	})

	didRun := false

	next := plan.TaskFunc("next", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		didRun = true

		return &orchestrator.Result{}, nil
	})

	next.DependsOn(fail)

	_, err := plan.Run(context.Background())
	s.Error(err)
	s.False(didRun)
}

func (s *PlanIntegrationSuite) TestRunContinueOnError() {
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Continue),
	)

	plan.TaskFunc("fail", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("boom")
	})

	didRun := false

	plan.TaskFunc("independent", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		didRun = true

		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.True(didRun)
	s.Len(report.Tasks, 2)
}

func (s *PlanIntegrationSuite) TestContinueStrategy() {
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Continue),
	)

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("a failed")
	})

	plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	}).DependsOn(a)

	// c is independent of a — should still run
	plan.TaskFunc("c", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err) // Continue doesn't return error

	s.Len(report.Tasks, 3)

	results := make(map[string]orchestrator.Status)
	for _, r := range report.Tasks {
		results[r.Name] = r.Status
	}

	s.Equal(orchestrator.StatusFailed, results["a"])
	s.Equal(orchestrator.StatusSkipped, results["b"]) // dependent of failed task
	s.Equal(orchestrator.StatusChanged, results["c"]) // independent, should run
}

func (s *PlanIntegrationSuite) TestContinueStrategyTransitive() {
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Continue),
	)

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("a failed")
	})

	b := plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})
	b.DependsOn(a)

	c := plan.TaskFunc("c", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})
	c.DependsOn(b)

	report, err := plan.Run(context.Background())
	s.NoError(err)

	results := make(map[string]orchestrator.Status)
	for _, r := range report.Tasks {
		results[r.Name] = r.Status
	}

	s.Equal(orchestrator.StatusFailed, results["a"])
	s.Equal(orchestrator.StatusSkipped, results["b"])
	s.Equal(orchestrator.StatusSkipped, results["c"]) // transitive skip
}

func (s *PlanIntegrationSuite) TestRetryStrategy() {
	attempts := 0

	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Retry(2)),
	)

	plan.TaskFunc("flaky", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		attempts++
		if attempts < 3 {
			return nil, fmt.Errorf("attempt %d failed", attempts)
		}

		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.Equal(3, attempts) // 1 initial + 2 retries
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestRetryExhausted() {
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Retry(1)),
	)

	plan.TaskFunc("always-fail", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("permanent failure")
	})

	report, err := plan.Run(context.Background())
	s.Error(err)
	s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestRunCycleDetection() {
	plan := orchestrator.NewPlan(nil)
	a := plan.Task("a", &orchestrator.Op{Operation: "noop"})
	b := plan.Task("b", &orchestrator.Op{Operation: "noop"})
	a.DependsOn(b)
	b.DependsOn(a)

	_, err := plan.Run(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "cycle")
}

func (s *PlanIntegrationSuite) TestRunOpTaskRequiresClient() {
	plan := orchestrator.NewPlan(nil)
	plan.Task("install", &orchestrator.Op{
		Operation: "command.exec",
		Target:    "_any",
		Params:    map[string]any{"command": "uptime"},
	})

	report, err := plan.Run(context.Background())
	s.Error(err)
	s.NotNil(report)
	s.Contains(report.Tasks[0].Error.Error(), "requires an OSAPI client")
}

func (s *PlanIntegrationSuite) TestPerTaskOnError() {
	plan := orchestrator.NewPlan(nil) // default StopAll

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("a failed")
	})
	a.OnError(orchestrator.Continue) // override: keep going

	plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err) // Continue on a, so no error returned

	results := make(map[string]orchestrator.Status)
	for _, r := range report.Tasks {
		results[r.Name] = r.Status
	}

	s.Equal(orchestrator.StatusFailed, results["a"])
	s.Equal(orchestrator.StatusChanged, results["b"])
}

func (s *PlanIntegrationSuite) TestPerTaskRetry() {
	attempts := 0

	plan := orchestrator.NewPlan(nil) // default StopAll

	plan.TaskFunc("flaky", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		attempts++
		if attempts < 2 {
			return nil, fmt.Errorf("attempt %d failed", attempts)
		}

		return &orchestrator.Result{Changed: true}, nil
	}).OnError(orchestrator.Retry(1))

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.Equal(2, attempts)
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestOpTaskCreateAndPoll() {
	pollCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000001",
				"status": "pending",
			})
		case r.Method == "GET" && r.URL.Path == "/job/00000000-0000-0000-0000-000000000001":
			pollCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			status := "pending"
			if pollCount >= 2 {
				status = "completed"
			}

			resp := map[string]any{
				"id":     "00000000-0000-0000-0000-000000000001",
				"status": status,
			}
			if status == "completed" {
				resp["result"] = map[string]any{"hostname": "web-01"}
			}

			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Use a very short poll interval for tests.
	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("get-hostname", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Len(report.Tasks, 1)
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
	s.GreaterOrEqual(pollCount, 2)
}

func (s *PlanIntegrationSuite) TestOpTaskCreateFailed() {
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("fail", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	report, err := plan.Run(context.Background())
	s.Error(err)
	s.NotNil(report)
	s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestOpTaskJobFailed() {
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000002",
				"status": "pending",
			})
		case r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			errMsg := "disk full"
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "00000000-0000-0000-0000-000000000002",
				"status": "failed",
				"error":  errMsg,
			})
		}
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("fail-job", &orchestrator.Op{
		Operation: "node.disk.get",
		Target:    "_any",
	})

	report, err := plan.Run(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "disk full")
	s.Equal(orchestrator.StatusFailed, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestOpTaskJobFailedNoErrorMessage() {
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000003",
				"status": "pending",
			})
		case r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "00000000-0000-0000-0000-000000000003",
				"status": "failed",
			})
		}
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("fail-no-msg", &orchestrator.Op{
		Operation: "node.disk.get",
		Target:    "_any",
	})

	_, err = plan.Run(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "job failed")
}

func (s *PlanIntegrationSuite) TestOpTaskPollContextCanceled() {
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method == "POST" && r.URL.Path == "/job" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000004",
				"status": "pending",
			})

			return
		}

		// Always return pending so the poller loops.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "00000000-0000-0000-0000-000000000004",
			"status": "pending",
		})
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("cancel-me", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = plan.Run(ctx)
	s.Error(err)
	s.ErrorIs(err, context.DeadlineExceeded)
}

func (s *PlanIntegrationSuite) TestOpTaskPollError() {
	requestCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		requestCount++

		if r.Method == "POST" && r.URL.Path == "/job" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000005",
				"status": "pending",
			})

			return
		}

		// Return bad status on poll.
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("poll-fail", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	_, err = plan.Run(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "unexpected status")
}

func (s *PlanIntegrationSuite) TestOpTaskCompletedWithNoResult() {
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000006",
				"status": "pending",
			})
		case r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "00000000-0000-0000-0000-000000000006",
				"status": "completed",
			})
		}
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("no-result", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestOpTaskCompletedWithNonMapResult() {
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000007",
				"status": "pending",
			})
		case r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "00000000-0000-0000-0000-000000000007",
				"status": "completed",
				"result": "just a string",
			})
		}
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("string-result", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestOpTaskPollNilStatus() {
	pollCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000008",
				"status": "pending",
			})
		case r.Method == "GET":
			pollCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			resp := map[string]any{
				"id": "00000000-0000-0000-0000-000000000008",
			}
			if pollCount >= 2 {
				resp["status"] = "completed"
			}
			// First poll: no status field (nil), second: completed.
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("nil-status", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestRetryWithNoHooks() {
	attempts := 0

	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Retry(1)),
	)

	plan.TaskFunc("flaky", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		attempts++
		if attempts < 2 {
			return nil, fmt.Errorf("attempt %d", attempts)
		}

		return &orchestrator.Result{Changed: true}, nil
	})

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.Equal(2, attempts)
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)
}

func (s *PlanIntegrationSuite) TestSkipWithNoHooks() {
	plan := orchestrator.NewPlan(
		nil,
		orchestrator.OnError(orchestrator.Continue),
	)

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("a failed")
	})

	plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	}).DependsOn(a)

	report, err := plan.Run(context.Background())
	s.NoError(err)

	results := make(map[string]orchestrator.Status)
	for _, r := range report.Tasks {
		results[r.Name] = r.Status
	}

	s.Equal(orchestrator.StatusSkipped, results["b"])
}

func (s *PlanIntegrationSuite) TestOnRetryHookCalled() {
	var retryEvents []string
	attempts := 0

	hooks := orchestrator.Hooks{
		OnRetry: func(
			task *orchestrator.Task,
			attempt int,
			err error,
		) {
			retryEvents = append(
				retryEvents,
				fmt.Sprintf("retry-%s-%d-%s", task.Name(), attempt, err),
			)
		},
	}

	plan := orchestrator.NewPlan(
		nil,
		orchestrator.WithHooks(hooks),
		orchestrator.OnError(orchestrator.Retry(2)),
	)

	plan.TaskFunc("flaky", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		attempts++
		if attempts < 3 {
			return nil, fmt.Errorf("fail-%d", attempts)
		}

		return &orchestrator.Result{Changed: true}, nil
	})

	_, err := plan.Run(context.Background())
	s.NoError(err)
	s.Len(retryEvents, 2)
	s.Contains(retryEvents[0], "retry-flaky-1-fail-1")
	s.Contains(retryEvents[1], "retry-flaky-2-fail-2")
}

func (s *PlanIntegrationSuite) TestOnSkipHookCalled() {
	var skipEvents []string

	hooks := orchestrator.Hooks{
		OnSkip: func(task *orchestrator.Task, reason string) {
			skipEvents = append(
				skipEvents,
				fmt.Sprintf("skip-%s-%s", task.Name(), reason),
			)
		},
	}

	plan := orchestrator.NewPlan(
		nil,
		orchestrator.WithHooks(hooks),
		orchestrator.OnError(orchestrator.Continue),
	)

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return nil, fmt.Errorf("a failed")
	})

	plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	}).DependsOn(a)

	_, err := plan.Run(context.Background())
	s.NoError(err)
	s.Len(skipEvents, 1)
	s.Contains(skipEvents[0], "skip-b-dependency failed")
}

func (s *PlanIntegrationSuite) TestOnSkipHookCalledForGuard() {
	var skipEvents []string

	hooks := orchestrator.Hooks{
		OnSkip: func(task *orchestrator.Task, reason string) {
			skipEvents = append(
				skipEvents,
				fmt.Sprintf("skip-%s-%s", task.Name(), reason),
			)
		},
	}

	plan := orchestrator.NewPlan(nil, orchestrator.WithHooks(hooks))

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	b := plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})
	b.DependsOn(a)
	b.When(func(_ orchestrator.Results) bool { return false })

	_, err := plan.Run(context.Background())
	s.NoError(err)
	s.Len(skipEvents, 1)
	s.Contains(skipEvents[0], "guard returned false")
}

func (s *PlanIntegrationSuite) TestOnSkipHookCalledForOnlyIfChanged() {
	var skipEvents []string

	hooks := orchestrator.Hooks{
		OnSkip: func(task *orchestrator.Task, reason string) {
			skipEvents = append(
				skipEvents,
				fmt.Sprintf("skip-%s-%s", task.Name(), reason),
			)
		},
	}

	plan := orchestrator.NewPlan(nil, orchestrator.WithHooks(hooks))

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	b := plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})
	b.DependsOn(a)
	b.OnlyIfChanged()

	_, err := plan.Run(context.Background())
	s.NoError(err)
	s.Len(skipEvents, 1)
	s.Contains(skipEvents[0], "no dependencies changed")
}

func (s *PlanIntegrationSuite) TestOnlyIfChangedRunsWhenDepChanged() {
	ran := false

	plan := orchestrator.NewPlan(nil)

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})

	b := plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		ran = true

		return &orchestrator.Result{Changed: false}, nil
	})
	b.DependsOn(a)
	b.OnlyIfChanged()

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.True(ran)
	s.Equal(orchestrator.StatusUnchanged, report.Tasks[1].Status)
}

func (s *PlanIntegrationSuite) TestOpTaskWithParams() {
	var receivedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/job":
			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-00000000000a",
				"status": "pending",
			})
		case r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":     "00000000-0000-0000-0000-00000000000a",
				"status": "completed",
			})
		}
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("exec", &orchestrator.Op{
		Operation: "command.exec.execute",
		Target:    "_any",
		Params: map[string]any{
			"command": "uptime",
			"args":    []string{"-s"},
		},
	})

	report, err := plan.Run(context.Background())
	s.Require().NoError(err)
	s.Equal(orchestrator.StatusChanged, report.Tasks[0].Status)

	// Verify params were included in the request.
	op, ok := receivedBody["operation"].(map[string]any)
	s.Require().True(ok)
	s.Equal("command.exec.execute", op["type"])
	s.Equal("uptime", op["command"])
}

func (s *PlanIntegrationSuite) TestOpTaskCreateNetworkError() {
	// Server that immediately closes connections.
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("net-error", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	_, err = plan.Run(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "create job")
}

func (s *PlanIntegrationSuite) TestOpTaskPollNetworkError() {
	srv := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if r.Method == "POST" && r.URL.Path == "/job" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"job_id": "00000000-0000-0000-0000-000000000009",
				"status": "pending",
			})

			return
		}

		// Close connection on poll.
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer srv.Close()

	origInterval := orchestrator.DefaultPollInterval
	orchestrator.DefaultPollInterval = 10 * time.Millisecond
	defer func() { orchestrator.DefaultPollInterval = origInterval }()

	client, err := osapi.New(srv.URL, "test-token")
	s.Require().NoError(err)

	plan := orchestrator.NewPlan(client)
	plan.Task("poll-net-error", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})

	_, err = plan.Run(context.Background())
	s.Error(err)
	s.Contains(err.Error(), "poll job")
}

func (s *PlanIntegrationSuite) TestHooksCalledDuringRun() {
	var events []string
	hooks := orchestrator.Hooks{
		BeforePlan: func(_ string) {
			events = append(events, "before-plan")
		},
		AfterPlan: func(_ *orchestrator.Report) {
			events = append(events, "after-plan")
		},
		BeforeLevel: func(level int, _ []*orchestrator.Task, _ bool) {
			events = append(events, fmt.Sprintf("before-level-%d", level))
		},
		AfterLevel: func(level int, _ []orchestrator.TaskResult) {
			events = append(events, fmt.Sprintf("after-level-%d", level))
		},
		BeforeTask: func(task *orchestrator.Task) {
			events = append(events, "before-"+task.Name())
		},
		AfterTask: func(_ *orchestrator.Task, r orchestrator.TaskResult) {
			events = append(events, "after-"+r.Name)
		},
	}

	plan := orchestrator.NewPlan(nil, orchestrator.WithHooks(hooks))

	a := plan.TaskFunc("a", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: true}, nil
	})

	plan.TaskFunc("b", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	}).DependsOn(a)

	report, err := plan.Run(context.Background())
	s.NoError(err)
	s.NotNil(report)

	s.Equal([]string{
		"before-plan",
		"before-level-0",
		"before-a",
		"after-a",
		"after-level-0",
		"before-level-1",
		"before-b",
		"after-b",
		"after-level-1",
		"after-plan",
	}, events)
}
