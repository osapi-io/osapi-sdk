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

type AgentPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *AgentPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *AgentPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.AgentList], error)
	}{
		{
			name: "when requesting agents returns no error",
			validateFunc: func(resp *osapi.Response[osapi.AgentList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(0, resp.Data.Total)
				suite.Empty(resp.Data.Agents)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"agents":[],"total":0}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Agent.List(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *AgentPublicTestSuite) TestListError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.AgentList], error)
	}{
		{
			name: "when server returns 401 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.AgentList], err error) {
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

			resp, err := sut.Agent.List(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *AgentPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		hostname     string
		validateFunc func(*osapi.Response[osapi.Agent], error)
	}{
		{
			name:     "when requesting agent details returns no error",
			hostname: "server1",
			validateFunc: func(resp *osapi.Response[osapi.Agent], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("server1", resp.Data.Hostname)
				suite.Equal("Ready", resp.Data.Status)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"hostname":"server1","status":"Ready"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Agent.Get(suite.ctx, tc.hostname)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *AgentPublicTestSuite) TestGetError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Agent], error)
	}{
		{
			name: "when server returns 404 returns NotFoundError",
			validateFunc: func(resp *osapi.Response[osapi.Agent], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
				suite.Equal("agent not found", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error":"agent not found"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Agent.Get(suite.ctx, "unknown-host")
			tc.validateFunc(resp, err)
		})
	}
}

func TestAgentPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AgentPublicTestSuite))
}
