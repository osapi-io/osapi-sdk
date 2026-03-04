// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

// Package main demonstrates a runnable orchestrator plan that discovers
// fleet information from a live OSAPI instance.
//
// DAG:
//
//	check-health
//	    ├── list-agents ───────┐
//	    ├── get-status  ───────┤
//	    ├── get-load    ───┐   │
//	    └── get-memory  ───┴───┴── print-summary (when: agents found)
//
// Run with: OSAPI_TOKEN="<jwt>" go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
	"github.com/osapi-io/osapi-sdk/pkg/osapi"
)

func main() {
	url := os.Getenv("OSAPI_URL")
	if url == "" {
		url = "http://localhost:8080"
	}

	token := os.Getenv("OSAPI_TOKEN")
	if token == "" {
		log.Fatal("OSAPI_TOKEN environment variable is required")
	}

	client := osapi.New(url, token)

	hooks := orchestrator.Hooks{
		BeforePlan: func(summary orchestrator.PlanSummary) {
			fmt.Println("=== Execution Plan ===")
			fmt.Printf("Plan: %d tasks, %d steps\n", summary.TotalTasks, len(summary.Steps))
			fmt.Println()
		},
		AfterPlan: func(report *orchestrator.Report) {
			fmt.Printf(
				"\n=== Complete: %s in %s ===\n",
				report.Summary(),
				report.Duration,
			)
		},
		BeforeLevel: func(
			level int,
			tasks []*orchestrator.Task,
			parallel bool,
		) {
			names := make([]string, len(tasks))
			for i, t := range tasks {
				names[i] = t.Name()
			}

			mode := "sequential"
			if parallel {
				mode = "parallel"
			}

			fmt.Printf(
				"\n>>> Step %d (%s): %s\n",
				level+1,
				mode,
				strings.Join(names, ", "),
			)
		},
		AfterLevel: func(level int, results []orchestrator.TaskResult) {
			changed := 0
			for _, r := range results {
				if r.Changed {
					changed++
				}
			}

			fmt.Printf(
				"<<< Step %d done: %d/%d changed\n",
				level+1,
				changed,
				len(results),
			)
		},
		BeforeTask: func(task *orchestrator.Task) {
			fmt.Printf("  [start] %s  (custom function)\n", task.Name())
		},
		AfterTask: func(
			_ *orchestrator.Task,
			result orchestrator.TaskResult,
		) {
			fmt.Printf(
				"  [%s] %s  changed=%v duration=%s\n",
				result.Status,
				result.Name,
				result.Changed,
				result.Duration,
			)
		},
		OnRetry: func(
			task *orchestrator.Task,
			attempt int,
			err error,
		) {
			fmt.Printf(
				"  [retry] %s  attempt=%d error=%q\n",
				task.Name(),
				attempt,
				err,
			)
		},
		OnSkip: func(task *orchestrator.Task, reason string) {
			fmt.Printf("  [skip] %s  reason=%q\n", task.Name(), reason)
		},
	}

	plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))

	// Level 0: verify OSAPI is healthy before doing anything.
	checkHealth := plan.TaskFunc("check-health", func(
		ctx context.Context,
		c *osapi.Client,
	) (*orchestrator.Result, error) {
		_, err := c.Health.Liveness(ctx)
		if err != nil {
			return nil, fmt.Errorf("health check: %w", err)
		}

		return &orchestrator.Result{Changed: false}, nil
	})

	// Level 1: discover agents (parallel with others).
	listAgents := plan.TaskFunc("list-agents", func(
		ctx context.Context,
		c *osapi.Client,
	) (*orchestrator.Result, error) {
		resp, err := c.Agent.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list agents: %w", err)
		}

		hostnames := make([]string, len(resp.Data.Agents))
		for i, a := range resp.Data.Agents {
			hostnames[i] = a.Hostname
		}

		return &orchestrator.Result{
			Changed: false,
			Data: map[string]any{
				"total":     resp.Data.Total,
				"hostnames": hostnames,
			},
		}, nil
	})

	// Level 1: get system status (parallel with list-agents).
	getStatus := plan.TaskFunc("get-status", func(
		ctx context.Context,
		c *osapi.Client,
	) (*orchestrator.Result, error) {
		resp, err := c.Health.Status(ctx)
		if err != nil {
			return nil, fmt.Errorf("get status: %w", err)
		}

		return &orchestrator.Result{
			Changed: false,
			Data: map[string]any{
				"status":  resp.Data.Status,
				"version": resp.Data.Version,
				"uptime":  resp.Data.Uptime,
			},
		}, nil
	})

	// Level 1: get load averages (parallel with list-agents).
	getLoad := plan.TaskFunc("get-load", func(
		ctx context.Context,
		c *osapi.Client,
	) (*orchestrator.Result, error) {
		_, err := c.Node.Load(ctx, "_any")
		if err != nil {
			return nil, fmt.Errorf("get load: %w", err)
		}

		return &orchestrator.Result{Changed: false}, nil
	})

	// Level 1: get memory info (parallel with list-agents).
	getMemory := plan.TaskFunc("get-memory", func(
		ctx context.Context,
		c *osapi.Client,
	) (*orchestrator.Result, error) {
		_, err := c.Node.Memory(ctx, "_any")
		if err != nil {
			return nil, fmt.Errorf("get memory: %w", err)
		}

		return &orchestrator.Result{Changed: false}, nil
	})

	// Level 2: print summary — only if agents were found.
	summary := plan.TaskFunc("print-summary", func(
		_ context.Context,
		_ *osapi.Client,
	) (*orchestrator.Result, error) {
		return &orchestrator.Result{Changed: false}, nil
	})

	// Wire dependencies.
	listAgents.DependsOn(checkHealth)
	getStatus.DependsOn(checkHealth)
	getLoad.DependsOn(checkHealth)
	getMemory.DependsOn(checkHealth)
	summary.DependsOn(listAgents, getStatus, getLoad, getMemory)

	// Only print summary if at least one agent was found.
	summary.When(func(results orchestrator.Results) bool {
		r := results.Get("list-agents")
		if r == nil {
			return false
		}

		total, _ := r.Data["total"].(int)

		return total > 0
	})

	// Run the plan.
	_, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
