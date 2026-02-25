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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/osapi"
)

type JobPublicTestSuite struct {
	suite.Suite

	ctx    context.Context
	server *httptest.Server
	sut    *osapi.Client
}

func (suite *JobPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()

	suite.server = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}),
	)

	var err error
	suite.sut, err = osapi.New(
		suite.server.URL,
		"test-token",
		osapi.WithLogger(slog.Default()),
	)
	suite.Require().NoError(err)
}

func (suite *JobPublicTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *JobPublicTestSuite) TestCreate() {
	tests := []struct {
		name         string
		operation    map[string]interface{}
		target       string
		validateFunc func(error)
	}{
		{
			name:      "when creating job returns no error",
			operation: map[string]interface{}{"type": "system.hostname.get"},
			target:    "_any",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Job.Create(suite.ctx, tc.operation, tc.target)
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestGet() {
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
			_, err := suite.sut.Job.Get(suite.ctx, tc.id)
			tc.validateFunc(err)
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
			_, err := suite.sut.Job.Delete(suite.ctx, tc.id)
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		params       osapi.ListParams
		validateFunc func(error)
	}{
		{
			name:   "when no filters returns no error",
			params: osapi.ListParams{},
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name: "when all filters provided returns no error",
			params: osapi.ListParams{
				Status: "completed",
				Limit:  10,
				Offset: 5,
			},
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Job.List(suite.ctx, tc.params)
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestQueueStats() {
	tests := []struct {
		name         string
		validateFunc func(error)
	}{
		{
			name: "when requesting queue stats returns no error",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Job.QueueStats(suite.ctx)
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestWorkers() {
	tests := []struct {
		name         string
		validateFunc func(error)
	}{
		{
			name: "when requesting workers returns no error",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Job.Workers(suite.ctx)
			tc.validateFunc(err)
		})
	}
}

func (suite *JobPublicTestSuite) TestRetry() {
	tests := []struct {
		name         string
		id           string
		target       string
		validateFunc func(error)
	}{
		{
			name:   "when valid UUID with empty target returns no error",
			id:     "550e8400-e29b-41d4-a716-446655440000",
			target: "",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:   "when valid UUID with target returns no error",
			id:     "550e8400-e29b-41d4-a716-446655440000",
			target: "web-01",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:   "when invalid UUID returns error",
			id:     "not-a-uuid",
			target: "",
			validateFunc: func(err error) {
				suite.Error(err)
				suite.Contains(err.Error(), "invalid job ID")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Job.Retry(suite.ctx, tc.id, tc.target)
			tc.validateFunc(err)
		})
	}
}

func TestJobPublicTestSuite(t *testing.T) {
	suite.Run(t, new(JobPublicTestSuite))
}
