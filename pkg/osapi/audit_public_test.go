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

type AuditPublicTestSuite struct {
	suite.Suite

	ctx    context.Context
	server *httptest.Server
	sut    *osapi.Client
}

func (suite *AuditPublicTestSuite) SetupTest() {
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

func (suite *AuditPublicTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *AuditPublicTestSuite) TestList() {
	tests := []struct {
		name         string
		limit        int
		offset       int
		validateFunc func(error)
	}{
		{
			name:   "when listing audit entries returns no error",
			limit:  20,
			offset: 0,
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Audit.List(suite.ctx, tc.limit, tc.offset)
			tc.validateFunc(err)
		})
	}
}

func (suite *AuditPublicTestSuite) TestGet() {
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
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Audit.Get(suite.ctx, tc.id)
			tc.validateFunc(err)
		})
	}
}

func (suite *AuditPublicTestSuite) TestExport() {
	tests := []struct {
		name         string
		validateFunc func(error)
	}{
		{
			name: "when exporting audit entries returns no error",
			validateFunc: func(err error) {
				suite.NoError(err)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			_, err := suite.sut.Audit.Export(suite.ctx)
			tc.validateFunc(err)
		})
	}
}

func TestAuditPublicTestSuite(t *testing.T) {
	suite.Run(t, new(AuditPublicTestSuite))
}
