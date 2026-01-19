//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newSecretListCommand creates and returns a new cobra.Command for listing all
// secret paths. It configures a command that retrieves and displays all
// available secret paths from the system.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured list command
//
// The command will:
//  1. Make a network request to retrieve all available secret paths
//  2. Display the results in a formatted list
//  3. Show "No secrets found" if the system is empty
//
// Output format:
//
//	Secrets:
//	- secret/path1
//	- secret/path2
//	- secret/path3
//
// Note: Requires an initialized SPIKE system and valid authentication
func newSecretListCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all secret paths",
		Run: func(cmd *cobra.Command, args []string) {
			trust.AuthenticateForPilot(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable.")
				return
			}

			api := spike.NewWithSource(source)

			ctx := context.Background()

			keys, err := api.ListSecretKeys(ctx)
			if stdout.HandleAPIError(cmd, err) {
				return
			}
			if keys == nil {
				cmd.Println("No secrets found.")
				return
			}

			if len(*keys) == 0 {
				cmd.Println("No secrets found.")
				return
			}

			for _, key := range *keys {
				cmd.Printf("- %s\n", key)
			}
		},
	}

	return listCmd
}
