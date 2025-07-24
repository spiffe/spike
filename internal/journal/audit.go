//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package journal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	logger "github.com/spiffe/spike-sdk-go/log"
)

type AuditState string

const AuditEntryCreated AuditState = "audit-entry-created"
const AuditErrored AuditState = "audit-errored"
const AuditSuccess AuditState = "audit-success"

type AuditAction string

const AuditEnter AuditAction = "enter"
const AuditExit AuditAction = "exit"
const AuditCreate AuditAction = "create"
const AuditList AuditAction = "list"
const AuditDelete AuditAction = "delete"
const AuditRead AuditAction = "read"
const AuditUndelete AuditAction = "undelete"
const AuditFallback AuditAction = "fallback"
const AuditBlocked AuditAction = "blocked"

// AuditEntry represents a single audit log entry containing information about
// user actions within the system.
type AuditEntry struct {
	// Component is the name of the component that performed the action.
	Component string

	// TrailId is a unique identifier for the audit trail
	TrailId string

	// Timestamp indicates when the audited action occurred
	Timestamp time.Time

	// UserId identifies the user who performed the action
	UserId string

	// Action describes what operation was performed
	Action AuditAction

	// Path is the URL path of the request
	Path string

	// Resource identifies the object or entity acted upon
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

type AuditLogLine struct {
	Timestamp  time.Time  `json:"time"`
	AuditEntry AuditEntry `json:"audit"`
}

// Audit logs an audit entry as JSON to the standard log output.
// If JSON marshaling fails, it logs an error using the structured logger
// but continues execution.
func Audit(entry AuditEntry) {
	audit := AuditLogLine{
		Timestamp:  time.Now(),
		AuditEntry: entry,
	}

	body, err := json.Marshal(audit)
	if err != nil {
		// If you cannot audit, crashing is the best option.
		logger.FatalLn("Audit",
			"message", "Problem marshalling audit entry",
			"err", err.Error())
		return
	}

	fmt.Println(string(body))
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
	audit.Component = fName
	audit.Path = r.URL.Path
	audit.Resource = r.URL.RawQuery
	audit.Action = action
	Audit(*audit)
}
