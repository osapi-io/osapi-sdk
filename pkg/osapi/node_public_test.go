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

type NodePublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *NodePublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *NodePublicTestSuite) TestHostname() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.HostnameResult]], error)
	}{
		{
			name:   "when requesting hostname returns results",
			target: "_any",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.HostnameResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("00000000-0000-0000-0000-000000000001", resp.Data.JobID)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("test-host", resp.Data.Results[0].Hostname)
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
							`{"job_id":"00000000-0000-0000-0000-000000000001","results":[{"hostname":"test-host"}]}`,
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

			resp, err := sut.Node.Hostname(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestHostnameClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.HostnameResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.HostnameResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get hostname")
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

			resp, err := sut.Node.Hostname(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestHostnameNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.HostnameResult]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.HostnameResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.Hostname(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestHostnameError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.HostnameResult]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.HostnameResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Hostname(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestStatus() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.NodeStatus]], error)
	}{
		{
			name:   "when requesting status returns results",
			target: "_any",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.NodeStatus]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("web-01", resp.Data.Results[0].Hostname)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"results":[{"hostname":"web-01"}]}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Status(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestStatusError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.NodeStatus]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Status(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestStatusClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.NodeStatus]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get status")
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

			resp, err := sut.Node.Status(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestStatusNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.NodeStatus]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.NodeStatus]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.Status(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestDisk() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DiskResult]], error)
	}{
		{
			name:   "when requesting disk returns results",
			target: "_any",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DiskResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("disk-host", resp.Data.Results[0].Hostname)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"results":[{"hostname":"disk-host"}]}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Disk(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestDiskError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DiskResult]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DiskResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Disk(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestDiskClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DiskResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DiskResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get disk")
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

			resp, err := sut.Node.Disk(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestDiskNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DiskResult]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DiskResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.Disk(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestMemory() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.MemoryResult]], error)
	}{
		{
			name:   "when requesting memory returns results",
			target: "_any",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.MemoryResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("mem-host", resp.Data.Results[0].Hostname)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"results":[{"hostname":"mem-host"}]}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Memory(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestMemoryError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.MemoryResult]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.MemoryResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Memory(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestMemoryClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.MemoryResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.MemoryResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get memory")
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

			resp, err := sut.Node.Memory(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestMemoryNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.MemoryResult]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.MemoryResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.Memory(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestLoad() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.LoadResult]], error)
	}{
		{
			name:   "when requesting load returns results",
			target: "_any",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.LoadResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("load-host", resp.Data.Results[0].Hostname)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"results":[{"hostname":"load-host"}]}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Load(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestLoadError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.LoadResult]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.LoadResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Load(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestLoadClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.LoadResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.LoadResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get load")
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

			resp, err := sut.Node.Load(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestLoadNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.LoadResult]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.LoadResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.Load(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestOS() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.OSInfoResult]], error)
	}{
		{
			name:   "when requesting OS info returns results",
			target: "_any",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.OSInfoResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("os-host", resp.Data.Results[0].Hostname)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"results":[{"hostname":"os-host"}]}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.OS(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestOSError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.OSInfoResult]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.OSInfoResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.OS(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestOSClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.OSInfoResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.OSInfoResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get os")
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

			resp, err := sut.Node.OS(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestOSNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.OSInfoResult]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.OSInfoResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.OS(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUptime() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.UptimeResult]], error)
	}{
		{
			name:   "when requesting uptime returns results",
			target: "_any",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.UptimeResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("uptime-host", resp.Data.Results[0].Hostname)
				suite.Equal("2d3h", resp.Data.Results[0].Uptime)
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
						[]byte(`{"results":[{"hostname":"uptime-host","uptime":"2d3h"}]}`),
					)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Uptime(suite.ctx, tc.target)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUptimeError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.UptimeResult]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.UptimeResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Uptime(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUptimeClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.UptimeResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.UptimeResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get uptime")
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

			resp, err := sut.Node.Uptime(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUptimeNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.UptimeResult]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.UptimeResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.Uptime(suite.ctx, "_any")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestGetDNS() {
	tests := []struct {
		name         string
		target       string
		iface        string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DNSConfig]], error)
	}{
		{
			name:   "when requesting DNS returns results",
			target: "_any",
			iface:  "eth0",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSConfig]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("dns-host", resp.Data.Results[0].Hostname)
				suite.Equal([]string{"8.8.8.8"}, resp.Data.Results[0].Servers)
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
						[]byte(`{"results":[{"hostname":"dns-host","servers":["8.8.8.8"]}]}`),
					)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.GetDNS(suite.ctx, tc.target, tc.iface)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestGetDNSError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DNSConfig]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSConfig]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.GetDNS(suite.ctx, "_any", "eth0")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestGetDNSClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DNSConfig]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSConfig]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get dns")
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

			resp, err := sut.Node.GetDNS(suite.ctx, "_any", "eth0")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestGetDNSNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DNSConfig]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSConfig]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.GetDNS(suite.ctx, "_any", "eth0")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUpdateDNS() {
	tests := []struct {
		name          string
		target        string
		iface         string
		servers       []string
		searchDomains []string
		validateFunc  func(*osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], error)
	}{
		{
			name:          "when servers only provided sets servers",
			target:        "_any",
			iface:         "eth0",
			servers:       []string{"8.8.8.8", "8.8.4.4"},
			searchDomains: nil,
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("dns-host", resp.Data.Results[0].Hostname)
				suite.Equal("completed", resp.Data.Results[0].Status)
				suite.True(resp.Data.Results[0].Changed)
			},
		},
		{
			name:          "when search domains only provided sets search domains",
			target:        "_any",
			iface:         "eth0",
			servers:       nil,
			searchDomains: []string{"example.com"},
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name:          "when both provided sets servers and search domains",
			target:        "_any",
			iface:         "eth0",
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{"example.com"},
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
			},
		},
		{
			name:          "when neither provided sends empty body",
			target:        "_any",
			iface:         "eth0",
			servers:       nil,
			searchDomains: nil,
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], err error) {
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
					w.WriteHeader(http.StatusAccepted)
					_, _ = w.Write(
						[]byte(
							`{"results":[{"hostname":"dns-host","status":"completed","changed":true}]}`,
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

			resp, err := sut.Node.UpdateDNS(
				suite.ctx,
				tc.target,
				tc.iface,
				tc.servers,
				tc.searchDomains,
			)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUpdateDNSError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.UpdateDNS(suite.ctx, "_any", "eth0", []string{"8.8.8.8"}, nil)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUpdateDNSClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "update dns")
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

			resp, err := sut.Node.UpdateDNS(suite.ctx, "_any", "eth0", []string{"8.8.8.8"}, nil)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUpdateDNSNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], error)
	}{
		{
			name: "when server returns 202 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.DNSUpdateResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusAccepted, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusAccepted)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.UpdateDNS(suite.ctx, "_any", "eth0", []string{"8.8.8.8"}, nil)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestPing() {
	tests := []struct {
		name         string
		target       string
		address      string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.PingResult]], error)
	}{
		{
			name:    "when pinging address returns results",
			target:  "_any",
			address: "8.8.8.8",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.PingResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("ping-host", resp.Data.Results[0].Hostname)
				suite.Equal(4, resp.Data.Results[0].PacketsSent)
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
							`{"results":[{"hostname":"ping-host","packets_sent":4,"packets_received":4,"packet_loss":0.0}]}`,
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

			resp, err := sut.Node.Ping(suite.ctx, tc.target, tc.address)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestPingError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.PingResult]], error)
	}{
		{
			name: "when server returns 403 returns AuthError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.PingResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error":"forbidden"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Ping(suite.ctx, "_any", "8.8.8.8")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestPingClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.PingResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.PingResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "ping")
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

			resp, err := sut.Node.Ping(suite.ctx, "_any", "8.8.8.8")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestPingNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.PingResult]], error)
	}{
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.PingResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusOK, target.StatusCode)
				suite.Equal("nil response body", target.Message)
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

			resp, err := sut.Node.Ping(suite.ctx, "_any", "8.8.8.8")
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestExec() {
	tests := []struct {
		name         string
		req          osapi.ExecRequest
		validateFunc func(*osapi.Response[osapi.Collection[osapi.CommandResult]], error)
	}{
		{
			name: "when basic command returns results",
			req: osapi.ExecRequest{
				Command: "whoami",
				Target:  "_any",
			},
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("exec-host", resp.Data.Results[0].Hostname)
				suite.Equal("root\n", resp.Data.Results[0].Stdout)
				suite.Equal(0, resp.Data.Results[0].ExitCode)
			},
		},
		{
			name: "when all options provided returns results",
			req: osapi.ExecRequest{
				Command: "ls",
				Args:    []string{"-la", "/tmp"},
				Cwd:     "/tmp",
				Timeout: 10,
				Target:  "_any",
			},
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
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
					w.WriteHeader(http.StatusAccepted)
					_, _ = w.Write(
						[]byte(
							`{"results":[{"hostname":"exec-host","stdout":"root\n","exit_code":0,"changed":true}]}`,
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

			resp, err := sut.Node.Exec(suite.ctx, tc.req)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestExecClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.CommandResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "exec command")
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

			resp, err := sut.Node.Exec(suite.ctx, osapi.ExecRequest{
				Command: "whoami",
				Target:  "_any",
			})
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestExecNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.CommandResult]], error)
	}{
		{
			name: "when server returns 202 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusAccepted, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusAccepted)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Exec(suite.ctx, osapi.ExecRequest{
				Command: "whoami",
				Target:  "_any",
			})
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestExecError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.CommandResult]], error)
	}{
		{
			name: "when server returns 400 returns ValidationError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
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
					_, _ = w.Write([]byte(`{"error":"command is required"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Exec(suite.ctx, osapi.ExecRequest{
				Target: "_any",
			})
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestShell() {
	tests := []struct {
		name         string
		req          osapi.ShellRequest
		validateFunc func(*osapi.Response[osapi.Collection[osapi.CommandResult]], error)
	}{
		{
			name: "when basic command returns results",
			req: osapi.ShellRequest{
				Command: "uname -a",
				Target:  "_any",
			},
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Results, 1)
				suite.Equal("shell-host", resp.Data.Results[0].Hostname)
			},
		},
		{
			name: "when cwd and timeout provided returns results",
			req: osapi.ShellRequest{
				Command: "ls -la",
				Cwd:     "/var/log",
				Timeout: 15,
				Target:  "_any",
			},
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
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
					w.WriteHeader(http.StatusAccepted)
					_, _ = w.Write(
						[]byte(
							`{"results":[{"hostname":"shell-host","exit_code":0,"changed":false}]}`,
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

			resp, err := sut.Node.Shell(suite.ctx, tc.req)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestShellClientError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.CommandResult]], error)
	}{
		{
			name: "when client HTTP call fails returns error",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "shell command")
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

			resp, err := sut.Node.Shell(suite.ctx, osapi.ShellRequest{
				Command: "uname -a",
				Target:  "_any",
			})
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestShellNilResponse() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.CommandResult]], error)
	}{
		{
			name: "when server returns 202 with no JSON body returns UnexpectedStatusError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusAccepted, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusAccepted)
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Shell(suite.ctx, osapi.ShellRequest{
				Command: "uname -a",
				Target:  "_any",
			})
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *NodePublicTestSuite) TestShellError() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Response[osapi.Collection[osapi.CommandResult]], error)
	}{
		{
			name: "when server returns 400 returns ValidationError",
			validateFunc: func(resp *osapi.Response[osapi.Collection[osapi.CommandResult]], err error) {
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
					_, _ = w.Write([]byte(`{"error":"command is required"}`))
				}),
			)
			defer server.Close()

			sut := osapi.New(
				server.URL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.Node.Shell(suite.ctx, osapi.ShellRequest{
				Target: "_any",
			})
			tc.validateFunc(resp, err)
		})
	}
}

func TestNodePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodePublicTestSuite))
}
