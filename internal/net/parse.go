//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"net/http"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// readAndParseRequest reads the HTTP request body and parses it into a typed
// request struct in a single operation. This function combines
// ReadRequestBodyAndRespondOnFail and UnmarshalAndRespondOnFail to reduce
// boilerplate in route handlers.
//
// This function performs the following steps:
//  1. Reads the request body from the HTTP request
//  2. Returns ErrDataReadFailure if reading fails
//  3. Unmarshals the body into the request type
//  4. Returns ErrDataParseFailure (wrapping ErrDataUnmarshalFailure) if
//     unmarshaling fails
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
//   - *sdkErrors.SDKError - ErrDataReadFailure, ErrDataParseFailure, or nil
//
// Example usage:
//
//	request, err := net.readAndParseRequest[
//	    reqres.SecretDeleteRequest,
//	    reqres.SecretDeleteResponse](
//	    w, r,
//	    reqres.SecretDeleteResponse{Err: data.ErrBadInput},
//	)
//	if err != nil {
//	    return err
//	}
func readAndParseRequest[Req any, Res any](
	w http.ResponseWriter,
	r *http.Request,
	errorResponse Res,
) (*Req, *sdkErrors.SDKError) {
	requestBody, err := ReadRequestBodyAndRespondOnFail(w, r)
	if err != nil {
		return nil, err
	}

	request, err := UnmarshalAndRespondOnFail[Req, Res](
		requestBody, w, errorResponse,
	)
	if err != nil {
		failErr := sdkErrors.ErrDataParseFailure.Wrap(err)
		failErr.Msg = "problem parsing request body"
		return nil, failErr
	}

	return request, nil
}
