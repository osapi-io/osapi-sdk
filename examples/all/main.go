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

// Package main demonstrates every orchestrator feature: hooks for consumer-
// controlled logging at every lifecycle point, Op and TaskFunc tasks,
// dependencies, guards, Levels() for DAG inspection, error strategies,
// and detailed result reporting.
//
// This example serves as a reference for building tools like Terraform
// or Ansible that consume the SDK.
//
// DAG:
//
//	check-health
//	    ├── get-hostname ───┐
//	    ├── get-disk     ───┤
//	    ├── get-memory   ───┤
//	    └── get-load     ───┴── print-summary (when: hostname found)
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
	token := os.Getenv("OSAPI_TOKEN")
	if token == "" {
		log.Fatal("OSAPI_TOKEN is required")
	}

	url := os.Getenv("OSAPI_URL")
	if url == "" {
		url = "http://localhost:8080"
	}

	client, err := osapi.New(url, token)
	if err != nil {
		log.Fatal(err)
	}

	// --- Hooks: consumer-controlled logging at every stage ---

	hooks := orchestrator.Hooks{
		BeforePlan: func(explain string) {
			fmt.Println("=== Execution Plan ===")
			fmt.Print(explain)
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
				"\n>>> Level %d (%s): %s\n",
				level,
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
				"<<< Level %d done: %d/%d changed\n",
				level,
				changed,
				len(results),
			)
		},
		BeforeTask: func(task *orchestrator.Task) {
			if op := task.Operation(); op != nil {
				fmt.Printf(
					"  [start] %s  operation=%s target=%s\n",
					task.Name(),
					op.Operation,
					op.Target,
				)
			} else {
				fmt.Printf("  [start] %s  (custom function)\n", task.Name())
			}
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

	// --- Task definitions ---

	// Level 0: health check (no deps, functional task)
	checkHealth := plan.TaskFunc(
		"check-health",
		func(
			ctx context.Context,
			c *osapi.Client,
		) (*orchestrator.Result, error) {
			resp, err := c.Health.Liveness(ctx)
			if err != nil {
				return nil, fmt.Errorf("health check: %w", err)
			}

			if resp.StatusCode() != 200 {
				return nil, fmt.Errorf(
					"unhealthy: status %d",
					resp.StatusCode(),
				)
			}

			return &orchestrator.Result{Changed: false}, nil
		},
	)

	// Level 1: parallel queries (all depend on health)
	getHostname := plan.Task("get-hostname", &orchestrator.Op{
		Operation: "node.hostname.get",
		Target:    "_any",
	})
	getHostname.DependsOn(checkHealth)

	getDisk := plan.Task("get-disk", &orchestrator.Op{
		Operation: "node.disk.get",
		Target:    "_any",
	})
	getDisk.DependsOn(checkHealth)

	getMemory := plan.Task("get-memory", &orchestrator.Op{
		Operation: "node.memory.get",
		Target:    "_any",
	})
	getMemory.DependsOn(checkHealth)

	getLoad := plan.Task("get-load", &orchestrator.Op{
		Operation: "node.load.get",
		Target:    "_any",
	})
	getLoad.DependsOn(checkHealth)
	getLoad.OnError(orchestrator.Continue)

	// Level 2: summary (depends on all queries, guard condition)
	summary := plan.TaskFunc(
		"print-summary",
		func(
			_ context.Context,
			_ *osapi.Client,
		) (*orchestrator.Result, error) {
			fmt.Println("\n  --- Fleet Summary ---")
			fmt.Println("  All node queries completed.")

			return &orchestrator.Result{Changed: false}, nil
		},
	)
	summary.DependsOn(getHostname, getDisk, getMemory, getLoad)
	summary.When(func(results orchestrator.Results) bool {
		return results.Get("get-hostname") != nil
	})

	// --- Structured DAG access ---

	levels, err := plan.Levels()
	if err != nil {
		log.Fatalf("invalid plan: %v", err)
	}

	fmt.Printf(
		"DAG: %d tasks across %d levels\n\n",
		len(plan.Tasks()),
		len(levels),
	)

	for i, level := range levels {
		names := make([]string, len(level))
		for j, t := range level {
			names[j] = t.Name()
		}

		fmt.Printf("  Level %d: %s\n", i, strings.Join(names, ", "))
	}

	fmt.Println()

	// --- Run ---

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatalf("plan failed: %v", err)
	}

	// --- Detailed result inspection ---

	fmt.Println("\nDetailed results:")

	for _, r := range report.Tasks {
		status := string(r.Status)
		if r.Error != nil {
			status += fmt.Sprintf(" (%s)", r.Error)
		}

		fmt.Printf(
			"  %-20s status=%-12s changed=%-5v duration=%s\n",
			r.Name,
			status,
			r.Changed,
			r.Duration,
		)
	}
}
