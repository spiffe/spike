//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// newPolicyDeleteCommand creates a new Cobra command for policy deletion.
// It allows users to delete existing policies by providing the policy ID
// as a command line argument.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before attempting to delete a policy.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy deletion
//
// Command usage:
//
//	delete <policy-id>
//
// Arguments:
//   - policy-id: The unique identifier of the policy to delete (required)
//
// Example usage:
//
//	spike policy delete policy-123
//
// The command will:
//  1. Check if the system is initialized
//  2. Attempt to delete the policy with the specified ID
//  3. Confirm successful deletion or report any errors
//
// Error conditions:
//   - Missing policy ID argument
//   - System not initialized (requires running 'spike init' first)
//   - Policy not found
//   - Insufficient permissions
//   - Policy deletion failure
//
// Note: This operation cannot be undone. The policy will be permanently removed
// from the system.
func newPolicyDeleteCommand(source *workloadapi.X509Source) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <policy-id>",
		Short: "Delete a policy",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			api := spike.NewWithSource(source)

			state, err := api.CheckInitState()
			if err != nil {
				fmt.Println("Failed to check initialization state:", err)
				return
			}

			if state == data.NotInitialized {
				fmt.Println("Please initialize first by running 'spike init'.")
				return
			}

			policyID := args[0]
			err = api.DeletePolicy(policyID)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Policy deleted successfully")
		},
	}
}
