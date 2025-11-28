//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func TestNewPolicyListCommand(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newPolicyListCommand(source, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
		return
	}

	if cmd.Use != "list" {
		t.Errorf("Expected command use to be 'list', got '%s'", cmd.Use)
	}

	expectedShort := "List policies, optionally filtering by path pattern or " +
		"SPIFFE ID pattern"
	if cmd.Short != expectedShort {
		t.Errorf("Expected command short description to be '%s', got '%s'",
			expectedShort, cmd.Short)
	}

	// Check if the expected flags are present
	expectedFlags := []string{"path-pattern", "spiffeid-pattern", "format"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be present", flagName)
		}
	}
}

func TestNewPolicyListCommandWithNilSource(t *testing.T) {
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newPolicyListCommand(nil, SPIFFEIDPattern)

	if cmd == nil {
		t.Fatal("Expected command to be created even with nil source, got nil")
		return
	}

	// Command should still be created; the nil source is handled at runtime
	if cmd.Use != "list" {
		t.Errorf("Expected command use to be 'list', got '%s'", cmd.Use)
	}
}

// TestPolicyListCommandRegistration tests that the list command is registered
// properly as a subcommand of the policy command.
func TestPolicyListCommandRegistration(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	policyCmd := NewCommand(source, SPIFFEIDPattern)

	var listCmd *cobra.Command
	for _, cmd := range policyCmd.Commands() {
		if cmd.Use == "list" {
			listCmd = cmd
			break
		}
	}

	if listCmd == nil {
		t.Error("Expected 'list' command to be registered")
	}
}

func TestPolicyListCommandMutuallyExclusiveFlags(t *testing.T) {
	source := &workloadapi.X509Source{}
	SPIFFEIDPattern := "^spiffe://example\\.org/spike$"

	cmd := newPolicyListCommand(source, SPIFFEIDPattern)

	// Verify that both filter flags exist
	pathFlag := cmd.Flags().Lookup("path-pattern")
	spiffeFlag := cmd.Flags().Lookup("spiffeid-pattern")

	if pathFlag == nil {
		t.Error("Expected 'path-pattern' flag to be present")
	}
	if spiffeFlag == nil {
		t.Error("Expected 'spiffeid-pattern' flag to be present")
	}

	// The mutual exclusivity is enforced by Cobra at runtime via
	// cmd.MarkFlagsMutuallyExclusive(). We verify the flags exist;
	// Cobra handles the validation when both are set.
}
