//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func TestNewPolicyDeleteCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newPolicyDeleteCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	if cmd.Use != "delete [policy-name]" {
		t.Errorf("Expected command use to be 'delete [policy-name]', got '%s'",
			cmd.Use)
	}

	if cmd.Short != "Delete a policy" {
		t.Errorf("Expected command short description to be "+
			"'Delete a policy', got '%s'", cmd.Short)
	}

	// Check if the name flag is present
	flag := cmd.Flags().Lookup("name")
	if flag == nil {
		t.Error("Expected flag 'name' to be present")
	}
}

func TestNewPolicyDeleteCommandWithNilSource(t *testing.T) {
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newPolicyDeleteCommand(nil, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created even with nil source, got nil")
		return
	}

	// Command should still be created; the nil source is handled at runtime
	if cmd.Use != "delete [policy-name]" {
		t.Errorf("Expected command use to be 'delete [policy-name]', got '%s'",
			cmd.Use)
	}
}

// TestPolicyDeleteCommandRegistration tests that the delete command is
// registered properly as a subcommand of the policy command.
func TestPolicyDeleteCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	policyCmd := NewCommand(source, SPIFFEIDPattern)

	var deleteCmd *cobra.Command
	for _, cmd := range policyCmd.Commands() {
		if cmd.Use == "delete [policy-name]" {
			deleteCmd = cmd
			break
		}
	}

	if deleteCmd == nil {
		t.Error("Expected 'delete' command to be registered")
	}
}
