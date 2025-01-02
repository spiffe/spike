//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import (
	"os"
	"testing"
)

func Test_trustRoot(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		want       string
	}{
		{
			name:       "default env",
			beforeTest: nil,
			want:       "spike.ist",
		},
		{
			name: "default env",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env for SPIKE_TRUST_ROOT")
				}
			},
			want: "corp.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := trustRoot(); got != tt.want {
				t.Errorf("trustRoot() = %v, want %v", got, tt.want)
			}
			if err := os.Unsetenv("SPIKE_TRUST_ROOT"); err != nil {
				panic("failed to unset env SPIKE_TRUST_ROOT")
			}
		})
	}
}

func TestSpikeKeeperSpiffeId(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		want       string
	}{
		{
			name:       "default trust root",
			beforeTest: nil,
			want:       "spiffe://spike.ist/spike/keeper",
		},
		{
			name: "custom trust root",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			want: "spiffe://corp.com/spike/keeper",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := SpikeKeeperSpiffeId(); got != tt.want {
				t.Errorf("SpikeKeeperSpiffeId() = %v, want %v", got, tt.want)
			}
			if err := os.Unsetenv("SPIKE_TRUST_ROOT"); err != nil {
				panic("failed to unset env SPIKE_TRUST_ROOT")
			}
		})
	}
}

func TestSpikeNexusSpiffeId(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		want       string
	}{
		{
			name:       "default trust root",
			beforeTest: nil,
			want:       "spiffe://spike.ist/spike/nexus",
		},
		{
			name: "custom trust root",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env SPIKE_TRUST_ROOT")
				}
			},
			want: "spiffe://corp.com/spike/nexus",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := SpikeNexusSpiffeId(); got != tt.want {
				t.Errorf("SpikeNexusSpiffeId() = %v, want %v", got, tt.want)
			}
			if err := os.Unsetenv("SPIKE_TRUST_ROOT"); err != nil {
				panic("failed to unset env SPIKE_TRUST_ROOT")
			}
		})
	}
}

func TestSpikePilotSpiffeId(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		want       string
	}{
		{
			name:       "default trust root",
			beforeTest: nil,
			want:       "spiffe://spike.ist/spike/pilot/role/superuser",
		},
		{
			name: "custom trust root",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_TRUST_ROOT", "corp.com"); err != nil {
					panic("failed to set env for SPIKE_TRUST_ROOT")
				}
			},
			want: "spiffe://corp.com/spike/pilot/role/superuser",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := SpikePilotSpiffeId(); got != tt.want {
				t.Errorf("SpikePilotSpiffeId() = %v, want %v", got, tt.want)
			}
			if err := os.Unsetenv("SPIKE_TRUST_ROOT"); err != nil {
				panic("failed to unset env SPIKE_TRUST_ROOT")
			}
		})
	}
}
