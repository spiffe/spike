//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package journal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuditEntry_Fields(t *testing.T) {
	now := time.Now()
	entry := AuditEntry{
		Component: "nexus",
		TrailID:   "trail-123",
		Timestamp: now,
		UserID:    "user-456",
		Action:    AuditCreate,
		Path:      "/v1/secrets",
		Resource:  "path=secrets/db",
		SessionID: "session-789",
		State:     AuditSuccess,
		Err:       "",
		Duration:  100 * time.Millisecond,
	}

	if entry.Component != "nexus" {
		t.Errorf("Component = %q, want %q", entry.Component, "nexus")
	}
	if entry.TrailID != "trail-123" {
		t.Errorf("TrailID = %q, want %q", entry.TrailID, "trail-123")
	}
	if entry.Action != AuditCreate {
		t.Errorf("Action = %q, want %q", entry.Action, AuditCreate)
	}
	if entry.State != AuditSuccess {
		t.Errorf("State = %q, want %q", entry.State, AuditSuccess)
	}
}

func TestAuditAction_Constants(t *testing.T) {
	tests := []struct {
		action   AuditAction
		expected string
	}{
		{AuditEnter, "enter"},
		{AuditExit, "exit"},
		{AuditCreate, "create"},
		{AuditList, "list"},
		{AuditDelete, "delete"},
		{AuditRead, "read"},
		{AuditUndelete, "undelete"},
		{AuditFallback, "fallback"},
		{AuditBlocked, "blocked"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			if string(tt.action) != tt.expected {
				t.Errorf("AuditAction = %q, want %q", tt.action, tt.expected)
			}
		})
	}
}

func TestAuditState_Constants(t *testing.T) {
	tests := []struct {
		state    AuditState
		expected string
	}{
		{AuditEntryCreated, "audit-entry-created"},
		{AuditErrored, "audit-errored"},
		{AuditSuccess, "audit-success"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			if string(tt.state) != tt.expected {
				t.Errorf("AuditState = %q, want %q", tt.state, tt.expected)
			}
		})
	}
}

func TestAudit_OutputsValidJSON(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	entry := AuditEntry{
		Component: "test-component",
		TrailID:   "test-trail",
		Timestamp: time.Now(),
		UserID:    "test-user",
		Action:    AuditRead,
		Path:      "/test/path",
		Resource:  "test-resource",
		SessionID: "test-session",
		State:     AuditSuccess,
		Err:       "",
		Duration:  50 * time.Millisecond,
	}

	Audit(entry)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output is valid JSON
	var logLine AuditLogLine
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logLine); err != nil {
		t.Errorf("Audit() output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Verify fields are preserved
	if logLine.AuditEntry.Component != "test-component" {
		t.Errorf("AuditEntry.Component = %q, want %q",
			logLine.AuditEntry.Component, "test-component")
	}
	if logLine.AuditEntry.Action != AuditRead {
		t.Errorf("AuditEntry.Action = %q, want %q",
			logLine.AuditEntry.Action, AuditRead)
	}
	if logLine.AuditEntry.State != AuditSuccess {
		t.Errorf("AuditEntry.State = %q, want %q",
			logLine.AuditEntry.State, AuditSuccess)
	}
}

func TestAuditRequest_SetsFieldsCorrectly(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	req := httptest.NewRequest("GET", "/v1/secrets?path=db/creds", nil)
	audit := &AuditEntry{
		TrailID:   "trail-123",
		UserID:    "user-456",
		SessionID: "session-789",
		State:     AuditEntryCreated,
	}

	AuditRequest("TestFunction", req, audit, AuditRead)

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify the audit entry was updated
	if audit.Component != "TestFunction" {
		t.Errorf("audit.Component = %q, want %q", audit.Component, "TestFunction")
	}
	if audit.Path != "/v1/secrets" {
		t.Errorf("audit.Path = %q, want %q", audit.Path, "/v1/secrets")
	}
	if audit.Resource != "path=db/creds" {
		t.Errorf("audit.Resource = %q, want %q", audit.Resource, "path=db/creds")
	}
	if audit.Action != AuditRead {
		t.Errorf("audit.Action = %q, want %q", audit.Action, AuditRead)
	}

	// Verify valid JSON was output
	var logLine AuditLogLine
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logLine); err != nil {
		t.Errorf("AuditRequest() output is not valid JSON: %v", err)
	}
}

func TestAuditLogLine_JSONStructure(t *testing.T) {
	now := time.Now()
	logLine := AuditLogLine{
		Timestamp: now,
		AuditEntry: AuditEntry{
			Component: "nexus",
			TrailID:   "trail-123",
			Action:    AuditCreate,
			State:     AuditSuccess,
		},
	}

	data, err := json.Marshal(logLine)
	if err != nil {
		t.Fatalf("Failed to marshal AuditLogLine: %v", err)
	}

	// Verify JSON structure
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check "time" field exists
	if _, ok := decoded["time"]; !ok {
		t.Error("JSON missing 'time' field")
	}

	// Check "audit" field exists
	if _, ok := decoded["audit"]; !ok {
		t.Error("JSON missing 'audit' field")
	}
}
