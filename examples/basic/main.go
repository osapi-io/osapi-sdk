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

// Package main demonstrates basic usage of the OSAPI SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

	client, err := osapi.New(url, token)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Get hostname from any available agent.
	hostnameResp, err := client.Node.Hostname(ctx, "_any")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Hostname response status: %d\n", hostnameResp.StatusCode())

	// Execute a command on any available agent.
	execResp, err := client.Node.Exec(ctx, osapi.ExecRequest{
		Command: "uptime",
		Target:  "_any",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Exec response status: %d\n", execResp.StatusCode())

	// List recent audit log entries.
	auditResp, err := client.Audit.List(ctx, 10, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Audit response status: %d\n", auditResp.StatusCode())

	// Check API server health.
	healthResp, err := client.Health.Liveness(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Health response status: %d\n", healthResp.StatusCode())
}
