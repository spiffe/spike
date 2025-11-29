//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

func TestReadPolicyFromFile(t *testing.T) {
	// Create a temporary directory for test files:
	tempDir, err := os.MkdirTemp("", "spike-policy-test")
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
		fileContent string
		fileName    string
		want        data.PolicySpec
		wantErr     bool
		errContains string
	}{
		{
			name: "valid_policy_file",
			fileContent: `name: test-policy
spiffeidPattern: ^spiffe://example\.org/test/.*$
pathPattern: ^secrets/.*$
permissions:
  - read
  - write`,
			fileName: "valid-policy.yaml",
			want: data.PolicySpec{
				Name:            "test-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
				PathPattern:     "^secrets/.*$",
				Permissions:     []data.PolicyPermission{"read", "write"},
			},
			wantErr: false,
		},
		{
			name: "valid_policy_with_all_permissions",
			fileContent: `name: full-access-policy
spiffeidPattern: ^spiffe://example\.org/admin/.*$
pathPattern: ^secrets/.*$
permissions:
  - read
  - write
  - list
  - super`,
			fileName: "full-policy.yaml",
			want: data.PolicySpec{
				Name:            "full-access-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/admin/.*$",
				PathPattern:     "^secrets/.*$",
				Permissions:     []data.PolicyPermission{"read", "write", "list", "super"},
			},
			wantErr: false,
		},
		{
			name: "missing_name",
			fileContent: `spiffeid: ^spiffe://example\\.org/test/.*$
pathPattern: ^secrets/*$
permissions:
  - read`,
			fileName:    "missing-name.yaml",
			wantErr:     true,
			errContains: "policy name is required",
		},
		{
			name: "missing_spiffeid",
			fileContent: `name: test-policy
pathPattern: secrets/*
permissions:
  - read`,
			fileName:    "missing-spiffeid.yaml",
			wantErr:     true,
			errContains: "spiffeidPattern is required",
		},
		{
			name: "missing_path",
			fileContent: `name: test-policy
spiffeidPattern: ^spiffe://example\.org/test/.*$
permissions:
  - read`,
			fileName:    "missing-path.yaml",
			wantErr:     true,
			errContains: "pathPattern is required",
		},
		{
			name: "missing_permissions",
			fileContent: `name: test-policy
spiffeidPattern: ^spiffe://example\.org/test/.*$
pathPattern: ^secrets/.*$`,
			fileName:    "missing-permissions.yaml",
			wantErr:     true,
			errContains: "permissions are required",
		},
		{
			name: "empty_permissions_list",
			fileContent: `name: test-policy
spiffeidPattern: ^spiffe://example\.org/test/.*$
pathPattern: ^secrets/.*$
permissions: []`,
			fileName:    "empty-permissions.yaml",
			wantErr:     true,
			errContains: "permissions are required",
		},
		{
			name: "invalid_yaml",
			fileContent: `name: test-policy
spiffeidPattern: ^spiffe://example\.org/test/.*$
pathPattern: ^secrets/.*$
permissions: [
  - read
  - write`, // Invalid YAML - missing closing bracket
			fileName:    "invalid-yaml.yaml",
			wantErr:     true,
			errContains: "failed to parse YAML file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test file:
			filePath := filepath.Join(tempDir, tt.fileName)
			if writeErr := os.WriteFile(
				filePath, []byte(tt.fileContent), 0644,
			); writeErr != nil {
				t.Fatalf("Failed to create test file: %v", writeErr)
			}

			// Test the function
			got, readErr := readPolicyFromFile(filePath)

			// Check error expectations
			if tt.wantErr {
				if readErr == nil {
					t.Errorf("readPolicyFromFile() expected error but got none")
					return
				}
				if tt.errContains != "" && readErr.Error() == "" {
					t.Errorf("readPolicyFromFile() "+
						"expected error containing '%s', got empty error",
						tt.errContains)
					return
				}
				if tt.errContains != "" {
					// Check if error contains expected substring
					found := false
					if readErr != nil && len(readErr.Error()) > 0 {
						errorStr := readErr.Error()
						if len(errorStr) >= len(tt.errContains) {
							for i := 0; i <= len(errorStr)-len(tt.errContains); i++ {
								if errorStr[i:i+len(tt.errContains)] == tt.errContains {
									found = true
									break
								}
							}
						}
					}
					if !found {
						t.Errorf("readPolicyFromFile() "+
							"expected error containing '%s', got '%v'",
							tt.errContains, readErr)
					}
				}
				return
			}

			// Check the success case:
			if readErr != nil {
				t.Errorf("readPolicyFromFile() unexpected error: %v", readErr)
				return
			}

			// Compare results
			if got.Name != tt.want.Name {
				t.Errorf("readPolicyFromFile() Name = %v, want %v",
					got.Name, tt.want.Name)
			}
			if got.SpiffeIDPattern != tt.want.SpiffeIDPattern {
				t.Errorf("readPolicyFromFile() SpiffeIDPattern = %v, want %v",
					got.SpiffeIDPattern, tt.want.SpiffeIDPattern)
			}
			if got.PathPattern != tt.want.PathPattern {
				t.Errorf("readPolicyFromFile() PathPattern = %v, want %v",
					got.PathPattern, tt.want.PathPattern)
			}
			if len(got.Permissions) != len(tt.want.Permissions) {
				t.Errorf("readPolicyFromFile() "+
					"Permissions length = %v, want %v",
					len(got.Permissions), len(tt.want.Permissions))
			} else {
				for i, perm := range got.Permissions {
					if perm != tt.want.Permissions[i] {
						t.Errorf("readPolicyFromFile() "+
							"Permissions[%d] = %v, want %v",
							i, perm, tt.want.Permissions[i])
					}
				}
			}
		})
	}
}

func TestReadPolicyFromFileNotFound(t *testing.T) {
	_, err := readPolicyFromFile("/nonexistent/file.yaml")
	if err == nil {
		t.Error("readPolicyFromFile() " +
			"expected error for non-existent file but got none")
	}
	if err != nil && len(err.Error()) > 0 {
		// Check if the error contains "does not exist":
		errorStr := err.Error()
		expected := "does not exist"
		found := false
		if len(errorStr) >= len(expected) {
			for i := 0; i <= len(errorStr)-len(expected); i++ {
				if errorStr[i:i+len(expected)] == expected {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("readPolicyFromFile() expected error "+
				"containing 'does not exist', got '%v'", err)
		}
	}
}

func TestGetPolicyFromFlags(t *testing.T) {
	tests := []struct {
		name                 string
		inputName            string
		inputSpiffeIDPattern string
		inputPathPattern     string
		inputPerms           string
		want                 data.PolicySpec
		wantErr              bool
		errContains          string
	}{
		{
			name:                 "valid_flags",
			inputName:            "test-policy",
			inputSpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
			inputPathPattern:     "^secrets/.*$",
			inputPerms:           "read,write",
			want: data.PolicySpec{
				Name:            "test-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
				PathPattern:     "^secrets/.*$",
				Permissions:     []data.PolicyPermission{"read", "write"},
			},
			wantErr: false,
		},
		{
			name:                 "valid_flags_with_spaces",
			inputName:            "test-policy",
			inputSpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
			inputPathPattern:     "^secrets/.*$",
			inputPerms:           "read, write, list",
			want: data.PolicySpec{
				Name:            "test-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
				PathPattern:     "^secrets/.*$",
				Permissions:     []data.PolicyPermission{"read", "write", "list"},
			},
			wantErr: false,
		},
		{
			name:                 "single_permission",
			inputName:            "read-only-policy",
			inputSpiffeIDPattern: "^spiffe://example\\.org/readonly/.*$",
			inputPathPattern:     "^secrets/readonly/.*$",
			inputPerms:           "read",
			want: data.PolicySpec{
				Name:            "read-only-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/readonly/.*$",
				PathPattern:     "^secrets/readonly/.*$",
				Permissions:     []data.PolicyPermission{"read"},
			},
			wantErr: false,
		},
		{
			name:                 "all_permissions",
			inputName:            "admin-policy",
			inputSpiffeIDPattern: "^spiffe://example\\.org/admin/.*$",
			inputPathPattern:     "^.*$",
			inputPerms:           "read,write,list,super",
			want: data.PolicySpec{
				Name:            "admin-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/admin/.*$",
				PathPattern:     "^.*$",
				Permissions:     []data.PolicyPermission{"read", "write", "list", "super"},
			},
			wantErr: false,
		},
		{
			name:                 "missing_name",
			inputName:            "",
			inputSpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
			inputPathPattern:     "^secrets/.*$",
			inputPerms:           "read",
			wantErr:              true,
			errContains:          "--name",
		},
		{
			name:                 "missing_spiffeid",
			inputName:            "test-policy",
			inputSpiffeIDPattern: "",
			inputPathPattern:     "^secrets/.*$",
			inputPerms:           "read",
			wantErr:              true,
			errContains:          "--spiffeid-pattern",
		},
		{
			name:                 "missing_path",
			inputName:            "test-policy",
			inputSpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
			inputPathPattern:     "",
			inputPerms:           "read",
			wantErr:              true,
			errContains:          "--path-pattern",
		},
		{
			name:                 "missing_permissions",
			inputName:            "test-policy",
			inputSpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
			inputPathPattern:     "^secrets/.*$",
			inputPerms:           "",
			wantErr:              true,
			errContains:          "--permissions",
		},
		{
			name:                 "multiple_missing_flags",
			inputName:            "",
			inputSpiffeIDPattern: "",
			inputPathPattern:     "^secrets/.*$",
			inputPerms:           "read",
			wantErr:              true,
			errContains:          "required flags are missing",
		},
		{
			name:                 "empty_permissions_after_split",
			inputName:            "test-policy",
			inputSpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
			inputPathPattern:     "^secrets/.*$",
			inputPerms:           ",,,",
			want: data.PolicySpec{
				Name:            "test-policy",
				SpiffeIDPattern: "^spiffe://example\\.org/test/.*$",
				PathPattern:     "^secrets/.*$",
				Permissions:     []data.PolicyPermission{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPolicyFromFlags(tt.inputName,
				tt.inputSpiffeIDPattern, tt.inputPathPattern, tt.inputPerms)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("getPolicyFromFlags() expected error but got none")
					return
				}
				if tt.errContains != "" {
					// Check if error contains expected substring
					found := false
					if err != nil && len(err.Error()) > 0 {
						errorStr := err.Error()
						if len(errorStr) >= len(tt.errContains) {
							for i := 0; i <= len(errorStr)-len(tt.errContains); i++ {
								if errorStr[i:i+len(tt.errContains)] == tt.errContains {
									found = true
									break
								}
							}
						}
					}
					if !found {
						t.Errorf("getPolicyFromFlags() "+
							"expected error containing '%s', got '%v'",
							tt.errContains, err)
					}
				}
				return
			}

			// Check the success case:
			if err != nil {
				t.Errorf("getPolicyFromFlags() unexpected error: %v", err)
				return
			}

			// Compare results
			if got.Name != tt.want.Name {
				t.Errorf("getPolicyFromFlags() Name = %v, want %v",
					got.Name, tt.want.Name)
			}
			if got.SpiffeIDPattern != tt.want.SpiffeIDPattern {
				t.Errorf("getPolicyFromFlags() SpiffeIDPattern = %v, want %v",
					got.SpiffeIDPattern, tt.want.SpiffeIDPattern)
			}
			if got.PathPattern != tt.want.PathPattern {
				t.Errorf("getPolicyFromFlags() PathPattern = %v, want %v",
					got.PathPattern, tt.want.PathPattern)
			}
			if len(got.Permissions) != len(tt.want.Permissions) {
				t.Errorf("getPolicyFromFlags() "+
					"Permissions length = %v, want %v",
					len(got.Permissions), len(tt.want.Permissions))
			} else {
				for i, perm := range got.Permissions {
					if perm != tt.want.Permissions[i] {
						t.Errorf("getPolicyFromFlags() "+
							"Permissions[%d] = %v, want %v",
							i, perm, tt.want.Permissions[i])
					}
				}
			}
		})
	}
}

func TestNewPolicyCreateCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newPolicyCreateCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	if cmd.Use != "create" {
		t.Errorf("Expected command use to be 'create', got '%s'", cmd.Use)
	}

	if cmd.Short != "Create a new policy" {
		t.Errorf("Expected command short description to be "+
			"'Create a new policy', got '%s'", cmd.Short)
	}

	// Check if all required flags are present (create command
	// only has flag-based input)
	expectedFlags := []string{"name", "path-pattern", "spiffeid-pattern", "permissions"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be present", flagName)
		}
	}
}

func TestPolicyCreateCommandFlagValidation(t *testing.T) {

	tests := []struct {
		name        string
		flags       map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "missing all flags",
			flags:       map[string]string{},
			expectError: true,
			errorMsg:    "required flags are missing",
		},
		{
			name: "missing name flag",
			flags: map[string]string{
				"path-pattern":     "secrets/database/production",
				"spiffeid-pattern": "^spiffe://example\\.org/service/.*$",
				"permissions":      "read,write",
			},
			expectError: true,
			errorMsg:    "required flags are missing: --name",
		},
		{
			name: "missing path flag",
			flags: map[string]string{
				"name":             "test-policy",
				"spiffeid-pattern": "^spiffe://example\\.org/service/.*$",
				"permissions":      "read,write",
			},
			expectError: true,
			errorMsg:    "required flags are missing: --path",
		},
		{
			name: "missing spiffeid flag",
			flags: map[string]string{
				"name":         "test-policy",
				"path-pattern": "^secrets/database/production$",
				"permissions":  "read,write",
			},
			expectError: true,
			errorMsg:    "required flags are missing: --spiffeid-pattern",
		},
		{
			name: "missing permissions flag",
			flags: map[string]string{
				"name":             "test-policy",
				"path-pattern":     "^secrets/database/production$",
				"spiffeid-pattern": "^spiffe://example\\.org/service/.*$",
			},
			expectError: true,
			errorMsg:    "required flags are missing: --permissions",
		},
		{
			name: "all flags present",
			flags: map[string]string{
				"name":             "test-policy",
				"path-pattern":     "^secrets/database/production$",
				"spiffeid-pattern": "^spiffe://example\\.org/service/.*$",
				"permissions":      "read,write",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := getPolicyFromFlags(
				tt.flags["name"],
				tt.flags["spiffeid-pattern"],
				tt.flags["path-pattern"],
				tt.flags["permissions"],
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message "+
						"to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
					return
				}
				if policy.Name != tt.flags["name"] {
					t.Errorf("Expected policy name to be '%s', got '%s'",
						tt.flags["name"], policy.Name)
				}
				if policy.PathPattern != tt.flags["path-pattern"] {
					t.Errorf("Expected policy path to be '%s', got '%s'",
						tt.flags["path-pattern"], policy.PathPattern)
				}
				if policy.SpiffeIDPattern != tt.flags["spiffeid-pattern"] {
					t.Errorf("Expected policy spiffeid to be '%s', got '%s'",
						tt.flags["spiffeid-pattern"], policy.SpiffeIDPattern)
				}
			}
		})
	}
}

// Test that the create command is registered properly
func TestPolicyCreateCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	policyCmd := NewCommand(source, SPIFFEIDPattern)

	var createCmd *cobra.Command
	for _, cmd := range policyCmd.Commands() {
		if cmd.Use == "create" {
			createCmd = cmd
			break
		}
	}

	if createCmd == nil {
		t.Error("Expected 'create' command to be registered")
	}
}
