//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func TestNewSecretGetCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretGetCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	if cmd.Use != "get <path> [key]" {
		t.Errorf("Expected command use to be 'get <path> [key]', got '%s'",
			cmd.Use)
	}

	if cmd.Short != "Get secrets from the specified path" {
		t.Errorf("Expected command short description to be "+
			"'Get secrets from the specified path', got '%s'", cmd.Short)
	}

	// Check if the expected flags are present
	expectedFlags := []string{"version", "format"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be present", flagName)
		}
	}
}

func TestNewSecretGetCommandWithNilSource(t *testing.T) {
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretGetCommand(nil, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created even with nil source, got nil")
		return
	}

	// Command should still be created; nil source is handled at runtime
	if cmd.Use != "get <path> [key]" {
		t.Errorf("Expected command use to be 'get <path> [key]', got '%s'",
			cmd.Use)
	}
}

// TestSecretGetCommandRegistration tests that the get command is registered
// properly as a subcommand of the secret command.
func TestSecretGetCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	secretCmd := NewCommand(source, SPIFFEIDPattern)

	var getCmd *cobra.Command
	for _, cmd := range secretCmd.Commands() {
		if cmd.Use == "get <path> [key]" {
			getCmd = cmd
			break
		}
	}

	if getCmd == nil {
		t.Error("Expected 'get' command to be registered")
	}
}

func TestSecretGetCommandFlagDefaults(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretGetCommand(source, SPIFFEIDPattern)

	// Check version flag default
	versionFlag := cmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Fatal("Expected 'version' flag to be present")
		return
	}
	if versionFlag.DefValue != "0" {
		t.Errorf("Expected version default to be '0', got '%s'",
			versionFlag.DefValue)
	}
	if versionFlag.Shorthand != "v" {
		t.Errorf("Expected version shorthand to be 'v', got '%s'",
			versionFlag.Shorthand)
	}

	// Check format flag default
	formatFlag := cmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("Expected 'format' flag to be present")
		return
	}
	if formatFlag.DefValue != "plain" {
		t.Errorf("Expected format default to be 'plain', got '%s'",
			formatFlag.DefValue)
	}
	if formatFlag.Shorthand != "f" {
		t.Errorf("Expected format shorthand to be 'f', got '%s'",
			formatFlag.Shorthand)
	}
}
