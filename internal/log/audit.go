//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package log

import (
	"encoding/json"
	"log"
	"time"
)

type AuditState string

var AuditCreated AuditState = "created"
var AuditErrored AuditState = "error"
var AuditSuccess AuditState = "success"

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
	Action string

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
