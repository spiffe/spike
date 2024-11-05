//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"bytes"
	"errors"
	"io"
	"net/http"
)

func body(r *http.Response) (bod []byte, err error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}

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
//   - error: An error if any of the following occur:
//   - Connection failure during POST request
//   - Non-200 status code in response
//   - Failure to read response body
//   - Failure to close response body
//
// The function ensures proper cleanup by always attempting to close the
// response body, even if an error occurs during reading. Any error from closing
// the body is joined with any existing error using errors.Join.
//
// Example:
//
//	client := &http.Client{}
//	data := []byte(`{"key": "value"}`)
//	response, err := Post(client, "https://api.example.com/endpoint", data)
//	if err != nil {
//	    log.Fatalf("failed to post: %v", err)
//	}
func Post(client *http.Client, path string, mr []byte) ([]byte, error) {
	r, err := client.Post(path, "application/json", bytes.NewBuffer(mr))

	if err != nil {
		return []byte{}, errors.Join(
			errors.New("post: Problem connecting to peer"),
			err,
		)
	}

	if r.StatusCode != http.StatusOK {
		return []byte{}, errors.New("post: Problem connecting to peer")
	}

	b, err := body(r)
	if err != nil {
		return []byte{}, errors.Join(
			errors.New("post: Problem reading response body"),
			err,
		)
	}

	defer func(b io.ReadCloser) {
		if b == nil {
			return
		}
		err = errors.Join(err, b.Close())
	}(r.Body)

	return b, nil
}
