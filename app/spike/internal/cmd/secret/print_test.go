//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "UTC time",
			input:    time.Date(2025, 1, 15, 14, 30, 45, 0, time.UTC),
			expected: "2025-01-15 14:30:45 UTC",
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "0001-01-01 00:00:00 UTC",
		},
		{
			name:     "midnight",
			input:    time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			expected: "2025-06-01 00:00:00 UTC",
		},
		{
			name:     "end of day",
			input:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2025-12-31 23:59:59 UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(tt.input)
			if result != tt.expected {
				t.Errorf("formatTime() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// createTestCommandWithBuffer creates a command with a captured output buffer.
func createTestCommandWithBuffer() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(buf)
	return cmd, buf
}

func TestPrintSecretResponse_EmptyMetadata(t *testing.T) {
	cmd, buf := createTestCommandWithBuffer()

	response := &data.SecretMetadata{
		Metadata: data.SecretMetaDataContent{},
		Versions: map[int]data.SecretVersionInfo{},
	}

	printSecretResponse(cmd, response)

	output := buf.String()

	// Empty metadata should not print Metadata section
	if strings.Contains(output, "Metadata:") {
		t.Error("Should not print Metadata section when metadata is empty")
	}

	// Empty versions should not print Versions section
	if strings.Contains(output, "Secret Versions:") {
		t.Error("Should not print Versions section when versions is empty")
	}
}

func TestPrintSecretResponse_WithMetadata(t *testing.T) {
	cmd, buf := createTestCommandWithBuffer()

	createdTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2025, 1, 16, 14, 30, 0, 0, time.UTC)

	response := &data.SecretMetadata{
		Metadata: data.SecretMetaDataContent{
			CurrentVersion: 3,
			OldestVersion:  1,
			CreatedTime:    createdTime,
			UpdatedTime:    updatedTime,
			MaxVersions:    10,
		},
		Versions: map[int]data.SecretVersionInfo{},
	}

	printSecretResponse(cmd, response)

	output := buf.String()

	// Check metadata section is present
	if !strings.Contains(output, "Metadata:") {
		t.Error("Should print Metadata section")
	}

	// Check all metadata fields
	expectedFields := []string{
		"Current Version : 3",
		"Oldest Version  : 1",
		"Max Versions    : 10",
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Should contain %q", field)
		}
	}

	// Check timestamps are formatted
	if !strings.Contains(output, "Created Time") {
		t.Error("Should contain Created Time")
	}
	if !strings.Contains(output, "Last Updated") {
		t.Error("Should contain Last Updated")
	}
}

func TestPrintSecretResponse_WithVersions(t *testing.T) {
	cmd, buf := createTestCommandWithBuffer()

	createdTime1 := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	createdTime2 := time.Date(2025, 1, 16, 11, 0, 0, 0, time.UTC)
	deletedTime := time.Date(2025, 1, 17, 12, 0, 0, 0, time.UTC)

	response := &data.SecretMetadata{
		Metadata: data.SecretMetaDataContent{},
		Versions: map[int]data.SecretVersionInfo{
			1: {
				CreatedTime: createdTime1,
				DeletedTime: &deletedTime,
			},
			2: {
				CreatedTime: createdTime2,
				DeletedTime: nil,
			},
		},
	}

	printSecretResponse(cmd, response)

	output := buf.String()

	// Check versions section is present
	if !strings.Contains(output, "Secret Versions:") {
		t.Error("Should print Secret Versions section")
	}

	// Check version numbers are present
	if !strings.Contains(output, "Version 1:") {
		t.Error("Should contain Version 1")
	}
	if !strings.Contains(output, "Version 2:") {
		t.Error("Should contain Version 2")
	}

	// Check Created is present for versions
	if strings.Count(output, "Created:") < 2 {
		t.Error("Should contain Created for each version")
	}

	// Check deleted time is shown for version 1
	if !strings.Contains(output, "Deleted:") {
		t.Error("Should contain Deleted for soft-deleted version")
	}
}

func TestPrintSecretResponse_FullResponse(t *testing.T) {
	cmd, buf := createTestCommandWithBuffer()

	createdTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2025, 1, 20, 15, 0, 0, 0, time.UTC)
	versionCreated := time.Date(2025, 1, 18, 12, 0, 0, 0, time.UTC)

	response := &data.SecretMetadata{
		Metadata: data.SecretMetaDataContent{
			CurrentVersion: 5,
			OldestVersion:  2,
			CreatedTime:    createdTime,
			UpdatedTime:    updatedTime,
			MaxVersions:    20,
		},
		Versions: map[int]data.SecretVersionInfo{
			5: {
				CreatedTime: versionCreated,
				DeletedTime: nil,
			},
		},
	}

	printSecretResponse(cmd, response)

	output := buf.String()

	// Both sections should be present
	if !strings.Contains(output, "Metadata:") {
		t.Error("Should print Metadata section")
	}
	if !strings.Contains(output, "Secret Versions:") {
		t.Error("Should print Secret Versions section")
	}

	// Check separators are used
	separatorCount := strings.Count(output, strings.Repeat("-", 50))
	if separatorCount < 2 {
		t.Errorf("Should have at least 2 separators, got %d", separatorCount)
	}
}

func TestPrintSecretResponse_SeparatorLength(t *testing.T) {
	cmd, buf := createTestCommandWithBuffer()

	response := &data.SecretMetadata{
		Metadata: data.SecretMetaDataContent{
			CurrentVersion: 1,
			MaxVersions:    10,
			CreatedTime:    time.Now(),
			UpdatedTime:    time.Now(),
		},
		Versions: map[int]data.SecretVersionInfo{},
	}

	printSecretResponse(cmd, response)

	output := buf.String()

	// Check that separator is exactly 50 dashes
	expectedSeparator := strings.Repeat("-", 50)
	if !strings.Contains(output, expectedSeparator) {
		t.Error("Separator should be exactly 50 dashes")
	}
}
