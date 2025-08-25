//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

func TestContains(t *testing.T) {
	permissions := []data.PolicyPermission{
		data.PermissionRead,
		data.PermissionWrite,
		data.PermissionList,
	}

	testCases := []struct {
		name       string
		permission data.PolicyPermission
		expected   bool
	}{
		{"contains read", data.PermissionRead, true},
		{"contains write", data.PermissionWrite, true},
		{"contains list", data.PermissionList, true},
		{"does not contain super", data.PermissionSuper, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := contains(permissions, tc.permission)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for permission %v", tc.expected, result, tc.permission)
			}
		})
	}
}

func TestContains_EmptySlice(t *testing.T) {
	var permissions []data.PolicyPermission

	result := contains(permissions, data.PermissionRead)
	if result {
		t.Error("Expected false for empty permission slice")
	}
}

func TestHasAllPermissions_SuperPermissionJoker(t *testing.T) {
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
			result := hasAllPermissions(tc.haves, tc.wants)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for case: %s", tc.expected, result, tc.name)
			}
		})
	}
}

func TestHasAllPermissions_SpecificPermissions(t *testing.T) {
	// Test normal permission checking (without super)
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
			result := hasAllPermissions(tc.haves, tc.wants)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for case: %s", tc.expected, result, tc.name)
			}
		})
	}
}

func TestHasAllPermissions_SuperWithOtherPermissions(t *testing.T) {
	// Test edge cases where super is combined with other permissions
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
			expected: true, // super should grant write even though we don't explicitly have it
		},
		{
			name:     "read and super, wants multiple",
			haves:    []data.PolicyPermission{data.PermissionRead, data.PermissionSuper},
			wants:    []data.PolicyPermission{data.PermissionWrite, data.PermissionList},
			expected: true, // super should grant all
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
			result := hasAllPermissions(tc.haves, tc.wants)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for case: %s", tc.expected, result, tc.name)
			}
		})
	}
}

// Benchmark tests
func BenchmarkContains(b *testing.B) {
	permissions := []data.PolicyPermission{
		data.PermissionRead,
		data.PermissionWrite,
		data.PermissionList,
		data.PermissionSuper,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contains(permissions, data.PermissionWrite)
	}
}

func BenchmarkHasAllPermissions_WithSuper(b *testing.B) {
	haves := []data.PolicyPermission{data.PermissionSuper}
	wants := []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionList}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasAllPermissions(haves, wants)
	}
}

func BenchmarkHasAllPermissions_WithoutSuper(b *testing.B) {
	haves := []data.PolicyPermission{data.PermissionRead, data.PermissionWrite, data.PermissionList}
	wants := []data.PolicyPermission{data.PermissionRead, data.PermissionWrite}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasAllPermissions(haves, wants)
	}
}

func BenchmarkHasAllPermissions_LargePermissionSet(b *testing.B) {
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
		hasAllPermissions(haves, wants)
	}
}
