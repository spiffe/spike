//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"io"
	"net/http"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/net"
)

// ReadRequestBodyAndRespondOnFail reads the entire request body from an HTTP
// request. It returns the body as a byte slice if successful, or an error if
// reading fails. On error, it writes a 400 Bad Request status to the response
// writer and returns the error for propagation to the caller.
func ReadRequestBodyAndRespondOnFail(
	w http.ResponseWriter, r *http.Request,
) ([]byte, *sdkErrors.SDKError) {
	const fName = "ReadRequestBodyAndRespondOnFail"

	body, err := net.RequestBody(r)
	if err != nil {
		failErr := sdkErrors.ErrDataReadFailure.Wrap(err)
		failErr.Msg = "problem reading request body"

		// do not send the wrapped error to the client as it may contain
		// error details that an attacker can use and exploit.
		failJSON, err := json.Marshal(sdkErrors.ErrDataReadFailure)
		if err != nil {
			// Cannot even parse a generic struct, this is an internal error.
			w.WriteHeader(http.StatusInternalServerError)
			_, writeErr := io.WriteString(w, string(failJSON))
			if writeErr != nil {
				// Cannot even write the error response, this is a critical error.
				failErr = failErr.Wrap(writeErr)
				failErr.Msg = "problem writing response"
			}

			log.ErrorErr(fName, *failErr)

			return nil, failErr
		}

		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := io.WriteString(w, string(failJSON))
		if writeErr != nil {
			failErr = failErr.Wrap(writeErr)
			failErr.Msg = "problem writing response"
			// Cannot even write the error response, this is a critical error.
			// We can only log the error at this point.
			log.ErrorErr(fName, *failErr)
			return nil, failErr
		}

		return nil, failErr
	}

	return body, nil
}

// UnmarshalAndRespondOnFail unmarshals a JSON request body into a typed
// request struct.
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
//   - requestBody: The raw JSON request body to unmarshal
//   - w: The response writer for error handling
//   - errorResponseForBadRequest: A response object to send if unmarshaling
//     fails
//
// Returns:
//   - *Req: A pointer to the unmarshaled request struct, or nil if
//     unmarshaling failed
//   - *sdkErrors.SDKError: ErrDataUnmarshalFailure if unmarshaling fails, or
//     nil on success
//
// The function handles all error logging and response writing for the error
// case. Callers should check if the returned pointer is nil before proceeding.
func UnmarshalAndRespondOnFail[Req any, Res any](
	requestBody []byte,
	w http.ResponseWriter,
	errorResponseForBadRequest Res,
) (*Req, *sdkErrors.SDKError) {
	var request Req

	if unmarshalErr := json.Unmarshal(requestBody, &request); unmarshalErr != nil {
		failErr := sdkErrors.ErrDataUnmarshalFailure.Wrap(unmarshalErr)

		responseBodyForBadRequest, err := MarshalBodyAndRespondOnMarshalFail(
			errorResponseForBadRequest, w,
		)
		if noResponseSentYet := err == nil; noResponseSentYet {
			Respond(http.StatusBadRequest, responseBodyForBadRequest, w)
		}

		// If marshal succeeded, we already responded with a 400 Bad Request with
		// the errorResponseForBadRequest.
		// Otherwise, if marshal failed (err != nil; very unlikely), we already
		// responded with a 400 Bad Request in MarshalBodyAndRespondOnMarshalFail.
		// Either way, we don't need to respond again. Just return the error.
		return nil, failErr
	}

	// We were able to unmarshal the request successfully.
	// We didn't send any failure response to the client so far.
	// Return a pointer to the request to be handled by the calling site.
	return &request, nil
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
//   - *sdkErrors.SDKError: nil if validation passes, error otherwise
type GuardFunc[Req any] func(
	Req, http.ResponseWriter, *http.Request,
) *sdkErrors.SDKError

// ReadParseAndGuard reads the HTTP request body, parses it, and executes
// a guard function in a single operation. This function combines
// readAndParseRequest with guard execution to further reduce boilerplate.
//
// This function performs the following steps:
//  1. Reads the request body from the HTTP request
//  2. Unmarshals the body into the request type
//  3. Executes the guard function for validation
//  4. Returns the parsed request and any errors
//
// Type Parameters:
//   - Req: The request type to unmarshal into
//   - Res: The response type for error cases
//
// Parameters:
//   - w: The response writer for error handling
//   - r: The incoming HTTP request
//   - errorResponse: A response object to send if parsing fails
//   - guard: The guard function to execute for validation
//
// Returns:
//   - *Req: A pointer to the parsed request struct, or nil if any step failed
//   - *sdkErrors.SDKError: ErrDataReadFailure, ErrDataParseFailure, or error
//     from the guard function
//
// Example usage:
//
//	request, err := net.ReadParseAndGuard[
//	    reqres.ShardPutRequest,
//	    reqres.ShardPutResponse](
//	    w, r,
//	    reqres.ShardPutResponse{Err: data.ErrBadInput},
//	    guardShardPutRequest,
//	)
//	if err != nil {
//	    return err
//	}
func ReadParseAndGuard[Req any, Res any](
	w http.ResponseWriter, r *http.Request, errorResponse Res,
	guard GuardFunc[Req],
) (*Req, *sdkErrors.SDKError) {
	request, err := readAndParseRequest[Req, Res](w, r, errorResponse)
	if err != nil {
		return nil, err
	}

	if err = guard(*request, w, r); err != nil {
		return nil, err
	}

	return request, nil
}

// Fail sends an error response to the client.
//
// This function marshals the client response and sends it with the specified
// HTTP status code. It does not return a value; callers should return their
// own error after calling this function.
//
// Type Parameters:
//   - T: The response type to send to the client (e.g.,
//     reqres.ShardPutBadInput)
//
// Parameters:
//   - clientResponse: The response object to send to the client
//   - w: The HTTP response writer for error responses
//   - statusCode: The HTTP status code to send (e.g., http.StatusBadRequest)
//
// Example usage:
//
//	if request.Shard == nil {
//	    net.Fail(reqres.ShardPutBadInput, w, http.StatusBadRequest)
//	    return errors.ErrInvalidInput
//	}
func Fail[T any](
	clientResponse T,
	w http.ResponseWriter,
	statusCode int,
) {
	responseBody, marshalErr := MarshalBodyAndRespondOnMarshalFail(
		clientResponse, w,
	)
	if notRespondedYet := marshalErr == nil; notRespondedYet {
		Respond(statusCode, responseBody, w)
	}
}

// Success sends a success response with HTTP 200 OK.
//
// This is a convenience wrapper around Fail that sends a 200 OK status.
// It maintains semantic clarity by using the name "Success" rather than
// calling Fail directly at call sites.
//
// Type Parameters:
//   - T: The response type to send to the client (e.g.,
//     reqres.ShardPutSuccess)
//
// Parameters:
//   - clientResponse: The response object to send to the client
//   - w: The HTTP response writer
//
// Example usage:
//
//	state.SetShard(request.Shard)
//	net.Success(reqres.ShardPutSuccess, w)
//	return nil
func Success[T any](clientResponse T, w http.ResponseWriter) {
	Fail(clientResponse, w, http.StatusOK)
}

// SuccessWithResponseBody sends a success response with HTTP 200 OK and
// returns the response body for cleanup.
//
// This variant is used when the response body needs to be explicitly cleared
// from memory for security reasons, such as when returning sensitive
// cryptographic data. The caller is responsible for clearing the returned
// byte slice.
//
// Type Parameters:
//   - T: The response type to send to the client (e.g.,
//     reqres.ShardGetResponse)
//
// Parameters:
//   - clientResponse: The response object to send to the client
//   - w: The HTTP response writer
//
// Returns:
//   - []byte: The marshaled response body that should be cleared for security
//
// Example usage:
//
//	responseBody := net.SuccessWithResponseBody(
//	    reqres.ShardGetResponse{Shard: sh}.Success(), w,
//	)
//	defer func() {
//	    mem.ClearBytes(responseBody)
//	}()
//	return nil
func SuccessWithResponseBody[T any](
	clientResponse T, w http.ResponseWriter,
) []byte {
	responseBody, marshalErr := MarshalBodyAndRespondOnMarshalFail(
		clientResponse, w,
	)

	if alreadyResponded := marshalErr != nil; alreadyResponded {
		// Headers already sent. Just return the response body.
		return responseBody
	}

	Respond(http.StatusOK, responseBody, w)
	return responseBody
}
