//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func TestNewSecretPutCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretPutCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	if cmd.Use != "put <path> <key=value>..." {
		t.Errorf("Expected command use to be 'put <path> <key=value>...', "+
			"got '%s'", cmd.Use)
	}

	if cmd.Short != "Put secrets at the specified path" {
		t.Errorf("Expected command short description to be "+
			"'Put secrets at the specified path', got '%s'", cmd.Short)
	}
}

func TestNewSecretPutCommandWithNilSource(t *testing.T) {
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretPutCommand(nil, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created even with nil source, got nil")
		return
	}

	// Command should still be created; the nil source is handled at runtime
	if cmd.Use != "put <path> <key=value>..." {
		t.Errorf("Expected command use to be 'put <path> <key=value>...', "+
			"got '%s'", cmd.Use)
	}
}

// TestSecretPutCommandRegistration tests that the put command is registered
// properly as a subcommand of the secret command.
func TestSecretPutCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	secretCmd := NewCommand(source, SPIFFEIDPattern)

	var putCmd *cobra.Command
	for _, cmd := range secretCmd.Commands() {
		if cmd.Use == "put <path> <key=value>..." {
			putCmd = cmd
			break
		}
	}

	if putCmd == nil {
		t.Error("Expected 'put' command to be registered")
	}
}
