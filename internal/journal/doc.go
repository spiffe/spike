//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package journal provides audit logging for SPIKE components.
//
// This package records security-relevant events as structured JSON entries,
// enabling compliance tracking and forensic analysis. Each audit entry
// captures the actor (SPIFFE ID), action, resource, timing, and outcome.
//
// Key types:
//
//   - AuditEntry: Represents a single audit event with fields for component,
//     user ID, action, resource path, state, and duration.
//   - AuditAction: Defines the type of operation (enter, exit, create, read,
//     list, delete, undelete, blocked).
//   - AuditState: Indicates the outcome (audit-entry-created, audit-success,
//     audit-errored).
//
// Key functions:
//
//   - Audit: Writes an AuditEntry as a JSON log line to stdout.
//   - AuditRequest: Convenience function to log HTTP request details.
//
// Output format:
//
// Audit entries are written as JSON objects with a timestamp and nested
// audit data:
//
//	{"time":"2024-01-15T10:30:00Z","audit":{"component":"...","action":"..."}}
//
// If JSON marshaling fails, the package calls log.FatalLn to terminate,
// as audit failures are considered critical in a security context.
package journal
