//	  \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//	\\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package net

import (
	"net/http"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike-sdk-go/journal"
)

// Handler is a function type that processes HTTP requests with audit
// logging support.
//
// Parameters:
//   - w: HTTP response writer for sending the response
//   - r: HTTP request containing the incoming request data
//   - audit: Audit entry for logging the request lifecycle
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, error on failure
type Handler func(
	w http.ResponseWriter, r *http.Request, audit *journal.AuditEntry,
) *sdkErrors.SDKError

// HandleRoute wraps an HTTP handler with audit logging functionality.
// It creates and manages audit log entries for the request lifecycle,
// including
// - Generating unique trail IDs
// - Recording timestamps and durations
// - Tracking request status (created, success, error)
// - Capturing error information
//
// The wrapped handler is mounted at the root path ("/") and automatically
// logs entry and exit audit events for all requests.
//
// Parameters:
//   - h: Handler function to wrap with audit logging
func HandleRoute(h Handler) {
	http.HandleFunc("/", func(
		writer http.ResponseWriter, request *http.Request,
	) {
		now := time.Now()
		id := crypto.ID()

		entry := journal.AuditEntry{
			TrailID:   id,
			Timestamp: now,
			UserID:    "",
			Action:    journal.AuditEnter,
			Path:      request.URL.Path,
			Resource:  "",
			SessionID: "",
			State:     journal.AuditEntryCreated,
		}
		journal.Audit(entry)

		err := h(writer, request, &entry)
		if err == nil {
			entry.Action = journal.AuditExit
			entry.State = journal.AuditSuccess
		} else {
			entry.Action = journal.AuditExit
			entry.State = journal.AuditErrored
			entry.Err = err.Error()
		}

		entry.Duration = time.Since(now)
		journal.Audit(entry)
	})
}
