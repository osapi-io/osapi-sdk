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

// CommandService provides command execution operations.
type CommandService struct {
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

// Exec executes a command directly without a shell interpreter.
func (s *CommandService) Exec(
	ctx context.Context,
	req ExecRequest,
) (*gen.PostCommandExecResponse, error) {
	params := &gen.PostCommandExecParams{
		TargetHostname: &req.Target,
	}

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

	return s.client.PostCommandExecWithResponse(ctx, params, body)
}

// Shell executes a command through /bin/sh -c with shell features
// (pipes, redirects, variable expansion).
func (s *CommandService) Shell(
	ctx context.Context,
	req ShellRequest,
) (*gen.PostCommandShellResponse, error) {
	params := &gen.PostCommandShellParams{
		TargetHostname: &req.Target,
	}

	body := gen.CommandShellRequest{
		Command: req.Command,
	}

	if req.Cwd != "" {
		body.Cwd = &req.Cwd
	}

	if req.Timeout > 0 {
		body.Timeout = &req.Timeout
	}

	return s.client.PostCommandShellWithResponse(ctx, params, body)
}
