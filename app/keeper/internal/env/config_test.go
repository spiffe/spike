//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"reflect"
	"testing"
)

func TestKeeperId(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		want       string
	}{
		{
			name: "custom spike keeper id",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_KEEPER_ID", "corp.com"); err != nil {
					panic("failed to set env SPIKE_KEEPER_ID")
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
			if got := KeeperId(); got != tt.want {
				t.Errorf("KeeperId() = %v, want %v", got, tt.want)
			}
		})
		if err := os.Unsetenv("SPIKE_KEEPER_ID"); err != nil {
			panic("failed to unset env SPIKE_KEEPER_ID")
		}
	}
}

func TestPeers(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		want       map[string]string
	}{
		{
			name: "custom spike keeper peers",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_KEEPER_PEERS",
					`{"Peer_1":"https://localhost:8443","Peer_2":"https://localhost:8543","Peer_3":"https://localhost:8643"}`); err != nil {
					panic("failed to set env SPIKE_KEEPER_PEERS")
				}
			},
			want: map[string]string{
				"Peer_1": "https://localhost:8443",
				"Peer_2": "https://localhost:8543",
				"Peer_3": "https://localhost:8643",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := Peers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Peers() = %v, want %v", got, tt.want)
			}
		})
		if err := os.Unsetenv("SPIKE_KEEPER_PEERS"); err != nil {
			panic("failed to unset env SPIKE_KEEPER_PEERS")
		}
	}
}

func TestStateFileName(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		want       string
	}{
		{
			name: "default",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_KEEPER_ID", "corp.com"); err != nil {
					panic("failed to set env SPIKE_KEEPER_ID")
				}
			},
			want: "keeper-corp.com.state",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := StateFileName(); got != tt.want {
				t.Errorf("StateFileName() = %v, want %v", got, tt.want)
			}
		})
		if err := os.Unsetenv("SPIKE_KEEPER_ID"); err != nil {
			panic("failed to unset env SPIKE_KEEPER_ID")
		}
	}
}
