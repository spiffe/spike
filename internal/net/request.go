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

// ReadRequestBody reads the entire request body from an HTTP request.
// It returns the body as a byte slice if successful, or an error if reading
// fails. On error, it writes a 400 Bad Request status to the response writer
// and returns the error for propagation to the caller.
func ReadRequestBody(
	w http.ResponseWriter, r *http.Request,
) ([]byte, *sdkErrors.SDKError) {
	body, err := net.RequestBody(r)
	if err != nil {
		failErr := sdkErrors.ErrReadFailure.Wrap(err)
		failErr.Msg = "problem reading request body"

		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := io.WriteString(w, "")
		if writeErr != nil {
			failErr = failErr.Wrap(writeErr)
			failErr.Msg = "problem writing response"
		}

		return nil, failErr
	}

	return body, nil
}

// HandleRequestError handles HTTP request errors by writing a 400 Bad Request
// status to the response writer. If err is nil, it returns nil. Otherwise, it
// writes the error status and returns a joined error containing both the
// original error and any error encountered while writing the response.
func HandleRequestError(
	w http.ResponseWriter, err *sdkErrors.SDKError,
) *sdkErrors.SDKError {
	if err == nil {
		return nil
	}

	w.WriteHeader(http.StatusBadRequest)
	_, writeErr := io.WriteString(w, "")

	return sdkErrors.ErrBadRequest.Wrap(err).Wrap(writeErr)
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
	const fName = "HandleRequest"

	var request Req

	unmarshalErr := json.Unmarshal(requestBody, &request)
	var sdkErr *sdkErrors.SDKError
	if unmarshalErr != nil {
		sdkErr = sdkErrors.ErrUnmarshalFailure.Wrap(unmarshalErr)
	}

	if err := HandleRequestError(w, sdkErr); err != nil {
		log.ErrorErr(fName, *err)

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
//
// Returns:
//   - *Req - A pointer to the parsed request struct, or nil if parsing failed
//   - *sdkErrors.SDKError - ErrReadFailure, ErrParseFailure, or nil
//
// Example usage:
//
//	request, err := net.ReadAndParseRequest[
//	    reqres.SecretDeleteRequest,
//	    reqres.SecretDeleteResponse](
//	    w, r,
//	    reqres.SecretDeleteResponse{Err: data.ErrBadInput},
//	)
//	if err != nil {
//	    return err
//	}
func ReadAndParseRequest[Req any, Res any](
	w http.ResponseWriter,
	r *http.Request,
	errorResponse Res,
) (*Req, *sdkErrors.SDKError) {
	requestBody, err := ReadRequestBody(w, r)
	if err != nil {
		return nil, err
	}

	request := HandleRequest[Req, Res](requestBody, w, errorResponse)
	if request == nil {
		return nil, sdkErrors.ErrParseFailure
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
type GuardFunc[Req any] func(Req, http.ResponseWriter, *http.Request) *sdkErrors.SDKError

// ReadParseAndGuard reads the HTTP request body, parses it, and executes
// a guard function in a single operation. This function combines
// ReadAndParseRequest with guard execution to further reduce boilerplate.
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
//   - w: http.ResponseWriter - The response writer for error handling
//   - r: *http.Request - The incoming HTTP request
//   - errorResponse: Res - A response object to send if parsing fails
//   - guard: GuardFunc[Req] - The guard function to execute for validation
//
// Returns:
//   - *Req - A pointer to the parsed request struct, or nil if any step failed
//   - *sdkErrors.SDKError - ErrReadFailure, ErrParseFailure, or error from
//     the guard function
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
	w http.ResponseWriter,
	r *http.Request,
	errorResponse Res,
	guard GuardFunc[Req],
) (*Req, *sdkErrors.SDKError) {
	request, err := ReadAndParseRequest[Req, Res](w, r, errorResponse)
	if err != nil {
		return nil, err
	}

	if err = guard(*request, w, r); err != nil {
		return nil, err
	}

	return request, nil
}

//// FailIfError is a helper function that fails a request if an error occurred
//// during validation or processing.
////
//// This function provides a reusable pattern for validating inputs and
//// responding with appropriate error messages. If the internal error is nil,
//// the function returns nil immediately. Otherwise, it marshals the client
//// response and sends it with a 400 Bad Request status, then returns the
//// specified error to the caller.
////
//// Type Parameters:
////   - T: The response type to send to the client (e.g.,
////     reqres.PolicyCreateBadInput)
////
//// Parameters:
////   - internalError: The error to check (e.g., from validation functions).
////     If nil, the function returns nil immediately.
////   - errorToRespond: The error to return to the caller (e.g.,
////     apiErr.ErrInvalidInput)
////   - clientResponse: The response object to send to the client if there is
////     an error
////   - w: The HTTP response writer for error responses
////
//// Returns:
////   - error: Returns errorToRespond if internalError is not nil, otherwise
////     returns nil
////
//// Example usage:
////
////	err := validation.ValidateName(name)
////	if err := net.FailIfError(
////	    err, apiErr.ErrInvalidInput,
////	    reqres.PolicyCreateBadInput, w); err != nil {
////	    return err
////	}
//func FailIfError[T any](
//	internalError error, errorToRespond error,
//	clientResponse T, w http.ResponseWriter,
//) error {
//	if internalError != nil {
//		responseBody, marshalErr := MarshalBodyAndRespondOnMarshalFail(
//			clientResponse, w,
//		)
//		if alreadyResponded := marshalErr != nil; !alreadyResponded {
//			Respond(http.StatusBadRequest, responseBody, w)
//		}
//		return errorToRespond
//	}
//	return nil
//}
//
//// FailIf is a helper function that conditionally fails a request by sending
//// an error response based on a boolean condition.
////
//// This function provides a reusable pattern for conditional error responses,
//// such as authorization checks or validation conditions. If the condition is
//// true, it marshals the client response and sends it with the specified HTTP
//// status code, then returns the specified error to the caller.
////
//// Type Parameters:
////   - T: The response type to send to the client (e.g.,
////     reqres.ShardPutUnauthorized)
////
//// Parameters:
////   - condition: If true, fail the request with an error response
////   - clientResponse: The response object to send to the client if condition
////     is true
////   - w: The HTTP response writer for error responses
////   - statusCode: The HTTP status code to send (e.g., http.StatusUnauthorized)
////   - errorToRespond: The error to return to the caller (e.g.,
////     apiErr.ErrUnauthorized)
////
//// Returns:
////   - error: Returns errorToRespond if condition is true, otherwise returns
////     nil
////
//// Example usage:
////
////	if err := net.FailIf(
////	    !spiffeid.PeerCanTalkToKeeper(peerSPIFFEID.String()),
////	    reqres.ShardPutUnauthorized, w,
////	    http.StatusUnauthorized, apiErr.ErrUnauthorized,
////	); err != nil {
////	    return err
////	}
//func FailIf[T any](
//	condition bool,
//	clientResponse T,
//	w http.ResponseWriter,
//	statusCode int,
//	errorToRespond error,
//) error {
//	if condition {
//		responseBody, marshalErr := MarshalBodyAndRespondOnMarshalFail(
//			clientResponse, w,
//		)
//		if alreadyResponded := marshalErr != nil; !alreadyResponded {
//			Respond(statusCode, responseBody, w)
//		}
//		return errorToRespond
//	}
//	return nil
//}

// Fail sends an error response and returns the specified error.
//
// This function is used when a request should fail and an error response
// needs to be sent to the client. It marshals the client response and sends
// it with the specified HTTP status code, then returns the error to the
// caller for propagation up the call stack.
//
// Type Parameters:
//   - T: The response type to send to the client (e.g.,
//     reqres.ShardPutBadInput)
//
// Parameters:
//   - clientResponse: The response object to send to the client
//   - w: The HTTP response writer for error responses
//   - statusCode: The HTTP status code to send (e.g., http.StatusBadRequest)
//   - errorToReturn: The error to return to the caller (e.g.,
//     errors.ErrInvalidInput)
//
// Returns:
//   - error: Always returns errorToReturn
//
// Example usage:
//
//	if request.Shard == nil {
//	    return net.Fail(
//	        reqres.ShardPutBadInput, w,
//	        http.StatusBadRequest, errors.ErrInvalidInput,
//	    )
//	}
func Fail[T any](
	clientResponse T,
	w http.ResponseWriter,
	statusCode int,
	errorToReturn error,
) error {
	responseBody, marshalErr := MarshalBodyAndRespondOnMarshalFail(
		clientResponse, w,
	)
	if alreadyResponded := marshalErr != nil; !alreadyResponded {
		Respond(statusCode, responseBody, w)
	}
	return errorToReturn
}

// Success sends a success response with HTTP 200 OK.
//
// This function marshals the client response and sends it with a 200 OK
// status. It does not return a value, maintaining API symmetry with
// SuccessWithResponseBody (which returns the response body for cleanup).
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
	responseBody, marshalErr := MarshalBodyAndRespondOnMarshalFail(
		clientResponse, w,
	)
	if alreadyResponded := marshalErr != nil; !alreadyResponded {
		return
	}
	Respond(http.StatusOK, responseBody, w)
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
