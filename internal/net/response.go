//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"net/http"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/journal"
	"github.com/spiffe/spike-sdk-go/net"
)

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

	return net.RespondFallbackWithStatus(
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

	return net.RespondFallbackWithStatus(
		w, http.StatusServiceUnavailable, sdkErrors.ErrStateNotReady.Code,
	)
}
