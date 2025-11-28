//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func TestNewSecretDeleteCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretDeleteCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	if cmd.Use != "delete <path>" {
		t.Errorf("Expected command use to be 'delete <path>', got '%s'",
			cmd.Use)
	}

	if cmd.Short != "Delete secrets at the specified path" {
		t.Errorf("Expected command short description to be "+
			"'Delete secrets at the specified path', got '%s'", cmd.Short)
	}

	// Check if the `versions` flag is present
	flag := cmd.Flags().Lookup("versions")
	if flag == nil {
		t.Error("Expected flag 'versions' to be present")
	}
	if flag != nil && flag.Shorthand != "v" {
		t.Errorf("Expected flag shorthand to be 'v', got '%s'", flag.Shorthand)
	}
}

func TestNewSecretDeleteCommandWithNilSource(t *testing.T) {
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretDeleteCommand(nil, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created even with nil source, got nil")
		return
	}

	// Command should still be created; the nil source is handled at runtime
	if cmd.Use != "delete <path>" {
		t.Errorf("Expected command use to be 'delete <path>', got '%s'",
			cmd.Use)
	}
}

// TestSecretDeleteCommandRegistration tests that the delete command is
// registered properly as a subcommand of the secret command.
func TestSecretDeleteCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	secretCmd := NewCommand(source, SPIFFEIDPattern)

	var deleteCmd *cobra.Command
	for _, cmd := range secretCmd.Commands() {
		if cmd.Use == "delete <path>" {
			deleteCmd = cmd
			break
		}
	}

	if deleteCmd == nil {
		t.Error("Expected 'delete' command to be registered")
	}
}

func TestSecretDeleteCommandVersionsFlagDefault(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretDeleteCommand(source, SPIFFEIDPattern)

	flag := cmd.Flags().Lookup("versions")
	if flag == nil {
		t.Fatal("Expected 'versions' flag to be present")
		return
	}

	if flag.DefValue != "0" {
		t.Errorf("Expected default value to be '0', got '%s'", flag.DefValue)
	}
}
