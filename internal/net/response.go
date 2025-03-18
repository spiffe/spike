//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"

	"github.com/spiffe/spike/internal/log"
)

// MarshalBody serializes a response object to JSON and handles error cases.
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
//   - []byte - The marshaled JSON bytes, or nil if marshaling failed
func MarshalBody(res any, w http.ResponseWriter) []byte {
	body, err := json.Marshal(res)

	if err != nil {
		log.Log().Error("marshalBody",
			"msg", "Problem generating response",
			"err", err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		_, err = w.Write([]byte(`{"error":"internal server error"}`))
		if err != nil {
			log.Log().Error("marshalBody",
				"msg", "Problem writing response",
				"err", err.Error())
			return nil
		}

		return nil
	}

	return body
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
	w.Header().Set("Content-Type", "application/json")

	// Add cache invalidation headers
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(statusCode)

	_, err := w.Write(body)
	if err != nil {
		log.Log().Error("routeKeep",
			"msg", "Problem writing response",
			"err", err.Error())
	}
}

// Fallback handles requests to undefined routes by returning a 400 Bad Request.
//
// This function serves as a catch-all handler for undefined routes, logging the
// request details and returning a standardized error response. It uses
// MarshalBody to generate the response and handles any errors during response
// writing.
//
// Parameters:
//   - w: http.ResponseWriter - The response writer
//   - r: *http.Request - The incoming request
//
// The response always includes:
//   - Status: 400 Bad Request
//   - Content-Type: application/json
//   - Body: JSON object with an error field
func Fallback(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("fallback",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditFallback

	body := MarshalBody(reqres.FallbackResponse{Err: data.ErrBadInput}, w)
	if body == nil {
		return errors.New("failed to marshal response body")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	if _, err := w.Write(body); err != nil {
		log.Log().Error("routeFallback",
			"msg", "Problem writing response",
			"err", err.Error())
		return err
	}

	return nil
}

// NotReady handles requests when the system has not initialized its backing
// store with a root key by returning a 400 Bad Request.
//
// This function uses MarshalBody to generate the response and handles any
// errors during response writing.
//
// Parameters:
//   - w: http.ResponseWriter - The response writer
//   - r: *http.Request - The incoming request
//   - audit: *log.AuditEntry - The audit log entry for this request
//
// The response always includes:
//   - Status: 400 Bad Request
//   - Content-Type: application/json
//   - Body: JSON object with an error field containing ErrLowEntropy
//
// Returns:
//   - error: Returns nil on success, or an error if response marshaling or
//     writing fails
func NotReady(
	w http.ResponseWriter, r *http.Request, audit *log.AuditEntry,
) error {
	log.Log().Info("not-ready",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = log.AuditBlocked

	body := MarshalBody(reqres.FallbackResponse{Err: data.ErrNotReady}, w)
	if body == nil {
		return errors.New("failed to marshal response body")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)

	if _, err := w.Write(body); err != nil {
		log.Log().Error("routeNotReady",
			"msg", "Problem writing response",
			"err", err.Error())
		return err
	}

	return nil
}
