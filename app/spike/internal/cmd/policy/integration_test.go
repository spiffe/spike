//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// TestPolicySpecValidation tests the Spec struct validation
func TestPolicySpecValidation(t *testing.T) {
	tests := []struct {
		name   string
		policy data.PolicySpec
		valid  bool
	}{
		{
			name: "valid_policy_spec",
			policy: data.PolicySpec{
				Name:            "test-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
				PathPattern:     "^secrets/.*$",
				Permissions:     []data.PolicyPermission{"read", "write"},
			},
			valid: true,
		},
		{
			name: "empty_name",
			policy: data.PolicySpec{
				Name:            "",
				SpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
				PathPattern:     "^secrets/.*$",
				Permissions:     []data.PolicyPermission{"read"},
			},
			valid: false,
		},
		{
			name: "empty_spiffe_id",
			policy: data.PolicySpec{
				Name:            "test-policy",
				SpiffeIDPattern: "",
				PathPattern:     "secrets/.*",
				Permissions:     []data.PolicyPermission{"read"},
			},
			valid: false,
		},
		{
			name: "empty_path",
			policy: data.PolicySpec{
				Name:            "test-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
				PathPattern:     "",
				Permissions:     []data.PolicyPermission{"read"},
			},
			valid: false,
		},
		{
			name: "empty_permissions",
			policy: data.PolicySpec{
				Name:            "test-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
				PathPattern:     "^secrets/.*$",
				Permissions:     []data.PolicyPermission{},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert permissions slice to comma-separated string
			permsStr := ""
			if len(tt.policy.Permissions) > 0 {
				for i, perm := range tt.policy.Permissions {
					if i > 0 {
						permsStr += ","
					}
					permsStr += string(perm)
				}
			}

			// Test using readPolicyFromFile validation logic
			// Create a temporary file with this policy
			tempDir, err := os.MkdirTemp("", "spike-validation-test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer func(path string) {
				err := os.RemoveAll(path)
				if err != nil {
					t.Logf("Failed to remove temp directory: %v", err)
				}
			}(tempDir)

			yamlContent := "name: " + tt.policy.Name + "\n"
			yamlContent += "spiffeid: " + tt.policy.SpiffeIDPattern + "\n"
			yamlContent += "path: " + tt.policy.PathPattern + "\n"
			yamlContent += "permissions:\n"
			for _, perm := range tt.policy.Permissions {
				yamlContent += "  - " + string(perm) + "\n"
			}

			filePath := filepath.Join(tempDir, "test-policy.yaml")
			err = os.WriteFile(filePath, []byte(yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			_, err = readPolicyFromFile(filePath)
			isValid := err == nil

			if tt.valid && !isValid {
				t.Errorf("Expected policy to be valid, but got error: %v", err)
			}
			if !tt.valid && isValid {
				t.Errorf("Expected policy to be invalid, but it was accepted")
			}
		})
	}
}

// TestYAMLParsingEdgeCases tests various YAML parsing edge cases
func TestYAMLParsingEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "spike-yaml-edge-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}(tempDir)

	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		expectValue data.PolicySpec
	}{
		{
			name: "quoted_values",
			yamlContent: `name: "my-policy"
spiffeidPattern: "^spiffe://example\\.org/quoted/.*$"
pathPattern: "^secrets/quoted/.*$"
permissions:
  - "read"
  - "write"`,
			expectError: false,
			expectValue: data.PolicySpec{
				Name:            "my-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/quoted/.*$",
				PathPattern:     "^secrets/quoted/.*$",
				Permissions:     []data.PolicyPermission{"read", "write"},
			},
		},
		{
			name: "permissions_as_flow_sequence",
			yamlContent: `name: flow-policy
spiffeidPattern: ^spiffe://example\.org/flow/.*$
pathPattern: ^secrets/flow/.*$
permissions: [read, write, list]`,
			expectError: false,
			expectValue: data.PolicySpec{
				Name:            "flow-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/flow/.*$",
				PathPattern:     "^secrets/flow/.*$",
				Permissions:     []data.PolicyPermission{"read", "write", "list"},
			},
		},
		{
			name: "multiline_spiffe_id",
			yamlContent: `name: multiline-policy
spiffeidPattern: >
  ^spiffe://example\.org/multiline/.*$
pathPattern: ^secrets/multiline/.*$
permissions:
  - read`,
			expectError: false,
			expectValue: data.PolicySpec{
				Name:            "multiline-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/multiline/.*$\n",
				PathPattern:     "^secrets/multiline/.*$",
				Permissions:     []data.PolicyPermission{"read"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tt.name+".yaml")
			err := os.WriteFile(filePath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			policy, err := readPolicyFromFile(filePath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Trim any trailing whitespace from SpiffeID for comparison
			if len(policy.SpiffeIDPattern) > 0 && policy.SpiffeIDPattern[len(policy.SpiffeIDPattern)-1] == '\n' {
				policy.SpiffeIDPattern = policy.SpiffeIDPattern[:len(policy.SpiffeIDPattern)-1]
			}
			if len(tt.expectValue.SpiffeIDPattern) > 0 &&
				tt.expectValue.SpiffeIDPattern[len(tt.expectValue.SpiffeIDPattern)-1] == '\n' {
				tt.expectValue.SpiffeIDPattern = tt.expectValue.SpiffeIDPattern[:len(tt.expectValue.SpiffeIDPattern)-1]
			}

			if policy.Name != tt.expectValue.Name {
				t.Errorf("Name = %v, want %v", policy.Name, tt.expectValue.Name)
			}
			if policy.SpiffeIDPattern != tt.expectValue.SpiffeIDPattern {
				t.Errorf("SpiffeID = %v, want %v",
					policy.SpiffeIDPattern, tt.expectValue.SpiffeIDPattern)
			}
			if policy.PathPattern != tt.expectValue.PathPattern {
				t.Errorf("Path = %v, want %v", policy.PathPattern, tt.expectValue.PathPattern)
			}
			if len(policy.Permissions) != len(tt.expectValue.Permissions) {
				t.Errorf("Permissions length = %d, want %d",
					len(policy.Permissions), len(tt.expectValue.Permissions))
			} else {
				for i, perm := range policy.Permissions {
					if perm != tt.expectValue.Permissions[i] {
						t.Errorf("Permissions[%d] = %v, want %v",
							i, perm, tt.expectValue.Permissions[i])
					}
				}
			}
		})
	}
}
