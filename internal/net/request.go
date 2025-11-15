//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	apiErr "github.com/spiffe/spike-sdk-go/api/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/net"
)

// ReadRequestBody reads the entire request body from an HTTP request.
// It returns the body as a byte slice if successful. If there is an error
// reading the body or if the body is nil, it writes a 400 Bad Request status
// to the response writer and returns an empty byte slice. Any errors
// encountered are logged.
func ReadRequestBody(w http.ResponseWriter, r *http.Request) []byte {
	body, err := net.RequestBody(r)
	if err != nil {
		log.Log().Info("readRequestBody",
			"message", "Problem reading request body",
			"err", err.Error())

		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Info("readRequestBody",
				"message", "Problem writing response",
				"err", err.Error())
		}

		return []byte{}
	}

	if body == nil {
		log.Log().Info("readRequestBody", "message", "No request body.")

		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, "")
		if err != nil {
			log.Log().Info("readRequestBody",
				"message", "Problem writing response",
				"err", err.Error())
		}
		return []byte{}
	}

	return body
}

// HandleRequestError handles HTTP request errors by writing a 400 Bad Request
// status to the response writer. If err is nil, it returns nil. Otherwise, it
// writes the error status and returns a joined error containing both the
// original error and any error encountered while writing the response.
func HandleRequestError(w http.ResponseWriter, err error) error {
	if err == nil {
		return nil
	}

	w.WriteHeader(http.StatusBadRequest)
	_, writeErr := io.WriteString(w, "")

	return errors.Join(err, writeErr)
}

// HandleRequest unmarshals a JSON request body into a typed request struct.
//
// This is a generic function that handles the common pattern of unmarshaling
// and validating incoming JSON requests. If unmarshaling fails, it sends the
// provided error response to the client with a 400 Bad Request status.
//
// Type Parameters:
//   - Req: The request type to unmarshal into
//   - Res: The response type for error cases
//
// Parameters:
//   - requestBody: []byte - The raw JSON request body to unmarshal
//   - w: http.ResponseWriter - The response writer for error handling
//   - errorResponse: Res - A response object to send if unmarshaling fails
//
// Returns:
//   - *Req - A pointer to the unmarshaled request struct, or nil if
//     unmarshaling failed
//
// The function handles all error logging and response writing for the error
// case. Callers should check if the returned pointer is nil before proceeding.
func HandleRequest[Req any, Res any](
	requestBody []byte,
	w http.ResponseWriter,
	errorResponse Res,
) *Req {
	var request Req

	if err := HandleRequestError(
		w, json.Unmarshal(requestBody, &request),
	); err != nil {
		log.Log().Error("HandleRequest",
			"message", "Problem unmarshalling request",
			"err", err.Error())

		responseBody, err := MarshalBodyAndRespondOnMarshalFail(errorResponse, w)
		if err == nil {
			Respond(http.StatusBadRequest, responseBody, w)
		}

		return nil
	}

	return &request
}

// ReadAndParseRequest reads the HTTP request body and parses it into a typed
// request struct in a single operation. This function combines ReadRequestBody
// and HandleRequest to reduce boilerplate in route handlers.
//
// This function performs the following steps:
//  1. Reads the request body from the HTTP request
//  2. Returns ErrReadFailure if reading fails
//  3. Unmarshals the body into the request type
//  4. Returns ErrParseFailure if unmarshaling fails
//  5. Returns the parsed request and nil error on success
//
// Type Parameters:
//   - Req: The request type to unmarshal into
//   - Res: The response type for error cases
//
// Parameters:
//   - w: http.ResponseWriter - The response writer for error handling
//   - r: *http.Request - The incoming HTTP request
//   - errorResponse: Res - A response object to send if parsing fails
//   - logContext: string - Optional context string for logging (e.g., function
//     name). If empty, no additional logging is performed beyond the default.
//
// Returns:
//   - *Req - A pointer to the parsed request struct, or nil if parsing failed
//   - error - apiErr.ErrReadFailure, apiErr.ErrParseFailure, or nil
//
// Example usage:
//
//	request, err := net.ReadAndParseRequest[
//	    reqres.SecretDeleteRequest,
//	    reqres.SecretDeleteResponse](
//	    w, r,
//	    reqres.SecretDeleteResponse{Err: data.ErrBadInput},
//	    "RouteDeleteSecret",
//	)
//	if err != nil {
//	    return err
//	}
func ReadAndParseRequest[Req any, Res any](
	w http.ResponseWriter,
	r *http.Request,
	errorResponse Res,
	logContext string,
) (*Req, error) {
	requestBody := ReadRequestBody(w, r)
	if requestBody == nil {
		if logContext != "" {
			log.Log().Error(logContext, "message", "failed to read request body")
		}
		return nil, apiErr.ErrReadFailure
	}

	request := HandleRequest[Req, Res](requestBody, w, errorResponse)
	if request == nil {
		if logContext != "" {
			log.Log().Error(logContext, "message", "failed to parse request body")
		}
		return nil, apiErr.ErrParseFailure
	}

	return request, nil
}

// GuardFunc is a function type for request guard/validation functions.
// Guard functions validate requests and return an error if validation fails.
// They typically check authentication, authorization, and input validation.
//
// Type Parameters:
//   - Req: The request type to validate
//
// Parameters:
//   - request: The request to validate
//   - w: http.ResponseWriter for writing error responses
//   - r: *http.Request for accessing request context
//
// Returns:
//   - error: nil if validation passes, error otherwise
type GuardFunc[Req any] func(Req, http.ResponseWriter, *http.Request) error

// ReadParseAndGuard reads the HTTP request body, parses it, and executes
// a guard function in a single operation. This function combines
// ReadAndParseRequest with guard execution to further reduce boilerplate.
//
// This function performs the following steps:
//  1. Reads the request body from the HTTP request
//  2. Unmarshals the body into the request type
//  3. Executes the guard function for validation
//  4. Logs errors if logContext is provided
//  5. Returns the parsed request and any errors
//
// Type Parameters:
//   - Req: The request type to unmarshal into
//   - Res: The response type for error cases
//
// Parameters:
//   - w: http.ResponseWriter - The response writer for error handling
//   - r: *http.Request - The incoming HTTP request
//   - errorResponse: Res - A response object to send if parsing fails
//   - guard: GuardFunc[Req] - The guard function to execute for validation
//   - logContext: string - Optional context string for logging (e.g., function
//     name). If empty, no additional logging is performed beyond the default.
//
// Returns:
//   - *Req - A pointer to the parsed request struct, or nil if any step failed
//   - error - apiErr.ErrReadFailure, apiErr.ErrParseFailure, or error from
//     guard function
//
// Example usage:
//
//	request, err := net.ReadParseAndGuard[
//	    reqres.ShardPutRequest,
//	    reqres.ShardPutResponse](
//	    w, r,
//	    reqres.ShardPutResponse{Err: data.ErrBadInput},
//	    guardShardPutRequest,
//	    "RouteContribute",
//	)
//	if err != nil {
//	    return err
//	}
func ReadParseAndGuard[Req any, Res any](
	w http.ResponseWriter,
	r *http.Request,
	errorResponse Res,
	guard GuardFunc[Req],
	logContext string,
) (*Req, error) {
	request, err := ReadAndParseRequest[Req, Res](
		w, r, errorResponse, logContext,
	)
	if err != nil {
		return nil, err
	}

	err = guard(*request, w, r)
	if err != nil {
		if logContext != "" {
			log.Log().Error(logContext, "message", "guard trap", "err", err.Error())
		}
		return nil, err
	}

	return request, nil
}

// FailIfError is a helper function that fails a request if an error occurred
// during validation or processing.
//
// This function provides a reusable pattern for validating inputs and
// responding with appropriate error messages. If the internal error is nil,
// the function returns nil immediately. Otherwise, it marshals the client
// response and sends it with a 400 Bad Request status, then returns the
// specified error to the caller.
//
// Type Parameters:
//   - T: The response type to send to the client (e.g.,
//     reqres.PolicyCreateBadInput)
//
// Parameters:
//   - internalError: The error to check (e.g., from validation functions).
//     If nil, the function returns nil immediately.
//   - errorToRespond: The error to return to the caller (e.g.,
//     apiErr.ErrInvalidInput)
//   - clientResponse: The response object to send to the client if there is
//     an error
//   - w: The HTTP response writer for error responses
//
// Returns:
//   - error: Returns errorToRespond if internalError is not nil, otherwise
//     returns nil
//
// Example usage:
//
//	err := validation.ValidateName(name)
//	if err := net.FailIfError(
//	    err, apiErr.ErrInvalidInput,
//	    reqres.PolicyCreateBadInput, w); err != nil {
//	    return err
//	}
func FailIfError[T any](
	internalError error, errorToRespond error,
	clientResponse T, w http.ResponseWriter,
) error {
	if internalError != nil {
		responseBody, marshalErr := MarshalBodyAndRespondOnMarshalFail(
			clientResponse, w,
		)
		if alreadyResponded := marshalErr != nil; !alreadyResponded {
			Respond(http.StatusBadRequest, responseBody, w)
		}
		return errorToRespond
	}
	return nil
}

// FailIf is a helper function that conditionally fails a request by sending
// an error response based on a boolean condition.
//
// This function provides a reusable pattern for conditional error responses,
// such as authorization checks or validation conditions. If the condition is
// true, it marshals the client response and sends it with the specified HTTP
// status code, then returns the specified error to the caller.
//
// Type Parameters:
//   - T: The response type to send to the client (e.g.,
//     reqres.ShardPutUnauthorized)
//
// Parameters:
//   - condition: If true, fail the request with an error response
//   - clientResponse: The response object to send to the client if condition
//     is true
//   - w: The HTTP response writer for error responses
//   - statusCode: The HTTP status code to send (e.g., http.StatusUnauthorized)
//   - errorToRespond: The error to return to the caller (e.g.,
//     apiErr.ErrUnauthorized)
//
// Returns:
//   - error: Returns errorToRespond if condition is true, otherwise returns
//     nil
//
// Example usage:
//
//	if err := net.FailIf(
//	    !spiffeid.PeerCanTalkToKeeper(peerSPIFFEID.String()),
//	    reqres.ShardPutUnauthorized, w,
//	    http.StatusUnauthorized, apiErr.ErrUnauthorized,
//	); err != nil {
//	    return err
//	}
func FailIf[T any](
	condition bool,
	clientResponse T,
	w http.ResponseWriter,
	statusCode int,
	errorToRespond error,
) error {
	if condition {
		responseBody, marshalErr := MarshalBodyAndRespondOnMarshalFail(
			clientResponse, w,
		)
		if alreadyResponded := marshalErr != nil; !alreadyResponded {
			Respond(statusCode, responseBody, w)
		}
		return errorToRespond
	}
	return nil
}
