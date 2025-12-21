//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/validation"
)

func TestValidPermission(t *testing.T) {
	tests := []struct {
		name     string
		perm     string
		expected bool
	}{
		{"valid read", "read", true},
		{"valid write", "write", true},
		{"valid list", "list", true},
		{"valid execute", "execute", true},
		{"valid super", "super", true},
		{"invalid permission", "delete", false},
		{"empty permission", "", false},
		{"uppercase permission", "READ", false},
		{"mixed case permission", "Read", false},
		{"permission with spaces", " read ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidPermission(tt.perm)
			if result != tt.expected {
				t.Errorf("validPermission(%q) = %v, want %v",
					tt.perm, result, tt.expected)
			}
		})
	}
}

func TestValidPermissionsList(t *testing.T) {
	result := validation.ValidPermissionsList()

	expectedPerms := []data.PolicyPermission{
		data.PermissionRead,
		data.PermissionWrite,
		data.PermissionList,
		data.PermissionExecute,
		data.PermissionSuper,
	}

	for _, p := range expectedPerms {
		if !contains(result, string(p)) {
			t.Errorf(
				"ValidPermissionsList missing permission %q",
				p,
			)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPerms   []data.PolicyPermission
		wantErr     bool
		errContains string
	}{
		{
			name:      "single valid permission",
			input:     "read",
			wantPerms: []data.PolicyPermission{data.PermissionRead},
			wantErr:   false,
		},
		{
			name:  "multiple valid permissions",
			input: "read,write,list",
			wantPerms: []data.PolicyPermission{
				data.PermissionRead,
				data.PermissionWrite,
				data.PermissionList,
			},
			wantErr: false,
		},
		{
			name:  "permissions with spaces",
			input: "read, write, list",
			wantPerms: []data.PolicyPermission{
				data.PermissionRead,
				data.PermissionWrite,
				data.PermissionList,
			},
			wantErr: false,
		},
		{
			name:      "super permission",
			input:     "super",
			wantPerms: []data.PolicyPermission{data.PermissionSuper},
			wantErr:   false,
		},
		{
			name:        "invalid permission",
			input:       "invalid",
			wantErr:     true,
			errContains: "invalid permission",
		},
		{
			name:        "mixed valid and invalid",
			input:       "read,invalid,write",
			wantErr:     true,
			errContains: "invalid permission",
		},
		{
			name:        "empty string",
			input:       "",
			wantErr:     true,
			errContains: "no valid permissions",
		},
		{
			name:        "only commas",
			input:       ",,,",
			wantErr:     true,
			errContains: "no valid permissions",
		},
		{
			name:        "only spaces",
			input:       "   ",
			wantErr:     true,
			errContains: "no valid permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms, err := validation.ValidatePermissions(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePermissions(%q) expected error, got nil",
						tt.input)
					return
				}
				if tt.errContains != "" &&
					!containsSubstring(err.Msg, tt.errContains) {
					t.Errorf("ValidatePermissions(%q) error = %q, "+
						"want error containing %q",
						tt.input, err.Msg, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidatePermissions(%q) unexpected error: %v",
					tt.input, err)
				return
			}

			if len(perms) != len(tt.wantPerms) {
				t.Errorf("ValidatePermissions(%q) got %d permissions, want %d",
					tt.input, len(perms), len(tt.wantPerms))
				return
			}

			for i, p := range perms {
				if p != tt.wantPerms[i] {
					t.Errorf("ValidatePermissions(%q)[%d] = %q, want %q",
						tt.input, i, p, tt.wantPerms[i])
				}
			}
		})
	}
}

func TestValidateDataDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		dir         string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty path",
			dir:         "",
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "root path",
			dir:         "/",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:        "etc path",
			dir:         "/etc",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:        "etc subdirectory",
			dir:         "/etc/spike",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:        "sys path",
			dir:         "/sys",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:        "proc path",
			dir:         "/proc",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:        "dev path",
			dir:         "/dev",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:        "bin path",
			dir:         "/bin",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:        "usr path",
			dir:         "/usr",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:        "boot path",
			dir:         "/boot",
			wantErr:     true,
			errContains: "restricted",
		},
		{
			name:    "valid temp directory",
			dir:     tempDir,
			wantErr: false,
		},
		{
			name:    "valid subdirectory of temp",
			dir:     filepath.Join(tempDir, "subdir"),
			wantErr: false,
		},
		{
			name:        "non-existent parent",
			dir:         "/nonexistent/path/to/dir",
			wantErr:     true,
			errContains: "parent directory does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDataDirectory(tt.dir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDataDirectory(%q) expected error, got nil",
						tt.dir)
					return
				}
				if tt.errContains != "" &&
					!containsSubstring(err.Msg, tt.errContains) {
					t.Errorf("validateDataDirectory(%q) error = %q, "+
						"want error containing %q",
						tt.dir, err.Msg, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("validateDataDirectory(%q) unexpected error: %v",
					tt.dir, err)
			}
		})
	}
}

func TestValidateDataDirectory_FileNotDirectory(t *testing.T) {
	// Create a temporary file (not a directory)
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "testfile")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := validateDataDirectory(tempFile)
	if err == nil {
		t.Error("validateDataDirectory() expected error for file path, got nil")
		return
	}
	if !containsSubstring(err.Msg, "not a directory") {
		t.Errorf("validateDataDirectory() error = %q, "+
			"want error containing 'not a directory'", err.Msg)
	}
}

func TestVersionConstants(t *testing.T) {
	// Verify that version constants are set (not empty)
	if NexusVersion == "" {
		t.Error("NexusVersion should not be empty")
	}
	if PilotVersion == "" {
		t.Error("PilotVersion should not be empty")
	}
	if KeeperVersion == "" {
		t.Error("KeeperVersion should not be empty")
	}
	if BootstrapVersion == "" {
		t.Error("BootstrapVersion should not be empty")
	}

	// Verify all versions are the same (from app.Version)
	if NexusVersion != PilotVersion ||
		PilotVersion != KeeperVersion ||
		KeeperVersion != BootstrapVersion {
		t.Error("All version constants should be equal")
	}
}

func TestValidPermissions_AllPermissionsPresent(t *testing.T) {
	// Verify ValidPermissions contains all expected permissions
	expectedPerms := []data.PolicyPermission{
		data.PermissionRead,
		data.PermissionWrite,
		data.PermissionList,
		data.PermissionExecute,
		data.PermissionSuper,
	}

	validPermsList := validation.ValidPermissionsList()

	for _, expected := range expectedPerms {
		if !validation.ValidPermission(string(expected)) {
			t.Errorf("ValidPermissions missing %q", expected)
		}
	}

	permStrings := strings.Split(validPermsList, ", ")
	if len(permStrings) != len(expectedPerms) {
		t.Errorf("ValidPermissionsList has %d items, want %d.",
			len(permStrings), len(expectedPerms))
	}
}

func TestRestrictedPaths(t *testing.T) {
	// Verify restrictedPaths contains critical system directories
	expectedPaths := []string{"/", "/etc", "/sys", "/proc", "/dev", "/bin",
		"/sbin", "/usr", "/lib", "/lib64", "/boot", "/root"}

	for _, expected := range expectedPaths {
		found := false
		for _, actual := range restrictedPaths {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("restrictedPaths missing %q", expected)
		}
	}
}
