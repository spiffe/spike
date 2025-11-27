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
		failErr := *sdkErrors.ErrAPIInternal.Clone()
		failErr.Msg = "problem generating response"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		internalErrJSON, marshalErr := json.Marshal(failErr)

		// Add extra info "after" marshaling to avoid leaking internal error details
		wrappedErr := failErr.Wrap(err)

		if marshalErr != nil {
			wrappedErr = wrappedErr.Wrap(marshalErr)
			// Cannot marshal; try a generic message instead.
			internalErrJSON = []byte(`{"error":"internal server error"}`)
		}
		_, err = w.Write(internalErrJSON)
		if err != nil {
			wrappedErr = wrappedErr.Wrap(err)
			// At this point, we cannot respond. So there is not much to send.
			// We cannot even send a generic error message.
			// We can only log the error.
		}

		// Log the chained error.
		log.ErrorErr(fName, *wrappedErr)
		return nil, wrappedErr
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

// ErrorResponder defines an interface for response types that can generate
// standard error responses. All SDK response types in the reqres package
// implement this interface through their NotFound() and Internal() methods.
type ErrorResponder[T any] interface {
	NotFound() T
	Internal() T
}

// HandleError processes errors from state operations (database, storage) and
// sends appropriate HTTP responses. It uses generics to work with any response
// type that implements the ErrorResponder interface.
//
// Use this function in route handlers after state operations (Get, Put, Delete,
// List, etc.) that may return "not found" or internal errors. Do NOT use this
// for authentication/authorization or input validation errors in guard/intercept
// functions; those have different semantics (400 Bad Request, 401 Unauthorized)
// that don't map to the 404/500 distinction this function provides, so they
// should use net.Fail directly.
//
// The function distinguishes between two types of errors:
//   - sdkErrors.ErrEntityNotFound: Returns HTTP 404 Not Found when the
//     requested resource does not exist
//   - Other errors: Returns HTTP 500 Internal Server Error for backend or
//     server-side failures
//
// Parameters:
//   - err: The error that occurred during the state operation
//   - w: The HTTP response writer for sending error responses
//   - response: A zero-value response instance used to generate error responses
//
// Returns:
//   - *sdkErrors.SDKError: The error that was passed in (for chaining),
//     or nil if err was nil
//
// Example usage:
//
//	// In a route handler after a state operation:
//	if err != nil {
//	    return net.HandleError(err, w, reqres.SecretGetResponse{})
//	}
//
//	// In guard/intercept functions, use net.Fail directly instead:
//	if !authorized {
//	    net.Fail(response.Unauthorized(), w, http.StatusUnauthorized)
//	    return sdkErrors.ErrAccessUnauthorized
//	}
func HandleError[T ErrorResponder[T]](
	err *sdkErrors.SDKError, w http.ResponseWriter, response T,
) *sdkErrors.SDKError {
	if err == nil {
		return nil
	}
	if err.Is(sdkErrors.ErrEntityNotFound) {
		Fail(response.NotFound(), w, http.StatusNotFound)
		return err
	}
	// Backend or other server-side failure
	Fail(response.Internal(), w, http.StatusInternalServerError)
	return err
}

// InternalErrorResponder defines an interface for response types that can
// generate internal error responses. This is a subset of ErrorResponder for
// cases where only internal errors are possible (no "not found" scenario).
type InternalErrorResponder[T any] interface {
	Internal() T
}

// HandleInternalError sends an HTTP 500 Internal Server Error response and
// returns the provided SDK error. Use this for operations where the only
// possible error is an internal/server error (no "not found" case), such as
// cryptographic operations, Shamir secret sharing validation, or system
// initialization checks.
//
// Like HandleError, this is intended for route handlers after state or system
// operations. Do NOT use this for authentication/authorization or input
// validation errors in guard/intercept functions; those have different semantics
// (400 Bad Request, 401 Unauthorized) that this function doesn't handle, so they
// should use net.Fail directly.
//
// Parameters:
//   - err: The SDK error that occurred
//   - w: The HTTP response writer for sending error responses
//   - response: A zero-value response instance used to generate the error
//
// Returns:
//   - *sdkErrors.SDKError: The error that was passed in
//
// Example usage:
//
//	if cipher == nil {
//	    return net.HandleInternalError(
//	        sdkErrors.ErrCryptoCipherNotAvailable, w,
//	        reqres.BootstrapVerifyResponse{},
//	    )
//	}
func HandleInternalError[T InternalErrorResponder[T]](
	err *sdkErrors.SDKError, w http.ResponseWriter, response T,
) *sdkErrors.SDKError {
	Fail(response.Internal(), w, http.StatusInternalServerError)
	return err
}
