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

// NetworkService provides network management operations.
type NetworkService struct {
	client *gen.ClientWithResponses
}

// GetDNS retrieves DNS configuration for a network interface on the
// target host.
func (s *NetworkService) GetDNS(
	ctx context.Context,
	target string,
	interfaceName string,
) (*gen.GetNetworkDNSByInterfaceResponse, error) {
	params := &gen.GetNetworkDNSByInterfaceParams{
		TargetHostname: &target,
	}

	return s.client.GetNetworkDNSByInterfaceWithResponse(ctx, interfaceName, params)
}

// UpdateDNS updates DNS configuration for a network interface on the
// target host.
func (s *NetworkService) UpdateDNS(
	ctx context.Context,
	target string,
	interfaceName string,
	servers []string,
	searchDomains []string,
) (*gen.PutNetworkDNSResponse, error) {
	params := &gen.PutNetworkDNSParams{
		TargetHostname: &target,
	}

	body := gen.DNSConfigUpdateRequest{
		InterfaceName: interfaceName,
	}

	if len(servers) > 0 {
		body.Servers = &servers
	}

	if len(searchDomains) > 0 {
		body.SearchDomains = &searchDomains
	}

	return s.client.PutNetworkDNSWithResponse(ctx, params, body)
}

// Ping sends an ICMP ping to the specified address from the target host.
func (s *NetworkService) Ping(
	ctx context.Context,
	target string,
	address string,
) (*gen.PostNetworkPingResponse, error) {
	params := &gen.PostNetworkPingParams{
		TargetHostname: &target,
	}

	body := gen.PostNetworkPingJSONRequestBody{
		Address: address,
	}

	return s.client.PostNetworkPingWithResponse(ctx, params, body)
}
