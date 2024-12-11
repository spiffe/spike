//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// newSecretMetadataGetCommand creates and returns a new cobra.Command for
// retrieving secrets. It configures a command that fetches and displays secret
// data from a specified path.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// The command accepts a single argument:
//   - path: Location of the secret to retrieve
//
// Flags:
//   - --version, -v (int): Specific version of the secret to retrieve
//     (default 0) where 0 represents the current version
//
// Returns:
//   - *cobra.Command: Configured get command
//
// The command will:
//  1. Verify SPIKE initialization status via admin token
//  2. Retrieve the secret metadata from the specified path and version
//  3. Display all metadata fields and secret versions
//
// Error cases:
//   - SPIKE not initialized: Prompts user to run 'spike init'
//   - Secret not found: Displays appropriate message
//   - Read errors: Displays error message
func newSecretMetadataGetCommand(
	source *workloadapi.X509Source,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Manage secret metadata",
	}

	var getCmd = &cobra.Command{
		Use:   "get <path>",
		Short: "Gets secret metadata from the specified path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			api := spike.NewWithSource(source)

			state, err := api.CheckInitState()
			if err != nil {
				fmt.Println("Failed to check initialization state:")
				fmt.Println(err.Error())
				return
			}

			if state == data.NotInitialized {
				fmt.Println(
					"Please initialize SPIKE first by running 'spike init'.",
				)
				return
			}

			path := args[0]
			version, _ := cmd.Flags().GetInt("version")

			secret, err := api.GetSecretMetadata(path, version)
			if err != nil {
				fmt.Println("Error reading secret:", err.Error())
				return
			}

			if secret == nil {
				fmt.Println("Secret not found.")
				return
			}

			printSecretResponse(secret)
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")

	cmd.AddCommand(getCmd)

	return cmd
}
