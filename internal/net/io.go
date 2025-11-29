//	  \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//	\\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"io"
	"net/http"
)

// body reads and returns all bytes from an HTTP response body. This is a
// helper function that wraps io.ReadAll for use with HTTP responses.
//
// Parameters:
//   - r: The HTTP response containing the body to read
//
// Returns:
//   - []byte: The complete response body as a byte slice
//   - error: Any error encountered while reading the body
//
// Note: This function does not close the response body. The caller is
// responsible for closing r.Body after calling this function.
func body(r *http.Response) ([]byte, error) {
	data, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		return nil, readErr
	}

	return data, nil
}
