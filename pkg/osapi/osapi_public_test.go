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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/osapi"
)

type ClientPublicTestSuite struct {
	suite.Suite

	server *httptest.Server
}

func (suite *ClientPublicTestSuite) SetupTest() {
	suite.server = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		}),
	)
}

func (suite *ClientPublicTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *ClientPublicTestSuite) TestNew() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Client, error)
	}{
		{
			name: "when creating client returns all services",
			validateFunc: func(c *osapi.Client, err error) {
				suite.NoError(err)
				suite.NotNil(c)
				suite.NotNil(c.System)
				suite.NotNil(c.Network)
				suite.NotNil(c.Command)
				suite.NotNil(c.Job)
				suite.NotNil(c.Health)
				suite.NotNil(c.Audit)
				suite.NotNil(c.Metrics)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			c, err := osapi.New(suite.server.URL, "test-token")
			tc.validateFunc(c, err)
		})
	}
}

func (suite *ClientPublicTestSuite) TestNewWithHTTPTransport() {
	tests := []struct {
		name         string
		validateFunc func(*osapi.Client, error)
	}{
		{
			name: "when custom transport provided creates client",
			validateFunc: func(c *osapi.Client, err error) {
				suite.NoError(err)
				suite.NotNil(c)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			customTransport := &http.Transport{}
			c, err := osapi.New(
				suite.server.URL,
				"test-token",
				osapi.WithHTTPTransport(customTransport),
				osapi.WithLogger(slog.Default()),
			)
			tc.validateFunc(c, err)
		})
	}
}

func TestClientPublicTestSuite(t *testing.T) {
	suite.Run(t, new(ClientPublicTestSuite))
}
