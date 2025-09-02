//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

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

		responseBody := MarshalBody(errorResponse, w)
		if responseBody == nil {
			return nil
		}

		Respond(http.StatusBadRequest, responseBody, w)
		return nil
	}

	return &request
}
