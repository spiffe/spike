//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func TestNewPolicyGetCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newPolicyGetCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	if cmd.Use != "get [policy-id]" {
		t.Errorf("Expected command use to be 'get [policy-id]', got '%s'",
			cmd.Use)
	}

	if cmd.Short != "Get policy details" {
		t.Errorf("Expected command short description to be "+
			"'Get policy details', got '%s'", cmd.Short)
	}

	// Check if the expected flags are present
	expectedFlags := []string{"name", "format"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be present", flagName)
		}
	}
}

func TestNewPolicyGetCommandWithNilSource(t *testing.T) {
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newPolicyGetCommand(nil, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created even with nil source, got nil")
		return
	}

	// Command should still be created; nil source is handled at runtime
	if cmd.Use != "get [policy-id]" {
		t.Errorf("Expected command use to be 'get [policy-id]', got '%s'",
			cmd.Use)
	}
}

// TestPolicyGetCommandRegistration tests that the get command is registered
// properly as a subcommand of the policy command.
func TestPolicyGetCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	policyCmd := NewCommand(source, SPIFFEIDPattern)

	var getCmd *cobra.Command
	for _, cmd := range policyCmd.Commands() {
		if cmd.Use == "get [policy-id]" {
			getCmd = cmd
			break
		}
	}

	if getCmd == nil {
		t.Error("Expected 'get' command to be registered")
	}
}
