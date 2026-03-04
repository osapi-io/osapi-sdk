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

type HealthPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *HealthPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *HealthPublicTestSuite) TestLiveness() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.HealthStatus], error)
	}{
		{
			name: "when checking liveness returns health status",
			validateFunc: func(resp *osapi.Response[osapi.HealthStatus], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("ok", resp.Data.Status)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"status":"ok"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Health.Liveness(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *HealthPublicTestSuite) TestReady() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.ReadyStatus], error)
	}{
		{
			name: "when checking readiness returns ready status",
			validateFunc: func(resp *osapi.Response[osapi.ReadyStatus], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("ready", resp.Data.Status)
				suite.False(resp.Data.ServiceUnavailable)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"status":"ready"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Health.Ready(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *HealthPublicTestSuite) TestReady503() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.ReadyStatus], error)
	}{
		{
			name: "when server returns 503 returns ready status with ServiceUnavailable",
			validateFunc: func(resp *osapi.Response[osapi.ReadyStatus], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("not_ready", resp.Data.Status)
				suite.Equal("nats down", resp.Data.Error)
				suite.True(resp.Data.ServiceUnavailable)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte(`{"status":"not_ready","error":"nats down"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Health.Ready(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *HealthPublicTestSuite) TestStatus() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.SystemStatus], error)
	}{
		{
			name: "when checking status returns system status",
			validateFunc: func(resp *osapi.Response[osapi.SystemStatus], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("ok", resp.Data.Status)
				suite.Equal("1.0.0", resp.Data.Version)
				suite.Equal("1h", resp.Data.Uptime)
				suite.False(resp.Data.ServiceUnavailable)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"status":"ok","version":"1.0.0","uptime":"1h"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Health.Status(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *HealthPublicTestSuite) TestStatus503() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.SystemStatus], error)
	}{
		{
			name: "when server returns 503 returns status with ServiceUnavailable",
			validateFunc: func(resp *osapi.Response[osapi.SystemStatus], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("degraded", resp.Data.Status)
				suite.True(resp.Data.ServiceUnavailable)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte(`{"status":"degraded","version":"1.0.0","uptime":"1h"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Health.Status(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *HealthPublicTestSuite) TestStatusAuthError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.SystemStatus], error)
	}{
		{
			name: "when server returns 401 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.SystemStatus], err error) {
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

			resp, err := sut.Health.Status(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func TestHealthPublicTestSuite(t *testing.T) {
	suite.Run(t, new(HealthPublicTestSuite))
}
