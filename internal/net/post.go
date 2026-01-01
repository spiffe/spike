//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"bytes"
	"io"
	"net/http"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
)

// Post performs an HTTP POST request with a JSON payload and returns the
// response body. It handles the common cases of connection errors, non-200
// status codes, and proper response body handling.
//
// Parameters:
//   - client: An *http.Client used to make the request, typically
//     configured with TLS settings.
//   - path: The URL path to send the POST request to.
//   - mr: A byte slice containing the marshaled JSON request body.
//
// Returns:
//   - []byte: The response body if the request is successful.
//   - *sdkErrors.SDKError: An error if any of the following occur:
//   - sdkErrors.ErrAPIPostFailed if request creation fails
//   - sdkErrors.ErrNetPeerConnection if connection fails or non-success
//     status
//   - sdkErrors.ErrAPINotFound if status is 404
//   - sdkErrors.ErrAccessUnauthorized if status is 401
//   - sdkErrors.ErrNetReadingResponseBody if reading response fails
//
// The function ensures proper cleanup by always attempting to close the
// response body via a deferred function. Close errors are logged but not
// returned to the caller.
//
// Example:
//
//	client := &http.Client{}
//	data := []byte(`{"key": "value"}`)
//	response, err := Post(client, "https://api.example.com/endpoint", data)
//	if err != nil {
//	    log.Fatalf("failed to post: %v", err)
//	}
func Post(
	client *http.Client, path string, mr []byte,
) ([]byte, *sdkErrors.SDKError) {
	const fName = "Post"

	// Create the request while preserving the mTLS client
	req, reqErr := http.NewRequest("POST", path, bytes.NewBuffer(mr))
	if reqErr != nil {
		failErr := sdkErrors.ErrAPIPostFailed.Wrap(reqErr)
		failErr.Msg = "failed to create request"
		return nil, failErr
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Use the existing mTLS client to make the request
	r, doErr := client.Do(req)
	if doErr != nil {
		failErr := sdkErrors.ErrNetPeerConnection.Wrap(doErr)
		return nil, failErr
	}

	// Ensure the response body is always closed to prevent resource leaks
	defer func(b io.ReadCloser) {
		if b == nil {
			return
		}
		if closeErr := b.Close(); closeErr != nil {
			failErr := sdkErrors.ErrFSStreamCloseFailed.Wrap(closeErr)
			failErr.Msg = "failed to close response body"
			log.WarnErr(fName, *failErr)
		}
	}(r.Body)

	if r.StatusCode != http.StatusOK {
		if r.StatusCode == http.StatusNotFound {
			return nil, sdkErrors.ErrAPINotFound.Clone()
		}

		if r.StatusCode == http.StatusUnauthorized {
			return nil, sdkErrors.ErrAccessUnauthorized.Clone()
		}

		return nil, sdkErrors.ErrNetPeerConnection.Clone()
	}

	b, bodyErr := body(r)
	if bodyErr != nil {
		failErr := sdkErrors.ErrNetReadingResponseBody.Wrap(bodyErr)
		return nil, failErr
	}

	return b, nil
}
