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

// ExecRequest contains parameters for direct command execution.
type ExecRequest struct {
	// Command is the binary to execute (required).
	Command string

	// Args is the argument list passed to the command.
	Args []string

	// Cwd is the working directory. Empty uses the agent default.
	Cwd string

	// Timeout in seconds. Zero uses the server default (30s).
	Timeout int

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// ShellRequest contains parameters for shell command execution.
type ShellRequest struct {
	// Command is the shell command string passed to /bin/sh -c (required).
	Command string

	// Cwd is the working directory. Empty uses the agent default.
	Cwd string

	// Timeout in seconds. Zero uses the server default (30s).
	Timeout int

	// Target specifies the host: "_any", "_all", hostname, or
	// label ("group:web").
	Target string
}

// Status retrieves node status (OS info, disk, memory, load) from the
// target host.
func (s *NodeService) Status(
	ctx context.Context,
	target string,
) (*gen.GetNodeStatusResponse, error) {
	return s.client.GetNodeStatusWithResponse(ctx, target)
}

// Hostname retrieves the hostname from the target host.
func (s *NodeService) Hostname(
	ctx context.Context,
	target string,
) (*gen.GetNodeHostnameResponse, error) {
	return s.client.GetNodeHostnameWithResponse(ctx, target)
}

// Disk retrieves disk usage information from the target host.
func (s *NodeService) Disk(
	ctx context.Context,
	target string,
) (*gen.GetNodeDiskResponse, error) {
	return s.client.GetNodeDiskWithResponse(ctx, target)
}

// Memory retrieves memory usage information from the target host.
func (s *NodeService) Memory(
	ctx context.Context,
	target string,
) (*gen.GetNodeMemoryResponse, error) {
	return s.client.GetNodeMemoryWithResponse(ctx, target)
}

// Load retrieves load average information from the target host.
func (s *NodeService) Load(
	ctx context.Context,
	target string,
) (*gen.GetNodeLoadResponse, error) {
	return s.client.GetNodeLoadWithResponse(ctx, target)
}

// OS retrieves operating system information from the target host.
func (s *NodeService) OS(
	ctx context.Context,
	target string,
) (*gen.GetNodeOSResponse, error) {
	return s.client.GetNodeOSWithResponse(ctx, target)
}

// Uptime retrieves uptime information from the target host.
func (s *NodeService) Uptime(
	ctx context.Context,
	target string,
) (*gen.GetNodeUptimeResponse, error) {
	return s.client.GetNodeUptimeWithResponse(ctx, target)
}

// GetDNS retrieves DNS configuration for a network interface on the
// target host.
func (s *NodeService) GetDNS(
	ctx context.Context,
	target string,
	interfaceName string,
) (*gen.GetNodeNetworkDNSByInterfaceResponse, error) {
	return s.client.GetNodeNetworkDNSByInterfaceWithResponse(ctx, target, interfaceName)
}

// UpdateDNS updates DNS configuration for a network interface on the
// target host.
func (s *NodeService) UpdateDNS(
	ctx context.Context,
	target string,
	interfaceName string,
	servers []string,
	searchDomains []string,
) (*gen.PutNodeNetworkDNSResponse, error) {
	body := gen.DNSConfigUpdateRequest{
		InterfaceName: interfaceName,
	}

	if len(servers) > 0 {
		body.Servers = &servers
	}

	if len(searchDomains) > 0 {
		body.SearchDomains = &searchDomains
	}

	return s.client.PutNodeNetworkDNSWithResponse(ctx, target, body)
}

// Ping sends an ICMP ping to the specified address from the target host.
func (s *NodeService) Ping(
	ctx context.Context,
	target string,
	address string,
) (*gen.PostNodeNetworkPingResponse, error) {
	body := gen.PostNodeNetworkPingJSONRequestBody{
		Address: address,
	}

	return s.client.PostNodeNetworkPingWithResponse(ctx, target, body)
}

// Exec executes a command directly without a shell interpreter.
func (s *NodeService) Exec(
	ctx context.Context,
	req ExecRequest,
) (*gen.PostNodeCommandExecResponse, error) {
	body := gen.CommandExecRequest{
		Command: req.Command,
	}

	if len(req.Args) > 0 {
		body.Args = &req.Args
	}

	if req.Cwd != "" {
		body.Cwd = &req.Cwd
	}

	if req.Timeout > 0 {
		body.Timeout = &req.Timeout
	}

	return s.client.PostNodeCommandExecWithResponse(ctx, req.Target, body)
}

// Shell executes a command through /bin/sh -c with shell features
// (pipes, redirects, variable expansion).
func (s *NodeService) Shell(
	ctx context.Context,
	req ShellRequest,
) (*gen.PostNodeCommandShellResponse, error) {
	body := gen.CommandShellRequest{
		Command: req.Command,
	}

	if req.Cwd != "" {
		body.Cwd = &req.Cwd
	}

	if req.Timeout > 0 {
		body.Timeout = &req.Timeout
	}

	return s.client.PostNodeCommandShellWithResponse(ctx, req.Target, body)
}
