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

type FilePublicTestSuite struct {
	suite.Suite

	ctx context.Context
}

func (suite *FilePublicTestSuite) SetupTest() {
	suite.ctx = context.Background()
}

func (suite *FilePublicTestSuite) TestUpload() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		validateFunc func(*osapi.Response[osapi.FileUpload], error)
	}{
		{
			name: "when uploading file returns result",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(
					[]byte(`{"name":"nginx.conf","sha256":"abc123","size":1024}`),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("nginx.conf", resp.Data.Name)
				suite.Equal("abc123", resp.Data.SHA256)
				suite.Equal(1024, resp.Data.Size)
			},
		},
		{
			name: "when server returns 400 returns ValidationError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"name is required"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name: "when server returns 403 returns AuthError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "upload file")
			},
		},
		{
			name: "when server returns 201 with no JSON body returns UnexpectedStatusError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileUpload], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.UnexpectedStatusError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusCreated, target.StatusCode)
				suite.Equal("nil response body", target.Message)
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.File.Upload(suite.ctx, "nginx.conf", []byte("content"))
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *FilePublicTestSuite) TestList() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		validateFunc func(*osapi.Response[osapi.FileList], error)
	}{
		{
			name: "when listing files returns results",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(
						`{"files":[{"name":"file1.txt","sha256":"aaa","size":100}],"total":1}`,
					),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileList], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Len(resp.Data.Files, 1)
				suite.Equal(1, resp.Data.Total)
				suite.Equal("file1.txt", resp.Data.Files[0].Name)
			},
		},
		{
			name: "when server returns 403 returns AuthError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileList], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *osapi.Response[osapi.FileList], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "list files")
			},
		},
		{
			name: "when server returns 200 with no JSON body returns UnexpectedStatusError",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileList], err error) {
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
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.File.List(suite.ctx)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *FilePublicTestSuite) TestGet() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		fileName     string
		validateFunc func(*osapi.Response[osapi.FileMetadata], error)
	}{
		{
			name:     "when getting file returns metadata",
			fileName: "nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(`{"name":"nginx.conf","sha256":"def456","size":512}`),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("nginx.conf", resp.Data.Name)
				suite.Equal("def456", resp.Data.SHA256)
				suite.Equal(512, resp.Data.Size)
			},
		},
		{
			name:     "when server returns 400 returns ValidationError",
			fileName: "nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"invalid file name"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name:     "when server returns 404 returns NotFoundError",
			fileName: "missing.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"file not found"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
			},
		},
		{
			name:     "when server returns 403 returns AuthError",
			fileName: "nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			fileName:  "nginx.conf",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "get file nginx.conf")
			},
		},
		{
			name:     "when server returns 200 with no JSON body returns UnexpectedStatusError",
			fileName: "nginx.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileMetadata], err error) {
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
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.File.Get(suite.ctx, tc.fileName)
			tc.validateFunc(resp, err)
		})
	}
}

func (suite *FilePublicTestSuite) TestDelete() {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		serverURL    string
		fileName     string
		validateFunc func(*osapi.Response[osapi.FileDelete], error)
	}{
		{
			name:     "when deleting file returns result",
			fileName: "old.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(
					[]byte(`{"name":"old.conf","deleted":true}`),
				)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.NoError(err)
				suite.NotNil(resp)
				suite.Equal("old.conf", resp.Data.Name)
				suite.True(resp.Data.Deleted)
			},
		},
		{
			name:     "when server returns 400 returns ValidationError",
			fileName: "old.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"invalid file name"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.ValidationError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusBadRequest, target.StatusCode)
			},
		},
		{
			name:     "when server returns 404 returns NotFoundError",
			fileName: "missing.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"file not found"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.NotFoundError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusNotFound, target.StatusCode)
			},
		},
		{
			name:     "when server returns 403 returns AuthError",
			fileName: "old.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"forbidden"}`))
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)

				var target *osapi.AuthError
				suite.True(errors.As(err, &target))
				suite.Equal(http.StatusForbidden, target.StatusCode)
			},
		},
		{
			name:      "when client HTTP call fails returns error",
			fileName:  "old.conf",
			serverURL: "http://127.0.0.1:0",
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
				suite.Error(err)
				suite.Nil(resp)
				suite.Contains(err.Error(), "delete file old.conf")
			},
		},
		{
			name:     "when server returns 200 with no JSON body returns UnexpectedStatusError",
			fileName: "old.conf",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			validateFunc: func(resp *osapi.Response[osapi.FileDelete], err error) {
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
			var (
				serverURL string
				cleanup   func()
			)

			if tc.serverURL != "" {
				serverURL = tc.serverURL
				cleanup = func() {}
			} else {
				server := httptest.NewServer(tc.handler)
				serverURL = server.URL
				cleanup = server.Close
			}
			defer cleanup()

			sut := osapi.New(
				serverURL,
				"test-token",
				osapi.WithLogger(slog.Default()),
			)

			resp, err := sut.File.Delete(suite.ctx, tc.fileName)
			tc.validateFunc(resp, err)
		})
	}
}

func TestFilePublicTestSuite(t *testing.T) {
	suite.Run(t, new(FilePublicTestSuite))
}
