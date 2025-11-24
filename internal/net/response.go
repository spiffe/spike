//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"net/http"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/internal/journal"
)

// MarshalBodyAndRespondOnMarshalFail serializes a response object to JSON and
// handles error cases.
//
// This function attempts to marshal the provided response object to JSON bytes.
// If marshaling fails, it sends a 500 Internal Server Error response to the
// client and returns nil. The function handles all error logging and response
// writing for the error case.
//
// Parameters:
//   - res: any - The response object to marshal to JSON
//   - w: http.ResponseWriter - The response writer for error handling
//
// Returns:
//   - []byte: The marshaled JSON bytes, or nil if marshaling failed
//   - *sdkErrors.SDKError: sdkErrors.ErrAPIInternal if marshaling failed,
//     nil otherwise
func MarshalBodyAndRespondOnMarshalFail(
	res any, w http.ResponseWriter,
) ([]byte, *sdkErrors.SDKError) {
	const fName = "MarshalBodyAndRespondOnMarshalFail"

	body, err := json.Marshal(res)
	// Since this function is typically called with sentinel error values,
	// this error should, typically, never happen.
	// That's why, instead of sending a "marshal failure" sentinel error,
	// we return an internal sentinel error (sdkErrors.ErrAPIInternal)
	if err != nil {
		// Chain an error for detailed internal logging.
		failErr := sdkErrors.ErrAPIInternal
		failErr.Msg = "problem generating response"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		internalErrJson, marshalErr := json.Marshal(failErr)

		// Add extra info "after" marshaling to avoid leaking internal error details
		failErr = failErr.Wrap(err)

		if marshalErr != nil {
			failErr = failErr.Wrap(marshalErr)
			// Cannot marshal; try a generic message instead.
			internalErrJson = []byte(`{"error":"internal server error"}`)
		}
		_, err = w.Write(internalErrJson)
		if err != nil {
			failErr = failErr.Wrap(err)
			// At this point, we cannot respond. So there is not much to send.
			// We cannot even send a generic error message.
			// We can only log the error.
		}

		// Log the chained error.
		log.ErrorErr(fName, *failErr)
		return nil, failErr
	}

	// body marshaled successfully
	return body, nil
}

// Respond writes a JSON response with the specified status code and body.
//
// This function sets the Content-Type header to application/json, adds cache
// invalidation headers (Cache-Control, Pragma, Expires), writes the provided
// status code, and sends the response body. Any errors during writing are
// logged but not returned to the caller.
//
// Parameters:
//   - statusCode: int - The HTTP status code to send
//   - body: []byte - The pre-marshaled JSON response body
//   - w: http.ResponseWriter - The response writer to use
func Respond(statusCode int, body []byte, w http.ResponseWriter) {
	const fName = "Respond"

	w.Header().Set("Content-Type", "application/json")

	// Add cache invalidation headers
	w.Header().Set(
		"Cache-Control",
		"no-store, no-cache, must-revalidate, private",
	)
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(statusCode)

	_, err := w.Write(body)
	if err != nil {
		// At this point, we cannot respond. So there is not much to send
		// back to the client. We can only log the error.
		// This should rarely, if ever, happen.
		failErr := sdkErrors.ErrAPIInternal.Wrap(err)
		log.ErrorErr(fName, *failErr)
	}
}

// Fallback handles requests to undefined routes by returning a 400 Bad Request.
//
// This function serves as a catch-all handler for undefined routes, logging the
// request details and returning a standardized error response. It uses
// MarshalBodyAndRespondOnMarshalFail to generate the response and handles any
// errors during response writing.
//
// Parameters:
//   - w: http.ResponseWriter - The response writer
//   - r: *http.Request - The incoming request
//   - audit: *journal.AuditEntry - The audit log entry for this request
//
// The response always includes:
//   - Status: 400 Bad Request
//   - Content-Type: application/json
//   - Body: JSON object with an error field
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or sdkErrors.ErrAPIInternal if
//     response marshaling or writing fails
func Fallback(
	w http.ResponseWriter, _ *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	audit.Action = journal.AuditFallback

	return respondFallbackWithStatus(
		w, http.StatusBadRequest, sdkErrors.ErrAPIBadRequest.Code,
	)
}

// NotReady handles requests when the system has not initialized its backing
// store with a root key by returning a 503 Service Unavailable.
//
// This function uses MarshalBodyAndRespondOnMarshalFail to generate the
// response and handles any errors during response writing.
//
// Parameters:
//   - w: http.ResponseWriter - The response writer
//   - r: *http.Request - The incoming request
//   - audit: *journal.AuditEntry - The audit log entry for this request
//
// The response always includes:
//   - Status: 503 Service Unavailable
//   - Content-Type: application/json
//   - Body: JSON object with an error field containing ErrStateNotReady
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or sdkErrors.ErrAPIInternal if
//     response marshaling or writing fails
func NotReady(
	w http.ResponseWriter, _ *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError {
	audit.Action = journal.AuditBlocked

	return respondFallbackWithStatus(
		w, http.StatusServiceUnavailable, sdkErrors.ErrStateNotReady.Code,
	)
}
