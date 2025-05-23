//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newSecretGetCommand creates and returns a new cobra.Command for retrieving
// secrets. It configures a command that fetches and displays secret data from a
// specified path.
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
//  2. Retrieve the secret from the specified path and version
//  3. Display all key-value pairs in the secret's data field
//
// Error cases:
//   - SPIKE not initialized: Prompts user to run 'spike init'
//   - Secret not found: Displays an appropriate message
//   - Read errors: Displays an error message
func newSecretGetCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get <path> [key]",
		Short: "Get secrets from the specified path",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			trust.Authenticate(spiffeId)

			api := spike.NewWithSource(source)

			path := args[0]
			version, _ := cmd.Flags().GetInt("version")

			if !validSecretPath(path) {
				return fmt.Errorf("invalid secret path: %s", path)
			}

			secret, err := api.GetSecretVersion(path, version)
			if err != nil {
				if err.Error() == "not ready" {
					stdout.PrintNotReady()
					return fmt.Errorf("server not ready")
				}

				return fmt.Errorf("failure reading secret: %v", err.Error())
			}

			if secret == nil {
				return fmt.Errorf("Secret not found")
			}

			d := secret.Data
			for k, v := range d {
				if len(args) < 2 || args[1] == "" {
					fmt.Printf("%s: %s\n", k, v)
				} else if args[1] == k {
					fmt.Printf("\n", v)
				}
			}
			return nil
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")

	return getCmd
}
