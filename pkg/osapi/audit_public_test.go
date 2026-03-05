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

type AuditPublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *AuditPublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *AuditPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		limit        int
		offset       int
		validateFunc func(*osapi.Response[osapi.AuditList], error)
	}{
		{
			name:   "when listing audit entries returns audit list",
			limit:  20,
			offset: 0,
			validateFunc: func(resp *osapi.Response[osapi.AuditList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(0, resp.Data.TotalItems)
				suite.Empty(resp.Data.Items)
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

			resp, err := sut.Audit.List(suite.ctx, tc.limit, tc.offset)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *AuditPublicTestSuite) TestListError() {
	tests := []struct {
		name         string
		setupServer  func() *httptest.Server
		validateFunc func(*osapi.Response[osapi.AuditList], error)
	}{
		{
			name: "when server returns 401 returns AuthError",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
					}),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusUnauthorized, target.StatusCode)
			},
		},
		{
			name: "when HTTP request fails returns error",
			setupServer: func() *httptest.Server {
				server := httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}),
				)
				server.Close()

				return server
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditList], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "list audit logs:")
			},
		},
		{
			name: "when response body is nil returns UnexpectedStatusError",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusOK)
					}),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := tc.setupServer()
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Audit.List(suite.ctx, 20, 0)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *AuditPublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		id           string
		validateFunc func(*osapi.Response[osapi.AuditEntry], error)
	}{
		{
			name: "when valid UUID returns audit entry",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			validateFunc: func(resp *osapi.Response[osapi.AuditEntry], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("550e8400-e29b-41d4-a716-446655440000", resp.Data.ID)
				suite.Equal("admin", resp.Data.User)
				suite.Equal("GET", resp.Data.Method)
				suite.Equal("/api/v1/health", resp.Data.Path)
			},
		},
		{
			name: "when invalid UUID returns error",
			id:   "not-a-uuid",
			validateFunc: func(resp *osapi.Response[osapi.AuditEntry], err error) {
				suite.Error(err)
				suite.Nil(resp)
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
							`{"entry":{"id":"550e8400-e29b-41d4-a716-446655440000","timestamp":"2026-01-01T00:00:00Z","user":"admin","roles":["admin"],"method":"GET","path":"/api/v1/health","response_code":200,"duration_ms":5,"source_ip":"127.0.0.1"}}`,
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

			resp, err := sut.Audit.Get(suite.ctx, tc.id)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *AuditPublicTestSuite) TestGetError() {
	tests := []struct {
		name         string
		setupServer  func() *httptest.Server
		validateFunc func(*osapi.Response[osapi.AuditEntry], error)
	}{
		{
			name: "when server returns 404 returns NotFoundError",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						_, _ = w.Write([]byte(`{"error":"audit entry not found"}`))
					}),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditEntry], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
			},
		},
		{
			name: "when HTTP request fails returns error",
			setupServer: func() *httptest.Server {
				server := httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}),
				)
				server.Close()

				return server
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditEntry], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get audit log")
			},
		},
		{
			name: "when response body is nil returns UnexpectedStatusError",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusOK)
					}),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditEntry], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := tc.setupServer()
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Audit.Get(suite.ctx, "550e8400-e29b-41d4-a716-446655440000")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *AuditPublicTestSuite) TestExport() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.AuditList], error)
	}{
		{
			name: "when exporting audit entries returns audit list",
			validateFunc: func(resp *osapi.Response[osapi.AuditList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal(0, resp.Data.TotalItems)
				suite.Empty(resp.Data.Items)
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

			resp, err := sut.Audit.Export(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *AuditPublicTestSuite) TestExportError() {
	tests := []struct {
		name         string
		setupServer  func() *httptest.Server
		validateFunc func(*osapi.Response[osapi.AuditList], error)
	}{
		{
			name: "when server returns 401 returns AuthError",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusUnauthorized)
						_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
					}),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusUnauthorized, target.StatusCode)
			},
		},
		{
			name: "when HTTP request fails returns error",
			setupServer: func() *httptest.Server {
				server := httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}),
				)
				server.Close()

				return server
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditList], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "export audit logs:")
			},
		},
		{
			name: "when response body is nil returns UnexpectedStatusError",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusOK)
					}),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.AuditList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Contains(target.Message, "nil response body")
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := tc.setupServer()
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Audit.Export(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func TestAuditPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AuditPublicTestSuite))
}
