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

// Package main demonstrates orchestrator usage with a webserver deployment plan.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/osapi-io/osapi-sdk/pkg/orchestrator"
)

func main() {
	plan := orchestrator.NewPlan()

	createUser := plan.Task("create-user", &orchestrator.Op{
		Operation: "command.exec",
		Target:    "_all",
		Params: map[string]any{
			"command": "useradd",
			"args":    []string{"-m", "www"},
		},
	})

	installNginx := plan.Task("install-nginx", &orchestrator.Op{
		Operation: "command.exec",
		Target:    "_all",
		Params: map[string]any{
			"command": "apt",
			"args":    []string{"install", "-y", "nginx"},
		},
	})

	configureDNS := plan.Task("configure-dns", &orchestrator.Op{
		Operation: "network.dns.update",
		Target:    "_all",
		Params: map[string]any{
			"address": "8.8.8.8",
		},
	})

	startNginx := plan.Task("start-nginx", &orchestrator.Op{
		Operation: "command.exec",
		Target:    "_all",
		Params: map[string]any{
			"command": "systemctl",
			"args":    []string{"start", "nginx"},
		},
	})

	installNginx.DependsOn(createUser)
	startNginx.DependsOn(installNginx, configureDNS)
	startNginx.OnlyIfChanged()

	report, err := plan.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(report.Summary())

	for _, r := range report.Tasks {
		fmt.Printf(
			"%s: %s (changed=%v, duration=%s)\n",
			r.Name,
			r.Status,
			r.Changed,
			r.Duration,
		)
	}
}
