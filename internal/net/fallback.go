//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/net"
)

// respondFallbackWithStatus writes a fallback JSON response with the given HTTP
// status code and error code. It sets appropriate headers to prevent caching.
//
// This function is used when the primary response handling fails or when a
// generic error response needs to be sent.
//
// Parameters:
//   - w: The HTTP response writer
//   - status: The HTTP status code to return
//   - code: The error code to include in the response body
//
// Returns:
//   - *sdkErrors.SDKError: An error if marshaling or writing fails,
//     nil on success
func respondFallbackWithStatus(
	w http.ResponseWriter, status int, code sdkErrors.ErrorCode,
) *sdkErrors.SDKError {
	body, err := net.MarshalBodyAndRespondOnMarshalFail(
		reqres.FallbackResponse{Err: code}, w,
	)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")

	// Add cache invalidation headers
	w.Header().Set(
		"Cache-Control",
		"no-store, no-cache, must-revalidate, private",
	)
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		failErr := sdkErrors.ErrAPIInternal.Wrap(err)
		return failErr
	}

	return nil
}
