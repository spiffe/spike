//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyPolicyFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "spike-apply-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}(tempDir)

	tests := []struct {
		name         string
		fileContent  string
		fileName     string
		expectedPath string
		wantErr      bool
		errContains  string
	}{
		{
			name: "policy_with_trailing_slash_path",
			fileContent: `name: test-policy
spiffeidPattern: ^spiffe://example\.org/test/.*$
pathPattern: ^secrets/database/production/.*$
permissions:
  - read
  - write`,
			fileName:     "trailing-slash.yaml",
			expectedPath: "^secrets/database/production/.*$",
			wantErr:      false,
		},
		{
			name: "policy_with_normalized_path",
			fileContent: `name: test-policy
spiffeidPattern: ^spiffe://example\.org/test/.*$
pathPattern: ^secrets/cache/redis$
permissions:
  - read`,
			fileName:     "normalized-path.yaml",
			expectedPath: "^secrets/cache/redis$",
			wantErr:      false,
		},
		{
			name: "policy_with_root_path",
			fileContent: `name: admin-policy
spiffeidPattern: ^spiffe://example\.org/admin/.*$
pathPattern: ^/$
permissions:
  - super`,
			fileName:     "root-path.yaml",
			expectedPath: "^/$",
			wantErr:      false,
		},
		{
			name: "policy_with_empty_path",
			fileContent: `name: invalid-policy
spiffeidPattern: ^spiffe://example\.org/test/.*$
pathPattern: ""
permissions:
  - read`,
			fileName:    "empty-path.yaml",
			wantErr:     true,
			errContains: "pathPattern is required",
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

			// Test reading the policy
			_, readErr := readPolicyFromFile(filePath)

			if tt.wantErr {
				if readErr == nil {
					t.Errorf("readPolicyFromFile() expected error but got none")
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
						t.Errorf(
							"readPolicyFromFile() expected error containing "+
								"'%s', got '%v'", tt.errContains, readErr)
					}
				}
				return
			}

			if readErr != nil {
				t.Errorf("readPolicyFromFile() unexpected error: %v", readErr)
				return
			}
		})
	}
}

func TestApplyPolicyFromFlags(t *testing.T) {
	tests := []struct {
		name         string
		inputName    string
		inputSpiffed string
		inputPath    string
		inputPerms   string
		expectedPath string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "flags_with_trailing_slash_path",
			inputName:    "test-policy",
			inputSpiffed: "^spiffe://example\\.org/test/.*$",
			inputPath:    "^secrets/database/production$",
			inputPerms:   "read,write",
			expectedPath: "^secrets/database/production$",
			wantErr:      false,
		},
		{
			name:         "flags_with_normalized_path",
			inputName:    "cache-policy",
			inputSpiffed: "^spiffe://example\\.org/cache/.*$",
			inputPath:    "^secrets/cache/redis$",
			inputPerms:   "read",
			expectedPath: "^secrets/cache/redis$",
			wantErr:      false,
		},
		//{
		//	name:         "flags_with_multiple_trailing_slashes",
		//	inputName:    "multi-slash-policy",
		//	inputSpiffed: "^spiffe://example\\.org/test/.*$",
		//	inputPathPattern:    "^secrets/test///$",
		//	inputPerms:   "read",
		//	expectedPath: "^secrets/test$",
		//	wantErr:      false,
		//},
		{
			name:         "flags_with_root_path",
			inputName:    "root-policy",
			inputSpiffed: "^spiffe://example\\.org/admin/.*$",
			inputPath:    "^/$",
			inputPerms:   "super",
			expectedPath: "^/$",
			wantErr:      false,
		},
		{
			name:         "missing_path_flag",
			inputName:    "incomplete-policy",
			inputSpiffed: "^spiffe://example\\.org/test/.*$",
			inputPath:    "",
			inputPerms:   "read",
			wantErr:      true,
			errContains:  "--path-pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getPolicyFromFlags(tt.inputName,
				tt.inputSpiffed, tt.inputPath, tt.inputPerms)

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

			if err != nil {
				t.Errorf("getPolicyFromFlags() unexpected error: %v", err)
				return
			}
		})
	}
}
