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

package osapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
)

type ClientInternalTestSuite struct {
	suite.Suite
}

func (s *ClientInternalTestSuite) TestNewGenClientError() {
	tests := []struct {
		name         string
		validateFunc func(*Client, error)
	}{
		{
			name: "when gen client creation fails returns error",
			validateFunc: func(c *Client, err error) {
				s.Error(err)
				s.Contains(err.Error(), "gen client error")
				s.Nil(c)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			original := newGenClient
			defer func() { newGenClient = original }()

			newGenClient = func(
				_ string,
				_ ...gen.ClientOption,
			) (*gen.ClientWithResponses, error) {
				return nil, fmt.Errorf("gen client error")
			}

			c, err := New("http://localhost:8080", "token")
			tt.validateFunc(c, err)
		})
	}
}

func TestClientInternalTestSuite(t *testing.T) {
	suite.Run(t, new(ClientInternalTestSuite))
}
