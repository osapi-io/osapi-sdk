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
	"context"

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
)

// NodeService provides node management operations.
type NodeService struct {
	client *gen.ClientWithResponses
}

// Hostname retrieves the hostname from the target host.
func (s *NodeService) Hostname(
	ctx context.Context,
	target string,
) (*gen.GetNodeHostnameResponse, error) {
	params := &gen.GetNodeHostnameParams{
		TargetHostname: &target,
	}

	return s.client.GetNodeHostnameWithResponse(ctx, params)
}

// Status retrieves node status (OS info, disk, memory, load) from the
// target host.
func (s *NodeService) Status(
	ctx context.Context,
	target string,
) (*gen.GetNodeStatusResponse, error) {
	params := &gen.GetNodeStatusParams{
		TargetHostname: &target,
	}

	return s.client.GetNodeStatusWithResponse(ctx, params)
}

// Agents retrieves active agents.
func (s *NodeService) Agents(
	ctx context.Context,
) (*gen.GetNodeResponse, error) {
	return s.client.GetNodeWithResponse(ctx)
}
