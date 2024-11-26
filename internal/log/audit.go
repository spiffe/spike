//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type AuditState string

const AuditCreated AuditState = "created"
const AuditErrored AuditState = "error"
const AuditSuccess AuditState = "success"

type AuditAction string

const AuditEnter AuditAction = "enter"
const AuditCreate AuditAction = "create"
const AuditList AuditAction = "list"
const AuditInitCheck AuditAction = "initialization-check"
const AuditLogin AuditAction = "login"
const AuditDelete AuditAction = "delete"
const AuditRead AuditAction = "read"
const AuditUndelete AuditAction = "undelete"
const AuditFallback AuditAction = "fallback"

// AuditEntry represents a single audit log entry containing information about
// user actions within the system.
type AuditEntry struct {
	// Id is a unique identifier for the audit trail
	TrailId string

	// Timestamp indicates when the audited action occurred
	Timestamp time.Time

	// UserId identifies the user who performed the action
	UserId string

	// Action describes what operation was performed
	Action AuditAction

	// Path is the URL path of the request
	Path string

	// Resource identifies the object or entity that was acted upon
	Resource string

	// SessionID links the action to a specific user session
	SessionID string

	// State represents the state of the resource after the action
	State AuditState

	// Err contains an error message if the action failed
	Err string

	// Duration is the time taken to process the action
	Duration time.Duration
}

// Audit logs an audit entry as JSON to the standard log output.
// If JSON marshaling fails, it logs an error using the structured logger
// but continues execution.
func Audit(entry AuditEntry) {
	body, err := json.Marshal(entry)
	if err != nil {
		Log().Error("Audit",
			"msg", "Problem marshalling audit entry",
			"err", err.Error())
		return
	}

	log.Println(string(body))
}

// AuditRequest logs the details of an HTTP request and updates the audit entry
// with the specified action. It captures the HTTP method, path, and query
// parameters of the request for audit logging purposes.
//
// Parameters:
//   - fName: The name of the function or component making the request
//   - r: The HTTP request being audited
//   - audit: A pointer to the AuditEntry to be updated
//   - action: The AuditAction to be recorded in the audit entry
func AuditRequest(fName string,
	r *http.Request, audit *AuditEntry, action AuditAction) {
	Log().Info(fName, "method", r.Method, "path", r.URL.Path,
		"query", r.URL.RawQuery)
	audit.Action = action
}
