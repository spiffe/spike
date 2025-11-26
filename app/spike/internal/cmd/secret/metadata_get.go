//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newSecretMetadataGetCommand creates and returns a new cobra.Command for
// retrieving secrets. It configures a command that fetches and displays secret
// data from a specified path.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID to authenticate with
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
//   - Secret not found: Displays an appropriate message
//   - Read errors: Displays an error message
func newSecretMetadataGetCommand(
	source *workloadapi.X509Source, SPIFFEID string,
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
			trust.AuthenticateForPilot(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable.")
				return
			}

			api := spike.NewWithSource(source)

			path := args[0]
			version, _ := cmd.Flags().GetInt("version")

			secret, err := api.GetSecretMetadata(path, version)
			if stdout.HandleAPIError(cmd, err) {
				return
			}

			if secret == nil {
				cmd.Println("Secret not found.")
				return
			}

			printSecretResponse(cmd, secret)
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")

	cmd.AddCommand(getCmd)

	return cmd
}
