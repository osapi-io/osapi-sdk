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

type CommandPublicTestSuite struct {
	suite.Suite

	ctx    context.Context
	server *httptest.Server
	sut    *osapi.Client
}

func (suite *CommandPublicTestSuite) SetupTest() {
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

func (suite *CommandPublicTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *CommandPublicTestSuite) TestExec() {
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
			_, err := suite.sut.Command.Exec(suite.ctx, tc.req)
			tc.validateFunc(err)
		})
	}
}

func (suite *CommandPublicTestSuite) TestShell() {
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
			_, err := suite.sut.Command.Shell(suite.ctx, tc.req)
			tc.validateFunc(err)
		})
	}
}

func TestCommandPublicTestSuite(t *testing.T) {
	suite.Run(t, new(CommandPublicTestSuite))
}
