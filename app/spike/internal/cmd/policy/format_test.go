//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// createTestCommandWithFormat creates a Cobra command with a format flag.
func createTestCommandWithFormat(format string) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", format, "Output format")
	return cmd
}

func TestFormatPoliciesOutput_EmptyList(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		policies *[]data.PolicyListItem
		expected string
	}{
		{
			name:     "nil policies human format",
			format:   "human",
			policies: nil,
			expected: "No policies found.",
		},
		{
			name:     "empty slice human format",
			format:   "human",
			policies: &[]data.PolicyListItem{},
			expected: "No policies found.",
		},
		{
			name:     "nil policies json format",
			format:   "json",
			policies: nil,
			expected: "[]",
		},
		{
			name:     "empty slice json format",
			format:   "json",
			policies: &[]data.PolicyListItem{},
			expected: "[]",
		},
		{
			name:     "default format (empty string)",
			format:   "",
			policies: nil,
			expected: "No policies found.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommandWithFormat(tt.format)
			result := formatPoliciesOutput(cmd, tt.policies)

			if result != tt.expected {
				t.Errorf("formatPoliciesOutput() = %q, want %q",
					result, tt.expected)
			}
		})
	}
}

func TestFormatPoliciesOutput_InvalidFormat(t *testing.T) {
	cmd := createTestCommandWithFormat("xml")
	policies := &[]data.PolicyListItem{}

	result := formatPoliciesOutput(cmd, policies)

	if !strings.Contains(result, "Error: Invalid format") {
		t.Errorf("formatPoliciesOutput() should return error for invalid format")
	}
	if !strings.Contains(result, "xml") {
		t.Errorf("formatPoliciesOutput() should mention the invalid format")
	}
}

func TestFormatPoliciesOutput_HumanFormat(t *testing.T) {
	policies := &[]data.PolicyListItem{
		{
			ID:   "123e4567-e89b-12d3-a456-426614174000",
			Name: "test-policy",
		},
	}
	cmd := createTestCommandWithFormat("human")
	result := formatPoliciesOutput(cmd, policies)

	normalized := normalizePolicyOutput(result)

	// Check header
	if !strings.Contains(result, "POLICIES") {
		t.Error("Human format should contain 'POLICIES' header")
	}

	// Check policy fields are present (PolicyListItem only has ID and Name)
	expectedFields := []string{
		"ID: 123e4567-e89b-12d3-a456-426614174000",
		"Name: test-policy",
	}

	// Check policy fields are present
	for _, field := range expectedFields {
		if !strings.Contains(normalized, field) {
			t.Errorf("Human format should contain %q", field)
		}
	}
}

func TestFormatPoliciesOutput_JSONFormat(t *testing.T) {
	policies := &[]data.PolicyListItem{
		{
			ID:   "123e4567-e89b-12d3-a456-426614174000",
			Name: "test-policy",
		},
	}

	cmd := createTestCommandWithFormat("json")
	result := formatPoliciesOutput(cmd, policies)

	// Verify it's valid JSON
	var decoded []data.PolicyListItem
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Errorf("JSON format should produce valid JSON: %v", err)
	}

	// Verify content
	if len(decoded) != 1 {
		t.Errorf("Expected 1 policy, got %d", len(decoded))
	}
	if decoded[0].Name != "test-policy" {
		t.Errorf("Policy name = %q, want %q", decoded[0].Name, "test-policy")
	}
}

func TestFormatPolicy_NilPolicy(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"human format", "human", "No policy found."},
		{"json format", "json", "No policy found."},
		{"default format", "", "No policy found."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommandWithFormat(tt.format)
			result := formatPolicy(cmd, nil)

			if result != tt.expected {
				t.Errorf("formatPolicy(nil) = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatPolicy_InvalidFormat(t *testing.T) {
	cmd := createTestCommandWithFormat("yaml")
	policy := &data.Policy{Name: "test"}

	result := formatPolicy(cmd, policy)

	if !strings.Contains(result, "Error: Invalid format") {
		t.Error("formatPolicy() should return error for invalid format")
	}
}

func TestFormatPolicy_HumanFormat(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	policy := &data.Policy{
		ID:              "123e4567-e89b-12d3-a456-426614174000",
		Name:            "admin-policy",
		SPIFFEIDPattern: "^spiffe://example\\.org/admin/.*$",
		PathPattern:     "^.*$",
		Permissions:     []data.PolicyPermission{"read", "write", "list", "super"},
		CreatedAt:       createdAt,
	}

	cmd := createTestCommandWithFormat("human")
	result := formatPolicy(cmd, policy)

	// Check header
	if !strings.Contains(result, "POLICY DETAILS") {
		t.Error("Human format should contain 'POLICY DETAILS' header")
	}

	// Check all fields are present
	expectedFields := []string{
		"ID: 123e4567-e89b-12d3-a456-426614174000",
		"Name: admin-policy",
		"SPIFFE ID Pattern: ^spiffe://example\\.org/admin/.*$",
		"Path Pattern: ^.*$",
		"Permissions: read, write, list, super",
	}

	for _, field := range expectedFields {
		if !strings.Contains(result, field) {
			t.Errorf("Human format should contain %q", field)
		}
	}
}

func TestFormatPolicy_JSONFormat(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	policy := &data.Policy{
		ID:              "123e4567-e89b-12d3-a456-426614174000",
		Name:            "test-policy",
		SPIFFEIDPattern: "^spiffe://example\\.org/.*$",
		PathPattern:     "^secrets/.*$",
		Permissions:     []data.PolicyPermission{"read"},
		CreatedAt:       createdAt,
	}

	cmd := createTestCommandWithFormat("json")
	result := formatPolicy(cmd, policy)

	// Verify it's valid JSON
	var decoded data.Policy
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Errorf("JSON format should produce valid JSON: %v", err)
	}

	// Verify content
	if decoded.Name != "test-policy" {
		t.Errorf("Policy name = %q, want %q", decoded.Name, "test-policy")
	}
	if decoded.ID != "123e4567-e89b-12d3-a456-426614174000" {
		t.Errorf("Policy ID = %q, want %q",
			decoded.ID, "123e4567-e89b-12d3-a456-426614174000")
	}
}

func TestFormatPoliciesOutput_MultiplePolicies(t *testing.T) {
	policies := &[]data.PolicyListItem{
		{
			ID:   "id-1",
			Name: "policy-one",
		},
		{
			ID:   "id-2",
			Name: "policy-two",
		},
		{
			ID:   "id-3",
			Name: "policy-three",
		},
	}

	cmd := createTestCommandWithFormat("human")
	result := formatPoliciesOutput(cmd, policies)

	normalized := normalizePolicyOutput(result)

	// Check all policies are present
	if !strings.Contains(normalized, "policy-one") {
		t.Errorf("Should contain policy-one")
	}
	if !strings.Contains(normalized, "policy-two") {
		t.Error("Should contain policy-two")
	}
	if !strings.Contains(normalized, "policy-three") {
		t.Error("Should contain policy-three")
	}
	// Check separators between policies
	separatorCount := strings.Count(normalized, "\n")
	if separatorCount < 3 {
		t.Errorf("Expected at least 3 separators, got %d", separatorCount)
	}
}
