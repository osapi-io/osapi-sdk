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
	"fmt"

	"github.com/osapi-io/osapi-sdk/pkg/osapi/gen"
)

// FileService provides file management operations for the Object Store.
type FileService struct {
	client *gen.ClientWithResponses
}

// Upload uploads a file to the Object Store.
func (s *FileService) Upload(
	ctx context.Context,
	name string,
	content []byte,
) (*Response[FileUpload], error) {
	body := gen.FileUploadRequest{
		Name:    name,
		Content: content,
	}

	resp, err := s.client.PostFileWithResponse(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON201 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileUploadFromGen(resp.JSON201), resp.Body), nil
}

// List retrieves all files stored in the Object Store.
func (s *FileService) List(
	ctx context.Context,
) (*Response[FileList], error) {
	resp, err := s.client.GetFilesWithResponse(ctx)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON401,
		resp.JSON403,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileListFromGen(resp.JSON200), resp.Body), nil
}

// Get retrieves metadata for a specific file in the Object Store.
func (s *FileService) Get(
	ctx context.Context,
	name string,
) (*Response[FileMetadata], error) {
	resp, err := s.client.GetFileByNameWithResponse(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get file %s: %w", name, err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileMetadataFromGen(resp.JSON200), resp.Body), nil
}

// Delete removes a file from the Object Store.
func (s *FileService) Delete(
	ctx context.Context,
	name string,
) (*Response[FileDelete], error) {
	resp, err := s.client.DeleteFileByNameWithResponse(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("delete file %s: %w", name, err)
	}

	if err := checkError(
		resp.StatusCode(),
		resp.JSON400,
		resp.JSON401,
		resp.JSON403,
		resp.JSON404,
		resp.JSON500,
	); err != nil {
		return nil, err
	}

	if resp.JSON200 == nil {
		return nil, &UnexpectedStatusError{APIError{
			StatusCode: resp.StatusCode(),
			Message:    "nil response body",
		}}
	}

	return NewResponse(fileDeleteFromGen(resp.JSON200), resp.Body), nil
}
