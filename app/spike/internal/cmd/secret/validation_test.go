//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import "testing"

func TestValidSecretPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Valid paths
		{"simple path", "secrets/db/password", true},
		{"single segment", "password", true},
		{"with dots", "secrets.db.password", true},
		{"with underscores", "secrets_db_password", true},
		{"with hyphens", "secrets-db-password", true},
		{"mixed separators", "secrets/db_password-v1", true},
		{"alphanumeric", "secret123", true},
		{"uppercase", "SECRETS/DB", true},
		{"mixed case", "Secrets/Database/Password", true},

		// Paths with allowed special characters
		{"with parentheses", "secrets/(dev)", true},
		{"with question mark", "secrets?", true},
		{"with plus", "secrets+extra", true},
		{"with asterisk", "secrets/*", true},
		{"with pipe", "secrets|backup", true},
		{"with brackets", "secrets[0]", true},
		{"with braces", "secrets{key}", true},
		{"with backslash", "secrets\\windows", true},

		// Invalid paths
		{"empty path", "", false},
		{"with spaces", "secrets db", false},
		{"with newline", "secrets\ndb", false},
		{"with tab", "secrets\tdb", false},
		{"with quotes", "secrets\"db\"", false},
		{"with single quotes", "secrets'db'", false},
		{"with backtick", "secrets`db`", false},
		{"with semicolon", "secrets;db", false},
		{"with ampersand", "secrets&db", false},
		{"with dollar", "secrets$db", false},
		{"with at sign", "secrets@db", false},
		{"with hash", "secrets#db", false},
		{"with percent", "secrets%db", false},
		{"with caret", "secrets^db", false},
		{"with exclamation", "secrets!db", false},
		{"with comma", "secrets,db", false},
		{"with less than", "secrets<db", false},
		{"with greater than", "secrets>db", false},
		{"with equals", "secrets=db", false},
		{"with colon", "secrets:db", false},
		{"only space", " ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validSecretPath(tt.path)
			if result != tt.expected {
				t.Errorf("validSecretPath(%q) = %v, want %v",
					tt.path, result, tt.expected)
			}
		})
	}
}

func TestValidSecretPath_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"leading slash", "/secrets/db", true},
		{"trailing slash", "secrets/db/", true},
		{"double slash", "secrets//db", true},
		{"only slash", "/", true},
		{"multiple slashes", "///", true},
		{"dots only", "...", true},
		{"relative path", "../secrets", true},
		{"current dir", "./secrets", true},
		{"very long path", "a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validSecretPath(tt.path)
			if result != tt.expected {
				t.Errorf("validSecretPath(%q) = %v, want %v",
					tt.path, result, tt.expected)
			}
		})
	}
}
