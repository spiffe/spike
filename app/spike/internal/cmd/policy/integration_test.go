//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPolicySpecValidation tests the Spec struct validation
func TestPolicySpecValidation(t *testing.T) {
	tests := []struct {
		name   string
		policy Spec
		valid  bool
	}{
		{
			name: "valid_policy_spec",
			policy: Spec{
				Name:        "test-policy",
				SpiffeID:    "spiffe://example\\.org/test/.*",
				Path:        "secrets/.*",
				Permissions: []string{"read", "write"},
			},
			valid: true,
		},
		{
			name: "empty_name",
			policy: Spec{
				Name:        "",
				SpiffeID:    "spiffe://example\\.org/test/.*",
				Path:        "secrets/.*",
				Permissions: []string{"read"},
			},
			valid: false,
		},
		{
			name: "empty_spiffe_id",
			policy: Spec{
				Name:        "test-policy",
				SpiffeID:    "",
				Path:        "secrets/.*",
				Permissions: []string{"read"},
			},
			valid: false,
		},
		{
			name: "empty_path",
			policy: Spec{
				Name:        "test-policy",
				SpiffeID:    "spiffe://example\\.org/test/.*",
				Path:        "",
				Permissions: []string{"read"},
			},
			valid: false,
		},
		{
			name: "empty_permissions",
			policy: Spec{
				Name:        "test-policy",
				SpiffeID:    "spiffe://example\\.org/test/.*",
				Path:        "secrets/.*",
				Permissions: []string{},
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
					permsStr += perm
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
			yamlContent += "spiffeid: " + tt.policy.SpiffeID + "\n"
			yamlContent += "path: " + tt.policy.Path + "\n"
			yamlContent += "permissions:\n"
			for _, perm := range tt.policy.Permissions {
				yamlContent += "  - " + perm + "\n"
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
		expectValue Spec
	}{
		{
			name: "quoted_values",
			yamlContent: `name: "my-policy"
spiffeid: "spiffe://example\.org/quoted/.*"
path: "secrets/quoted/.*"
permissions:
  - "read"
  - "write"`,
			expectError: false,
			expectValue: Spec{
				Name:        "my-policy",
				SpiffeID:    "spiffe://example\\.org/quoted/.*",
				Path:        "secrets/quoted/.*",
				Permissions: []string{"read", "write"},
			},
		},
		{
			name: "permissions_as_flow_sequence",
			yamlContent: `name: flow-policy
spiffeid: spiffe://example\.org/flow/.*
path: secrets/flow/.*
permissions: [read, write, list]`,
			expectError: false,
			expectValue: Spec{
				Name:        "flow-policy",
				SpiffeID:    "spiffe://example\\.org/flow/.*",
				Path:        "secrets/flow/.*",
				Permissions: []string{"read", "write", "list"},
			},
		},
		{
			name: "multiline_spiffe_id",
			yamlContent: `name: multiline-policy
spiffeid: >
  spiffe://example\.org/multiline/.*
path: secrets/multiline/.*
permissions:
  - read`,
			expectError: false,
			expectValue: Spec{
				Name:        "multiline-policy",
				SpiffeID:    "spiffe://example\\.org/multiline/.*\n",
				Path:        "secrets/multiline/.*",
				Permissions: []string{"read"},
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
			if len(policy.SpiffeID) > 0 && policy.SpiffeID[len(policy.SpiffeID)-1] == '\n' {
				policy.SpiffeID = policy.SpiffeID[:len(policy.SpiffeID)-1]
			}
			if len(tt.expectValue.SpiffeID) > 0 &&
				tt.expectValue.SpiffeID[len(tt.expectValue.SpiffeID)-1] == '\n' {
				tt.expectValue.SpiffeID = tt.expectValue.SpiffeID[:len(tt.expectValue.SpiffeID)-1]
			}

			if policy.Name != tt.expectValue.Name {
				t.Errorf("Name = %v, want %v", policy.Name, tt.expectValue.Name)
			}
			if policy.SpiffeID != tt.expectValue.SpiffeID {
				t.Errorf("SpiffeID = %v, want %v",
					policy.SpiffeID, tt.expectValue.SpiffeID)
			}
			if policy.Path != tt.expectValue.Path {
				t.Errorf("Path = %v, want %v", policy.Path, tt.expectValue.Path)
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
