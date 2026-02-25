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

	"github.com/google/uuid"

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
)

// AuditService provides audit log operations.
type AuditService struct {
	client *gen.ClientWithResponses
}

// List retrieves audit log entries with pagination.
func (s *AuditService) List(
	ctx context.Context,
	limit int,
	offset int,
) (*gen.GetAuditLogsResponse, error) {
	params := &gen.GetAuditLogsParams{
		Limit:  &limit,
		Offset: &offset,
	}

	return s.client.GetAuditLogsWithResponse(ctx, params)
}

// Get retrieves a single audit log entry by ID.
func (s *AuditService) Get(
	ctx context.Context,
	id string,
) (*gen.GetAuditLogByIDResponse, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return s.client.GetAuditLogByIDWithResponse(ctx, parsedID)
}

// Export retrieves all audit log entries for export.
func (s *AuditService) Export(
	ctx context.Context,
) (*gen.GetAuditExportResponse, error) {
	return s.client.GetAuditExportWithResponse(ctx)
}
