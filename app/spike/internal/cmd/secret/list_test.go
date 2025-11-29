//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func TestNewSecretListCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretListCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	if cmd.Use != "list" {
		t.Errorf("Expected command use to be 'list', got '%s'", cmd.Use)
	}

	if cmd.Short != "List all secret paths" {
		t.Errorf("Expected command short description to be "+
			"'List all secret paths', got '%s'", cmd.Short)
	}
}

func TestNewSecretListCommandWithNilSource(t *testing.T) {
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretListCommand(nil, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created even with nil source, got nil")
		return
	}

	// Command should still be created; nil source is handled at runtime
	if cmd.Use != "list" {
		t.Errorf("Expected command use to be 'list', got '%s'", cmd.Use)
	}
}

// TestSecretListCommandRegistration tests that the list command is registered
// properly as a subcommand of the secret command.
func TestSecretListCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	secretCmd := NewCommand(source, SPIFFEIDPattern)

	var listCmd *cobra.Command
	for _, cmd := range secretCmd.Commands() {
		if cmd.Use == "list" {
			listCmd = cmd
			break
		}
	}

	if listCmd == nil {
		t.Error("Expected 'list' command to be registered")
	}
}
