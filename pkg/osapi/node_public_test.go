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

type NodePublicTestSuite struct {
	suite.Suite

	ctx    context.Context
	server *httptest.Server
	sut    *osapi.Client
}

func (suite *NodePublicTestSuite) SetupTest() {
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

func (suite *NodePublicTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *NodePublicTestSuite) TestHostname() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(error)
	}{
		{
			name:   "when requesting hostname returns no error",
			target: "_any",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Hostname(suite.ctx, tc.target)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestStatus() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(error)
	}{
		{
			name:   "when requesting status returns no error",
			target: "_any",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Status(suite.ctx, tc.target)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestDisk() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(error)
	}{
		{
			name:   "when requesting disk returns no error",
			target: "_any",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Disk(suite.ctx, tc.target)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestMemory() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(error)
	}{
		{
			name:   "when requesting memory returns no error",
			target: "_any",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Memory(suite.ctx, tc.target)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestLoad() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(error)
	}{
		{
			name:   "when requesting load returns no error",
			target: "_any",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Load(suite.ctx, tc.target)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestOS() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(error)
	}{
		{
			name:   "when requesting OS info returns no error",
			target: "_any",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.OS(suite.ctx, tc.target)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestUptime() {
	tests := []struct {
		name         string
		target       string
		validateFunc func(error)
	}{
		{
			name:   "when requesting uptime returns no error",
			target: "_any",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Uptime(suite.ctx, tc.target)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestGetDNS() {
	tests := []struct {
		name         string
		target       string
		iface        string
		validateFunc func(error)
	}{
		{
			name:   "when requesting DNS returns no error",
			target: "_any",
			iface:  "eth0",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.GetDNS(suite.ctx, tc.target, tc.iface)
			tc.validateFunc(err)
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
		validateFunc  func(error)
	}{
		{
			name:          "when servers only provided sets servers",
			target:        "_any",
			iface:         "eth0",
			servers:       []string{"8.8.8.8", "8.8.4.4"},
			searchDomains: nil,
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:          "when search domains only provided sets search domains",
			target:        "_any",
			iface:         "eth0",
			servers:       nil,
			searchDomains: []string{"example.com"},
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:          "when both provided sets servers and search domains",
			target:        "_any",
			iface:         "eth0",
			servers:       []string{"8.8.8.8"},
			searchDomains: []string{"example.com"},
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name:          "when neither provided sends empty body",
			target:        "_any",
			iface:         "eth0",
			servers:       nil,
			searchDomains: nil,
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.UpdateDNS(
				suite.ctx,
				tc.target,
				tc.iface,
				tc.servers,
				tc.searchDomains,
			)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestPing() {
	tests := []struct {
		name         string
		target       string
		address      string
		validateFunc func(error)
	}{
		{
			name:    "when pinging address returns no error",
			target:  "_any",
			address: "8.8.8.8",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Ping(suite.ctx, tc.target, tc.address)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestExec() {
	tests := []struct {
		name         string
		req          osapi.ExecRequest
		validateFunc func(error)
	}{
		{
			name: "when basic command returns no error",
			req: osapi.ExecRequest{
				Command: "whoami",
				Target:  "_any",
			},
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name: "when all options provided returns no error",
			req: osapi.ExecRequest{
				Command: "ls",
				Args:    []string{"-la", "/tmp"},
				Cwd:     "/tmp",
				Timeout: 10,
				Target:  "_any",
			},
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Exec(suite.ctx, tc.req)
			tc.validateFunc(err)
		})
	}
}

func (suite *NodePublicTestSuite) TestShell() {
	tests := []struct {
		name         string
		req          osapi.ShellRequest
		validateFunc func(error)
	}{
		{
			name: "when basic command returns no error",
			req: osapi.ShellRequest{
				Command: "uname -a",
				Target:  "_any",
			},
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
		{
			name: "when cwd and timeout provided returns no error",
			req: osapi.ShellRequest{
				Command: "ls -la",
				Cwd:     "/var/log",
				Timeout: 15,
				Target:  "_any",
			},
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Node.Shell(suite.ctx, tc.req)
			tc.validateFunc(err)
		})
	}
}

func TestNodePublicTestSuite(t *testing.T) {
	suite.Run(t, new(NodePublicTestSuite))
}
