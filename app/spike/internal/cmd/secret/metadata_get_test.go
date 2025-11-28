//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func TestNewSecretMetadataGetCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretMetadataGetCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	// The metadata command is a parent command
	if cmd.Use != "metadata" {
		t.Errorf("Expected command use to be 'metadata', got '%s'", cmd.Use)
	}

	if cmd.Short != "Manage secret metadata" {
		t.Errorf("Expected command short description to be "+
			"'Manage secret metadata', got '%s'", cmd.Short)
	}

	// Check that the 'get' subcommand exists
	var getCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Use == "get <path>" {
			getCmd = sub
			break
		}
	}

	if getCmd == nil {
		t.Fatal("Expected 'get' subcommand to be present")
		return
	}

	if getCmd.Short != "Gets secret metadata from the specified path" {
		t.Errorf("Expected get subcommand short description to be "+
			"'Gets secret metadata from the specified path', got '%s'",
			getCmd.Short)
	}

	// Check if the version flag is present on the get subcommand
	flag := getCmd.Flags().Lookup("version")
	if flag == nil {
		t.Error("Expected flag 'version' to be present on get subcommand")
	}
	if flag != nil && flag.Shorthand != "v" {
		t.Errorf("Expected flag shorthand to be 'v', got '%s'", flag.Shorthand)
	}
}

func TestNewSecretMetadataGetCommandWithNilSource(t *testing.T) {
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretMetadataGetCommand(nil, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created even with nil source, got nil")
		return
	}

	// Command should still be created; nil source is handled at runtime
	if cmd.Use != "metadata" {
		t.Errorf("Expected command use to be 'metadata', got '%s'", cmd.Use)
	}
}

// TestSecretMetadataGetCommandRegistration tests that the metadata command is
// registered properly as a subcommand of the secret command.
func TestSecretMetadataGetCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	secretCmd := NewCommand(source, SPIFFEIDPattern)

	var metadataCmd *cobra.Command
	for _, cmd := range secretCmd.Commands() {
		if cmd.Use == "metadata" {
			metadataCmd = cmd
			break
		}
	}

	if metadataCmd == nil {
		t.Error("Expected 'metadata' command to be registered")
	}
}

func TestSecretMetadataGetCommandVersionFlagDefault(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newSecretMetadataGetCommand(source, SPIFFEIDPattern)

	// Get the 'get' subcommand
	var getCmd *cobra.Command
	for _, sub := range cmd.Commands() {
		if sub.Use == "get <path>" {
			getCmd = sub
			break
		}
	}

	if getCmd == nil {
		t.Fatal("Expected 'get' subcommand to be present")
		return
	}

	flag := getCmd.Flags().Lookup("version")
	if flag == nil {
		t.Fatal("Expected 'version' flag to be present")
		return
	}

	if flag.DefValue != "0" {
		t.Errorf("Expected default value to be '0', got '%s'", flag.DefValue)
	}
}
