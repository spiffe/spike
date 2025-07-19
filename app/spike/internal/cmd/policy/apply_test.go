//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty_path",
			input:    "",
			expected: "",
		},
		{
			name:     "root_path_unchanged",
			input:    "/",
			expected: "/",
		},
		{
			name:     "simple_path_no_trailing_slash",
			input:    "secrets/database",
			expected: "secrets/database",
		},
		{
			name:     "simple_path_with_trailing_slash",
			input:    "secrets/database/",
			expected: "secrets/database",
		},
		{
			name:     "complex_path_with_trailing_slash",
			input:    "secrets/web-service/production/",
			expected: "secrets/web-service/production",
		},
		{
			name:     "complex_path_no_trailing_slash",
			input:    "secrets/web-service/production",
			expected: "secrets/web-service/production",
		},
		{
			name:     "path_with_multiple_trailing_slashes",
			input:    "secrets/database///",
			expected: "secrets/database",
		},
		{
			name:     "absolute_path_with_trailing_slash",
			input:    "/secrets/database/",
			expected: "/secrets/database",
		},
		{
			name:     "absolute_path_no_trailing_slash",
			input:    "/secrets/database",
			expected: "/secrets/database",
		},
		{
			name:     "single_segment_with_trailing_slash",
			input:    "secrets/",
			expected: "secrets",
		},
		{
			name:     "wildcard_path_with_trailing_slash",
			input:    "secrets/*/",
			expected: "secrets/*",
		},
		{
			name:     "deeply_nested_path",
			input:    "secrets/app/env/production/database/primary/",
			expected: "secrets/app/env/production/database/primary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

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
spiffeid: spiffe://example.org/test/*
path: secrets/database/production/
permissions:
  - read
  - write`,
			fileName:     "trailing-slash.yaml",
			expectedPath: "secrets/database/production",
			wantErr:      false,
		},
		{
			name: "policy_with_normalized_path",
			fileContent: `name: test-policy
spiffeid: spiffe://example.org/test/*
path: secrets/cache/redis
permissions:
  - read`,
			fileName:     "normalized-path.yaml",
			expectedPath: "secrets/cache/redis",
			wantErr:      false,
		},
		{
			name: "policy_with_root_path",
			fileContent: `name: admin-policy
spiffeid: spiffe://example.org/admin/*
path: /
permissions:
  - super`,
			fileName:     "root-path.yaml",
			expectedPath: "/",
			wantErr:      false,
		},
		{
			name: "policy_with_empty_path",
			fileContent: `name: invalid-policy
spiffeid: spiffe://example.org/test/*
path: ""
permissions:
  - read`,
			fileName:    "empty-path.yaml",
			wantErr:     true,
			errContains: "path is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test file:
			filePath := filepath.Join(tempDir, tt.fileName)
			err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test reading the policy
			policy, err := readPolicyFromFile(filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("readPolicyFromFile() expected error but got none")
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
						t.Errorf("readPolicyFromFile() expected error containing '%s', got '%v'", tt.errContains, err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("readPolicyFromFile() unexpected error: %v", err)
				return
			}

			// Test path normalization
			normalizedPath := normalizePath(policy.Path)
			if normalizedPath != tt.expectedPath {
				t.Errorf("normalizePath(%q) = %q, want %q", policy.Path, normalizedPath, tt.expectedPath)
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
			inputSpiffed: "spiffe://example.org/test/*",
			inputPath:    "secrets/database/production/",
			inputPerms:   "read,write",
			expectedPath: "secrets/database/production",
			wantErr:      false,
		},
		{
			name:         "flags_with_normalized_path",
			inputName:    "cache-policy",
			inputSpiffed: "spiffe://example.org/cache/*",
			inputPath:    "secrets/cache/redis",
			inputPerms:   "read",
			expectedPath: "secrets/cache/redis",
			wantErr:      false,
		},
		{
			name:         "flags_with_multiple_trailing_slashes",
			inputName:    "multi-slash-policy",
			inputSpiffed: "spiffe://example.org/test/*",
			inputPath:    "secrets/test///",
			inputPerms:   "read",
			expectedPath: "secrets/test",
			wantErr:      false,
		},
		{
			name:         "flags_with_root_path",
			inputName:    "root-policy",
			inputSpiffed: "spiffe://example.org/admin/*",
			inputPath:    "/",
			inputPerms:   "super",
			expectedPath: "/",
			wantErr:      false,
		},
		{
			name:         "missing_path_flag",
			inputName:    "incomplete-policy",
			inputSpiffed: "spiffe://example.org/test/*",
			inputPath:    "",
			inputPerms:   "read",
			wantErr:      true,
			errContains:  "--path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := getPolicyFromFlags(tt.inputName, tt.inputSpiffed, tt.inputPath, tt.inputPerms)

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
						t.Errorf("getPolicyFromFlags() expected error containing '%s', got '%v'", tt.errContains, err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("getPolicyFromFlags() unexpected error: %v", err)
				return
			}

			// Test path normalization
			normalizedPath := normalizePath(policy.Path)
			if normalizedPath != tt.expectedPath {
				t.Errorf("normalizePath(%q) = %q, want %q", policy.Path, normalizedPath, tt.expectedPath)
			}
		})
	}
}

func TestPathNormalizationIntegration(t *testing.T) {
	// Test that various path formats all normalize correctly
	pathTests := []struct {
		original   string
		normalized string
	}{
		{"secrets/database/password", "secrets/database/password"},
		{"secrets/database/password/", "secrets/database/password"},
		{"secrets/web-service/config/", "secrets/web-service/config"},
		{"secrets/", "secrets"},
		{"/secrets/", "/secrets"},
		{"/", "/"},
		{"app/env/production/db/", "app/env/production/db"},
		{"cache/redis/session///", "cache/redis/session"},
	}

	for _, test := range pathTests {
		t.Run("normalize_"+test.original, func(t *testing.T) {
			result := normalizePath(test.original)
			if result != test.normalized {
				t.Errorf("normalizePath(%q) = %q, want %q", test.original, result, test.normalized)
			}
		})
	}
}
