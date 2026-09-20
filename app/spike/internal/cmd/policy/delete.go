//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// newPolicyDeleteCommand creates a new Cobra command for policy deletion.
// It allows users to delete existing policies by providing the policy name
// as a command line argument or with the --name flag.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before attempting to delete a policy.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy deletion
//
// Command usage:
//
//	delete [policy-name] [flags]
//
// Arguments:
//   - policy-name: The name of the policy to delete (optional
//     if --name is provided)
//
// Flags:
//   - --name: Policy name to delete
//
// Example usage:
//
//	spike policy delete web-service-policy
//	spike policy delete --name=web-service-policy
//
// The command will:
//  1. Check if the system is initialized
//  2. Get the policy name from arguments or --name flag
//  3. Prompt the user to confirm deletion
//  4. If confirmed, attempt to delete the policy with the specified name
//  5. Confirm successful deletion or report any errors
//
// Error conditions:
//   - Neither policy name argument nor --name flag provided
//   - Policy not found by name
//   - User cancels the operation
//   - System not initialized (requires running 'spike init' first)
//   - Insufficient permissions
//   - Policy deletion failure
//
// Note: This operation cannot be undone. The policy will be permanently removed
// from the system. The command requires confirmation before proceeding.
func newPolicyDeleteCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [policy-name]",
		Short: "Delete a policy",
		Long: `Delete a policy by name.

        You can provide either:
        - A policy name as an argument: spike policy delete web-service-policy
        - A policy name with the --name flag:
          spike policy delete --name=my-policy`,
		RunE: func(c *cobra.Command, args []string) error {
			spiffeid.IsPilotOperatorOrDie(SPIFFEID)

			api := spike.NewWithSource(source)

			// Policies are identified by name. The SDK method still calls its
			// parameter "id", but the value it expects is the policy name.
			policyName, nameErr := sendGetPolicyNameRequest(c, args, api)
			if nameErr != nil {
				return stdout.APIError(c, nameErr)
			}

			// Confirm deletion
			c.Printf("Are you sure you want to "+
				"delete policy '%s'? (y/N): ", policyName)
			reader := bufio.NewReader(os.Stdin)
			confirm, readErr := reader.ReadString('\n')
			// An answer without a trailing newline arrives with io.EOF and
			// is still an answer; any other read failure is not.
			if readErr != nil && !errors.Is(readErr, io.EOF) {
				return fmt.Errorf("failed to read confirmation: %w", readErr)
			}
			confirm = strings.TrimSpace(confirm)

			if confirm != "y" && confirm != "Y" {
				c.Println("Operation canceled.")
				return nil
			}

			ctx := context.Background()

			if deleteErr := api.DeletePolicy(ctx, policyName); deleteErr != nil {
				return stdout.APIError(c, deleteErr)
			}

			c.Println("Policy deleted successfully.")
			return nil
		},
	}

	addNameFlag(cmd)

	return cmd
}
