//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"testing"
)

func TestTlsPort(t *testing.T) {
	tests := []struct {
		name       string
		beforeTest func()
		want       string
	}{
		{
			name:       "default spike keeper tls port",
			beforeTest: nil,
			want:       ":8443",
		},
		{
			name: "custom spike keeper tls port",
			beforeTest: func() {
				if err := os.Setenv("SPIKE_KEEPER_TLS_PORT", ":6656"); err != nil {
					panic("failed to set env SPIKE_KEEPER_TLS_PORT")
				}
			},
			want: ":6656",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeTest != nil {
				tt.beforeTest()
			}
			if got := TLSPort(); got != tt.want {
				t.Errorf("TLSPort() = %v, want %v", got, tt.want)
			}
		})
		if err := os.Unsetenv("SPIKE_KEEPER_TLS_PORT"); err != nil {
			panic("failed to unset env SPIKE_KEEPER_TLS_PORT")
		}
	}
}
