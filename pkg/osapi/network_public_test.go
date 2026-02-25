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

type NetworkPublicTestSuite struct {
	suite.Suite

	ctx    context.Context
	server *httptest.Server
	sut    *osapi.Client
}

func (suite *NetworkPublicTestSuite) SetupTest() {
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

func (suite *NetworkPublicTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *NetworkPublicTestSuite) TestGetDNS() {
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
			_, err := suite.sut.Network.GetDNS(suite.ctx, tc.target, tc.iface)
			tc.validateFunc(err)
		})
	}
}

func (suite *NetworkPublicTestSuite) TestUpdateDNS() {
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
			_, err := suite.sut.Network.UpdateDNS(
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

func (suite *NetworkPublicTestSuite) TestPing() {
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
			_, err := suite.sut.Network.Ping(suite.ctx, tc.target, tc.address)
			tc.validateFunc(err)
		})
	}
}

func TestNetworkPublicTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkPublicTestSuite))
}
