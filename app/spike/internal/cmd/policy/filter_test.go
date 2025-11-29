//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import "testing"

func TestValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected bool
	}{
		// Valid UUIDs
		{
			name:     "valid lowercase UUID",
			uuid:     "123e4567-e89b-12d3-a456-426614174000",
			expected: true,
		},
		{
			name:     "valid uppercase UUID",
			uuid:     "123E4567-E89B-12D3-A456-426614174000",
			expected: true,
		},
		{
			name:     "valid mixed case UUID",
			uuid:     "123e4567-E89B-12d3-A456-426614174000",
			expected: true,
		},
		{
			name:     "all zeros UUID",
			uuid:     "00000000-0000-0000-0000-000000000000",
			expected: true,
		},
		{
			name:     "all fs UUID",
			uuid:     "ffffffff-ffff-ffff-ffff-ffffffffffff",
			expected: true,
		},

		// Invalid UUIDs
		{
			name:     "empty string",
			uuid:     "",
			expected: false,
		},
		{
			name:     "too short",
			uuid:     "123e4567-e89b-12d3-a456",
			expected: false,
		},
		{
			name:     "too long",
			uuid:     "123e4567-e89b-12d3-a456-426614174000-extra",
			expected: false,
		},
		{
			name:     "missing dashes",
			uuid:     "123e4567e89b12d3a456426614174000",
			expected: false,
		},
		{
			name:     "wrong dash positions",
			uuid:     "123e456-7e89b-12d3-a456-426614174000",
			expected: false,
		},
		{
			name:     "contains invalid characters",
			uuid:     "123e4567-e89b-12d3-a456-42661417400g",
			expected: false,
		},
		{
			name:     "contains spaces",
			uuid:     "123e4567 e89b 12d3 a456 426614174000",
			expected: false,
		},
		{
			name:     "random string",
			uuid:     "not-a-uuid-at-all",
			expected: false,
		},
		{
			name:     "policy name instead of UUID",
			uuid:     "my-policy-name",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validUUID(tt.uuid)
			if result != tt.expected {
				t.Errorf("validUUID(%q) = %v, want %v",
					tt.uuid, result, tt.expected)
			}
		})
	}
}
