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
	"fmt"

	"github.com/google/uuid"

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
)

// JobService provides job queue operations.
type JobService struct {
	client *gen.ClientWithResponses
}

// Create creates a new job with the given operation and target.
func (s *JobService) Create(
	ctx context.Context,
	operation map[string]interface{},
	target string,
) (*gen.PostJobResponse, error) {
	body := gen.CreateJobRequest{
		Operation:      operation,
		TargetHostname: target,
	}

	return s.client.PostJobWithResponse(ctx, body)
}

// Get retrieves a job by ID.
func (s *JobService) Get(
	ctx context.Context,
	id string,
) (*gen.GetJobByIDResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	return s.client.GetJobByIDWithResponse(ctx, parsedID)
}

// Delete deletes a job by ID.
func (s *JobService) Delete(
	ctx context.Context,
	id string,
) (*gen.DeleteJobByIDResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	return s.client.DeleteJobByIDWithResponse(ctx, parsedID)
}

// ListParams contains optional filters for listing jobs.
type ListParams struct {
	// Status filters by job status (e.g., "pending", "completed").
	Status string

	// Limit is the maximum number of results. Zero uses server default.
	Limit int

	// Offset is the number of results to skip. Zero starts from the
	// beginning.
	Offset int
}

// List retrieves jobs, optionally filtered by status.
func (s *JobService) List(
	ctx context.Context,
	params ListParams,
) (*gen.GetJobResponse, error) {
	p := &gen.GetJobParams{}

	if params.Status != "" {
		status := gen.GetJobParamsStatus(params.Status)
		p.Status = &status
	}

	if params.Limit > 0 {
		p.Limit = &params.Limit
	}

	if params.Offset > 0 {
		p.Offset = &params.Offset
	}

	return s.client.GetJobWithResponse(ctx, p)
}

// QueueStats retrieves job queue statistics.
func (s *JobService) QueueStats(
	ctx context.Context,
) (*gen.GetJobStatusResponse, error) {
	return s.client.GetJobStatusWithResponse(ctx)
}

// Retry retries a failed job by ID, optionally on a different target.
func (s *JobService) Retry(
	ctx context.Context,
	id string,
	target string,
) (*gen.RetryJobByIDResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid job ID: %w", err)
	}

	body := gen.RetryJobByIDJSONRequestBody{}
	if target != "" {
		body.TargetHostname = &target
	}

	return s.client.RetryJobByIDWithResponse(ctx, parsedID, body)
}
