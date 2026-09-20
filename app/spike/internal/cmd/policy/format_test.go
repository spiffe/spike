//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

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
			result, formatErr := formatPoliciesOutput(cmd, tt.policies)
			if formatErr != nil {
				t.Fatalf("formatPoliciesOutput() error = %v, want nil", formatErr)
			}

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

	result, formatErr := formatPoliciesOutput(cmd, policies)

	if formatErr == nil {
		t.Fatal("formatPoliciesOutput() error = nil, want an error")
	}
	if result != "" {
		t.Errorf("formatPoliciesOutput() = %q, want an empty string", result)
	}
	if !strings.Contains(formatErr.Error(), "invalid format") {
		t.Errorf("formatPoliciesOutput() error = %q, want invalid format",
			formatErr.Error())
	}
	if !strings.Contains(formatErr.Error(), "xml") {
		t.Errorf("formatPoliciesOutput() error = %q, want it to name xml",
			formatErr.Error())
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
	result, formatErr := formatPoliciesOutput(cmd, policies)
	if formatErr != nil {
		t.Fatalf("formatPoliciesOutput() error = %v, want nil", formatErr)
	}

	normalized := normalizePolicyOutput(result)

	// Check header
	if !strings.Contains(result, "POLICIES") {
		t.Error("Human format should contain 'POLICIES' header")
	}

	// The human format prints only the name; policies are keyed by name
	// and the vestigial ID is not shown.
	expectedFields := []string{
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
	result, formatErr := formatPoliciesOutput(cmd, policies)
	if formatErr != nil {
		t.Fatalf("formatPoliciesOutput() error = %v, want nil", formatErr)
	}

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
			result, formatErr := formatPolicy(cmd, nil)
			if formatErr != nil {
				t.Fatalf("formatPolicy() error = %v, want nil", formatErr)
			}

			if result != tt.expected {
				t.Errorf("formatPolicy(nil) = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatPolicy_InvalidFormat(t *testing.T) {
	cmd := createTestCommandWithFormat("xml")
	policy := &data.Policy{Name: "test"}

	result, formatErr := formatPolicy(cmd, policy)

	if formatErr == nil {
		t.Fatal("formatPolicy() error = nil, want an error")
	}
	if result != "" {
		t.Errorf("formatPolicy() = %q, want an empty string", result)
	}
	if !strings.Contains(formatErr.Error(), "invalid format") {
		t.Errorf("formatPolicy() error = %q, want invalid format",
			formatErr.Error())
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
	result, formatErr := formatPolicy(cmd, policy)
	if formatErr != nil {
		t.Fatalf("formatPolicy() error = %v, want nil", formatErr)
	}

	// Check header
	if !strings.Contains(result, "POLICY DETAILS") {
		t.Error("Human format should contain 'POLICY DETAILS' header")
	}

	// Check all fields are present
	expectedFields := []string{
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
	result, formatErr := formatPolicy(cmd, policy)
	if formatErr != nil {
		t.Fatalf("formatPolicy() error = %v, want nil", formatErr)
	}

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
	result, formatErr := formatPoliciesOutput(cmd, policies)
	if formatErr != nil {
		t.Fatalf("formatPoliciesOutput() error = %v, want nil", formatErr)
	}

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

func TestFormatPoliciesOutput_YAMLFormat(t *testing.T) {
	policies := &[]data.PolicyListItem{
		{
			ID:   "123e4567-e89b-12d3-a456-426614174000",
			Name: "test-policy",
		},
	}

	tests := []struct {
		name   string
		format string
	}{
		{"yaml full name", "yaml"},
		{"yaml alias y", "y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommandWithFormat(tt.format)
			result, formatErr := formatPoliciesOutput(cmd, policies)
			if formatErr != nil {
				t.Fatalf("formatPoliciesOutput() error = %v, want nil", formatErr)
			}

			// Check that output contains YAML-like content
			if !strings.Contains(result, "id:") ||
				!strings.Contains(result, "name:") {
				t.Errorf("YAML format should contain YAML fields")
			}
		})
	}
}

func TestFormatPoliciesOutput_FormatAliases(t *testing.T) {
	policies := &[]data.PolicyListItem{
		{
			ID:   "test-id",
			Name: "test-name",
		},
	}

	tests := []struct {
		name           string
		format         string
		shouldContain  string
		shouldNotError bool
	}{
		{"human alias h", "h", "POLICIES", true},
		{"human alias plain", "plain", "POLICIES", true},
		{"human alias p", "p", "POLICIES", true},
		{"json alias j", "j", `"id"`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommandWithFormat(tt.format)
			result, formatErr := formatPoliciesOutput(cmd, policies)

			if formatErr != nil && tt.shouldNotError {
				t.Errorf("Format alias %q should not produce an error: %v",
					tt.format, formatErr)
			}

			if !strings.Contains(result, tt.shouldContain) {
				t.Errorf("Format %q output should contain %q, got: %s",
					tt.format, tt.shouldContain, result)
			}
		})
	}
}

func TestFormatPolicy_YAMLFormat(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	policy := &data.Policy{
		ID:              "123e4567-e89b-12d3-a456-426614174000",
		Name:            "test-policy",
		SPIFFEIDPattern: "^spiffe://example\\.org/.*$",
		PathPattern:     "^secrets/.*$",
		Permissions:     []data.PolicyPermission{"read"},
		CreatedAt:       createdAt,
	}

	tests := []struct {
		name   string
		format string
	}{
		{"yaml full name", "yaml"},
		{"yaml alias y", "y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommandWithFormat(tt.format)
			result, formatErr := formatPolicy(cmd, policy)
			if formatErr != nil {
				t.Fatalf("formatPolicy() error = %v, want nil", formatErr)
			}

			// Check that output contains YAML-like content
			if !strings.Contains(result, "id:") ||
				!strings.Contains(result, "name:") {
				t.Errorf("YAML format should contain YAML fields")
			}
		})
	}
}
