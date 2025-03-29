//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import (
	"os"
	"testing"
)

func TestIsPilot(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		spiffeid   string
		want       bool
	}{
		{
			name:       "default valid spiffeid",
			beforeTest: nil,
			spiffeid:   "spiffe://spike.ist/spike/pilot/role/superuser",
			want:       true,
		},
		{
			name:       "default invalid spiffeid",
			beforeTest: nil,
			spiffeid:   "spiffe://test/spike/pilot/role/superuser",
			want:       false,
		},
		{
			name: "custom valid spiffeid",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			spiffeid: "spiffe://corp.com/spike/pilot/role/superuser",
			want:     true,
		},
		{
			name: "custom invalid spiffeid",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			spiffeid: "spiffe://invalid/spike/pilot/role/superuser",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := IsPilot(tt.spiffeid); got != tt.want {
				t.Errorf("IsPilot() = %v, want %v", got, tt.want)
			}
		})
		if err := os.Unsetenv("SPIKE_TRUST_ROOT"); err != nil {
			panic("failed to unset env SPIKE_TRUST_ROOT")
		}
	}
}

func TestIsKeeper(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		spiffeid   string
		want       bool
	}{
		{
			name:       "default valid spiffeid",
			beforeTest: nil,
			spiffeid:   "spiffe://spike.ist/spike/keeper",
			want:       true,
		},
		{
			name:       "default invalid spiffeid",
			beforeTest: nil,
			spiffeid:   "spiffe://test/spike/keeper",
			want:       false,
		},
		{
			name: "custom valid spiffeid",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			spiffeid: "spiffe://corp.com/spike/keeper",
			want:     true,
		},
		{
			name: "custom invalid spiffeid",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			spiffeid: "spiffe://invalid/spike/keeper",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := IsKeeper(tt.spiffeid); got != tt.want {
				t.Errorf("IsKeeper() = %v, want %v", got, tt.want)
			}
		})
		if err := os.Unsetenv("SPIKE_TRUST_ROOT"); err != nil {
			panic("failed to unset env SPIKE_TRUST_ROOT")
		}
	}
}

func TestIsNexus(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		spiffeid   string
		want       bool
	}{
		{
			name:       "default valid spiffeid",
			beforeTest: nil,
			spiffeid:   "spiffe://spike.ist/spike/nexus",
			want:       true,
		},
		{
			name:       "default invalid spiffeid",
			beforeTest: nil,
			spiffeid:   "spiffe://test/spike/nexus",
			want:       false,
		},
		{
			name: "custom valid spiffeid",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			spiffeid: "spiffe://corp.com/spike/nexus",
			want:     true,
		},
		{
			name: "custom invalid spiffeid",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			spiffeid: "spiffe://invalid/spike/nexus",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := IsNexus(tt.spiffeid); got != tt.want {
				t.Errorf("IsNexus() = %v, want %v", got, tt.want)
			}
		})
		if err := os.Unsetenv("SPIKE_TRUST_ROOT"); err != nil {
			panic("failed to unset env SPIKE_TRUST_ROOT")
		}
	}
}

func TestCanTalkToAnyone(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{
			name: "default",
			in:   "",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PeerCanTalkToAnyone(tt.in); got != tt.want {
				t.Errorf("PeerCanTalkToAnyone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanTalkToKeeper(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		spiffeid   string
		want       bool
	}{
		{
			name:       "default nexus spiffe id",
			beforeTest: nil,
			spiffeid:   "spiffe://spike.ist/spike/nexus",
			want:       true,
		},
		{
			name:       "default keeper spiffe id",
			beforeTest: nil,
			spiffeid:   "spiffe://spike.ist/spike/keeper",
			// Keepers cannot talk to keepers.
			want: false,
		},
		{
			name: "custom nexus spiffe id",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			spiffeid: "spiffe://corp.com/spike/nexus",
			want:     true,
		},
		{
			name: "custom keeper spiffe id",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			spiffeid: "spiffe://corp.com/spike/keeper",
			// Keepers cannot talk to keepers; only Nexus can talk to Keepers.
			want: false,
		},
		{
			name:       "pilot spiffe id",
			beforeTest: nil,
			spiffeid:   "spiffe://spike.ist/spike/pilot/role/superuser",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := PeerCanTalkToKeeper(tt.spiffeid); got != tt.want {
				t.Errorf("PeerCanTalkToKeeper() = %v, want %v", got, tt.want)
			}
		})
		if err := os.Unsetenv("SPIKE_TRUST_ROOT"); err != nil {
			panic("failed to unset env SPIKE_TRUST_ROOT")
		}
	}
}
