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

package osapi_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/osapi"
)

type JobPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *JobPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *JobPublicTestSuite) TestCreate() {
	tests := []struct {
		name         string
		operation    map[string]interface{}
		target       string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name:      "when creating job returns response",
			operation: map[string]interface{}{"type": "system.hostname.get"},
			target:    "_any",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", resp.Data.JobID)
				suite.Equal("pending", resp.Data.Status)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write(
						[]byte(
							`{"job_id":"550e8400-e29b-41d4-a716-446655440000","status":"pending"}`,
						),
					)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Create(suite.ctx, tc.operation, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestCreateError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name: "when server returns 400 returns ValidationError",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error":"validation failed"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Create(suite.ctx, map[string]interface{}{}, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestCreateHTTPError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name: "when HTTP request fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "create job")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
			)
			server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Create(suite.ctx, map[string]interface{}{}, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestCreateNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name: "when server returns 201 with empty body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Create(suite.ctx, map[string]interface{}{}, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		id           string
		validateFunc func(*osapi.Response[osapi.JobDetail], error)
	}{
		{
			name: "when valid UUID returns response",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", resp.Data.ID)
				suite.Equal("completed", resp.Data.Status)
			},
		},
		{
			name: "when invalid UUID returns error",
			id:   "not-a-uuid",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "invalid job ID")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(
						[]byte(
							`{"id":"550e8400-e29b-41d4-a716-446655440000","status":"completed"}`,
						),
					)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Get(suite.ctx, tc.id)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestGetHTTPError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobDetail], error)
	}{
		{
			name: "when HTTP request fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get job")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
			)
			server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Get(suite.ctx, "00000000-0000-0000-0000-000000000000")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestGetNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobDetail], error)
	}{
		{
			name: "when server returns 200 with empty body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Get(suite.ctx, "00000000-0000-0000-0000-000000000000")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestGetNotFound() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobDetail], error)
	}{
		{
			name: "when server returns 404 returns NotFoundError",
			validateFunc: func(resp *osapi.Response[osapi.JobDetail], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
				suite.Equal("job not found", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"job not found"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Get(suite.ctx, "550e8400-e29b-41d4-a716-446655440000")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestDelete() {
	tests := []struct {
		name         string
		id           string
		validateFunc func(error)
	}{
		{
			name: "when valid UUID returns no error",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name: "when invalid UUID returns error",
			id:   "not-a-uuid",
			validateFunc: func(err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "invalid job ID")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			err := sut.Job.Delete(suite.ctx, tc.id)
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestDeleteHTTPError() {
	tests := []struct {
		name         string
		validateFunc func(error)
	}{
		{
			name: "when HTTP request fails returns error",
			validateFunc: func(err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "delete job")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
			)
			server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			err := sut.Job.Delete(suite.ctx, "00000000-0000-0000-0000-000000000000")
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestDeleteNotFound() {
	tests := []struct {
		name         string
		validateFunc func(error)
	}{
		{
			name: "when server returns 404 returns NotFoundError",
			validateFunc: func(err error) {
				suite.Error(err)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
				suite.Equal("job not found", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"job not found"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			err := sut.Job.Delete(suite.ctx, "550e8400-e29b-41d4-a716-446655440000")
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		params       osapi.ListParams
		validateFunc func(*osapi.Response[osapi.JobList], error)
	}{
		{
			name:   "when no filters returns response",
			params: osapi.ListParams{},
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(0, resp.Data.TotalItems)
				suite.Empty(resp.Data.Items)
			},
		},
		{
			name: "when all filters provided returns response",
			params: osapi.ListParams{
				Status: "completed",
				Limit:  10,
				Offset: 5,
			},
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"items":[],"total_items":0}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.List(suite.ctx, tc.params)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestListHTTPError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobList], error)
	}{
		{
			name: "when HTTP request fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "list jobs")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
			)
			server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.List(suite.ctx, osapi.ListParams{})
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestListError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobList], error)
	}{
		{
			name: "when server returns 401 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusUnauthorized, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.List(suite.ctx, osapi.ListParams{})
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestListNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobList], error)
	}{
		{
			name: "when server returns 200 with empty body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.JobList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.List(suite.ctx, osapi.ListParams{})
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestQueueStats() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.QueueStats], error)
	}{
		{
			name: "when requesting queue stats returns response",
			validateFunc: func(resp *osapi.Response[osapi.QueueStats], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(5, resp.Data.TotalJobs)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"total_jobs":5}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.QueueStats(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestQueueStatsHTTPError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.QueueStats], error)
	}{
		{
			name: "when HTTP request fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.QueueStats], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "queue stats")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
			)
			server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.QueueStats(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestQueueStatsError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.QueueStats], error)
	}{
		{
			name: "when server returns 401 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.QueueStats], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusUnauthorized, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.QueueStats(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestQueueStatsNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.QueueStats], error)
	}{
		{
			name: "when server returns 200 with empty body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.QueueStats], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.QueueStats(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestRetry() {
	tests := []struct {
		name         string
		id           string
		target       string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name:   "when valid UUID with empty target returns response",
			id:     "550e8400-e29b-41d4-a716-446655440000",
			target: "",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", resp.Data.JobID)
				suite.Equal("pending", resp.Data.Status)
			},
		},
		{
			name:   "when valid UUID with target returns response",
			id:     "550e8400-e29b-41d4-a716-446655440000",
			target: "web-01",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name:   "when invalid UUID returns error",
			id:     "not-a-uuid",
			target: "",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "invalid job ID")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusCreated)
					_, _ = w.Write(
						[]byte(
							`{"job_id":"550e8400-e29b-41d4-a716-446655440000","status":"pending"}`,
						),
					)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Retry(suite.ctx, tc.id, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestRetryHTTPError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name: "when HTTP request fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "retry job")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}),
			)
			server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Retry(
				suite.ctx,
				"00000000-0000-0000-0000-000000000000",
				"",
			)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestRetryError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name: "when server returns 404 returns NotFoundError",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"job not found"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Retry(
				suite.ctx,
				"00000000-0000-0000-0000-000000000000",
				"",
			)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *JobPublicTestSuite) TestRetryNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.JobCreated], error)
	}{
		{
			name: "when server returns 201 with empty body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.JobCreated], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Job.Retry(
				suite.ctx,
				"00000000-0000-0000-0000-000000000000",
				"",
			)
			tc.validateFunc(resp, err)
		})
	}
}

func TestJobPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobPublicTestSuite))
}
