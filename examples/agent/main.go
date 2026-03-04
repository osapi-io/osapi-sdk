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

// Package main demonstrates agent facts retrieval using the orchestrator.
//
// Lists all registered agents, picks the first one, retrieves its detailed
// facts (OS info, CPU, memory, interfaces, labels), and prints a summary.
//
// DAG:
//
//	check-health
//	    └── list-agents ── get-agent-details (When: agents found)
//	                           └── print-facts (TaskFuncWithResults)
//
// Run with: OSAPI_TOKEN="<jwt>" go run main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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
			fmt.Printf("Plan: %d tasks, %d steps\n\n", summary.TotalTasks, len(summary.Steps))
		},
		AfterPlan: func(report *orchestrator.Report) {
			fmt.Printf("\n=== %s in %s ===\n", report.Summary(), report.Duration)
		},
		BeforeTask: func(task *orchestrator.Task) {
			fmt.Printf("  [start] %s\n", task.Name())
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
		OnSkip: func(task *orchestrator.Task, reason string) {
			fmt.Printf("  [skip] %s  reason=%q\n", task.Name(), reason)
		},
	}

	plan := orchestrator.NewPlan(client, orchestrator.WithHooks(hooks))

	// Level 0: health check.
	checkHealth := plan.TaskFunc(
		"check-health",
		func(
			ctx context.Context,
			c *osapi.Client,
		) (*orchestrator.Result, error) {
			_, err := c.Health.Liveness(ctx)
			if err != nil {
				return nil, fmt.Errorf("health check: %w", err)
			}

			return &orchestrator.Result{Changed: false}, nil
		},
	)

	// Level 1: list all agents and capture hostnames.
	listAgents := plan.TaskFunc(
		"list-agents",
		func(
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
				Changed: true,
				Data: map[string]any{
					"total":     resp.Data.Total,
					"hostnames": hostnames,
				},
			}, nil
		},
	)
	listAgents.DependsOn(checkHealth)

	// Level 2: get detailed facts for the first agent.
	getDetails := plan.TaskFuncWithResults(
		"get-agent-details",
		func(
			ctx context.Context,
			c *osapi.Client,
			results orchestrator.Results,
		) (*orchestrator.Result, error) {
			r := results.Get("list-agents")
			if r == nil {
				return &orchestrator.Result{Changed: false}, nil
			}

			hostnames, _ := r.Data["hostnames"].([]any)
			if len(hostnames) == 0 {
				return &orchestrator.Result{Changed: false}, nil
			}

			hostname, _ := hostnames[0].(string)

			resp, err := c.Agent.Get(ctx, hostname)
			if err != nil {
				return nil, fmt.Errorf("get agent %s: %w", hostname, err)
			}

			// Marshal the domain type into Result.Data.
			var data map[string]any

			b, err := json.Marshal(resp.Data)
			if err != nil {
				return nil, fmt.Errorf("marshal agent: %w", err)
			}

			if err := json.Unmarshal(b, &data); err != nil {
				return nil, fmt.Errorf("unmarshal agent: %w", err)
			}

			return &orchestrator.Result{
				Changed: true,
				Data:    data,
			}, nil
		},
	)
	getDetails.DependsOn(listAgents)
	getDetails.When(func(results orchestrator.Results) bool {
		r := results.Get("list-agents")
		if r == nil {
			return false
		}

		total, _ := r.Data["total"].(int)

		return total > 0
	})

	// Level 3: print agent facts from prior results.
	printFacts := plan.TaskFuncWithResults(
		"print-facts",
		func(
			_ context.Context,
			_ *osapi.Client,
			results orchestrator.Results,
		) (*orchestrator.Result, error) {
			r := results.Get("get-agent-details")
			if r == nil {
				return &orchestrator.Result{Changed: false}, nil
			}

			fmt.Println("\n  --- Agent Facts ---")

			if h, ok := r.Data["hostname"].(string); ok {
				fmt.Printf("  Hostname:       %s\n", h)
			}

			if s, ok := r.Data["status"].(string); ok {
				fmt.Printf("  Status:         %s\n", s)
			}

			if arch, ok := r.Data["architecture"].(string); ok {
				fmt.Printf("  Architecture:   %s\n", arch)
			}

			if kv, ok := r.Data["kernel_version"].(string); ok {
				fmt.Printf("  Kernel:         %s\n", kv)
			}

			if fqdn, ok := r.Data["fqdn"].(string); ok {
				fmt.Printf("  FQDN:           %s\n", fqdn)
			}

			if cpu, ok := r.Data["cpu_count"].(float64); ok {
				fmt.Printf("  CPUs:           %d\n", int(cpu))
			}

			if sm, ok := r.Data["service_mgr"].(string); ok {
				fmt.Printf("  Service Mgr:    %s\n", sm)
			}

			if pm, ok := r.Data["package_mgr"].(string); ok {
				fmt.Printf("  Package Mgr:    %s\n", pm)
			}

			if ifaces, ok := r.Data["interfaces"].([]any); ok {
				fmt.Printf("  Interfaces:     %d\n", len(ifaces))
				for _, iface := range ifaces {
					if m, ok := iface.(map[string]any); ok {
						name, _ := m["name"].(string)
						ipv4, _ := m["ipv4"].(string)
						fmt.Printf("    - %s: %s\n", name, ipv4)
					}
				}
			}

			return &orchestrator.Result{Changed: false}, nil
		},
	)
	printFacts.DependsOn(getDetails)

	// Run the plan.
	_, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}
