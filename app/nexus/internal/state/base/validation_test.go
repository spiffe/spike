//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/validation"
)

func TestVerifyPermissions_SuperPermissionJoker(t *testing.T) {
	// Test that super permission acts as a joker and grants all permissions
	testCases := []struct {
		name     string
		haves    []data.PolicyPermission
		wants    []data.PolicyPermission
		expected bool
	}{
		{
			name:     "super grants read",
			haves:    []data.PolicyPermission{data.PermissionSuper},
			wants:    []data.PolicyPermission{data.PermissionRead},
			expected: true,
		},
		{
			name:     "super grants write",
			haves:    []data.PolicyPermission{data.PermissionSuper},
			wants:    []data.PolicyPermission{data.PermissionWrite},
			expected: true,
		},
		{
			name:     "super grants list",
			haves:    []data.PolicyPermission{data.PermissionSuper},
			wants:    []data.PolicyPermission{data.PermissionList},
			expected: true,
		},
		{
			name:     "super grants multiple permissions",
			haves:    []data.PolicyPermission{data.PermissionSuper},
			wants:    []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionList},
			expected: true,
		},
		{
			name:     "super among other permissions",
			haves:    []data.PolicyPermission{data.PermissionRead, data.PermissionSuper, data.PermissionWrite},
			wants:    []data.PolicyPermission{data.PermissionList},
			expected: true,
		},
		{
			name:     "super grants empty wants",
			haves:    []data.PolicyPermission{data.PermissionSuper},
			wants:    []data.PolicyPermission{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validation.ValidatePolicyPermissions(tc.haves, tc.wants)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for case: %s", tc.expected, result, tc.name)
			}
		})
	}
}

func TestVerifyPermissions_SpecificPermissions(t *testing.T) {
	// Test normal permission checking (without `super`)
	testCases := []struct {
		name     string
		haves    []data.PolicyPermission
		wants    []data.PolicyPermission
		expected bool
	}{
		{
			name:     "has exact permission",
			haves:    []data.PolicyPermission{data.PermissionRead},
			wants:    []data.PolicyPermission{data.PermissionRead},
			expected: true,
		},
		{
			name:     "has all required permissions",
			haves:    []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionList},
			wants:    []data.PolicyPermission{data.PermissionRead, data.PermissionWrite},
			expected: true,
		},
		{
			name:     "missing one permission",
			haves:    []data.PolicyPermission{data.PermissionRead, data.PermissionWrite},
			wants:    []data.PolicyPermission{data.PermissionRead, data.PermissionList},
			expected: false,
		},
		{
			name:     "missing all permissions",
			haves:    []data.PolicyPermission{data.PermissionRead},
			wants:    []data.PolicyPermission{data.PermissionWrite, data.PermissionList},
			expected: false,
		},
		{
			name:     "empty haves, non-empty wants",
			haves:    []data.PolicyPermission{},
			wants:    []data.PolicyPermission{data.PermissionRead},
			expected: false,
		},
		{
			name:     "non-empty haves, empty wants",
			haves:    []data.PolicyPermission{data.PermissionRead, data.PermissionWrite},
			wants:    []data.PolicyPermission{},
			expected: true,
		},
		{
			name:     "both empty",
			haves:    []data.PolicyPermission{},
			wants:    []data.PolicyPermission{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validation.ValidatePolicyPermissions(tc.haves, tc.wants)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for case: %s", tc.expected, result, tc.name)
			}
		})
	}
}

func TestVerifyPermissions_SuperWithOtherPermissions(t *testing.T) {
	// Test edge cases where `super` is combined with other permissions
	testCases := []struct {
		name     string
		haves    []data.PolicyPermission
		wants    []data.PolicyPermission
		expected bool
	}{
		{
			name:     "super and read, wants write",
			haves:    []data.PolicyPermission{data.PermissionSuper, data.PermissionRead},
			wants:    []data.PolicyPermission{data.PermissionWrite},
			expected: true, // `super` should grant `write` even though we don't explicitly have it
		},
		{
			name:     "read and super, wants multiple",
			haves:    []data.PolicyPermission{data.PermissionRead, data.PermissionSuper},
			wants:    []data.PolicyPermission{data.PermissionWrite, data.PermissionList},
			expected: true, // `super` should grant all
		},
		{
			name:     "multiple permissions including super",
			haves:    []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionSuper, data.PermissionList},
			wants:    []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionList},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validation.ValidatePolicyPermissions(tc.haves, tc.wants)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for case: %s", tc.expected, result, tc.name)
			}
		})
	}
}

func BenchmarkVerifyPermissions_WithSuper(b *testing.B) {
	haves := []data.PolicyPermission{data.PermissionSuper}
	wants := []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionList}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validation.ValidatePolicyPermissions(haves, wants)
	}
}

func BenchmarkVerifyPermissions_WithoutSuper(b *testing.B) {
	haves := []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionList}
	wants := []data.PolicyPermission{data.PermissionRead, data.PermissionWrite}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validation.ValidatePolicyPermissions(haves, wants)
	}
}

func BenchmarkVerifyPermissions_LargePermissionSet(b *testing.B) {
	// Test with a larger set of permissions to see performance impact
	haves := []data.PolicyPermission{
		data.PermissionRead, data.PermissionWrite, data.PermissionList,
		data.PermissionRead, data.PermissionWrite, data.PermissionList, // duplicates to make it larger
		data.PermissionRead, data.PermissionWrite, data.PermissionList,
		data.PermissionSuper, // super at the end to test worst-case for the joker check
	}
	wants := []data.PolicyPermission{data.PermissionRead, data.PermissionWrite}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validation.ValidatePolicyPermissions(haves, wants)
	}
}
