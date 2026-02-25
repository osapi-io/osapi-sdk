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

package osapi

import (
	"context"

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
)

// HealthService provides health check operations.
type HealthService struct {
	client *gen.ClientWithResponses
}

// Liveness checks if the API server process is alive.
func (s *HealthService) Liveness(
	ctx context.Context,
) (*gen.GetHealthResponse, error) {
	return s.client.GetHealthWithResponse(ctx)
}

// Ready checks if the API server and its dependencies are ready to
// serve traffic.
func (s *HealthService) Ready(
	ctx context.Context,
) (*gen.GetHealthReadyResponse, error) {
	return s.client.GetHealthReadyWithResponse(ctx)
}

// Status returns detailed system status including component health,
// NATS info, stream stats, and job queue counts. Requires authentication.
func (s *HealthService) Status(
	ctx context.Context,
) (*gen.GetHealthStatusResponse, error) {
	return s.client.GetHealthStatusWithResponse(ctx)
}
