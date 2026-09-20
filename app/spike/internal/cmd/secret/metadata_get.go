//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"gopkg.in/yaml.v3"

	"github.com/spiffe/spike/app/spike/internal/cmd/flags"
	"github.com/spiffe/spike/app/spike/internal/cmd/format"
	"github.com/spiffe/spike/app/spike/internal/stdout"
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
//   - --format, -f (string): Output format. Valid options: human/h/plain/p,
//     json/j, yaml/y (default "human")
//
// Returns:
//   - *cobra.Command: Configured get command
//
// The command will:
//  1. Verify SPIKE initialization status via admin token
//  2. Retrieve the secret metadata from the specified path and version
//  3. Display metadata based on the --format flag
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
		RunE: func(cmd *cobra.Command, args []string) error {
			spiffeid.IsPilotOperatorOrDie(SPIFFEID)

			outputFormat, formatErr := format.GetFormat(cmd)
			if formatErr != nil {
				return formatErr
			}

			api := spike.NewWithSource(source)

			path := args[0]
			version, flagErr := flags.Int(cmd, "version")
			if flagErr != nil {
				return flagErr
			}

			ctx := context.Background()

			secret, apiErr := api.GetSecretMetadata(ctx, path, version)
			if apiErr != nil {
				return stdout.APIError(cmd, apiErr)
			}

			if secret == nil {
				return errors.New("secret not found")
			}

			switch outputFormat {
			case format.JSON:
				output, marshalErr := json.MarshalIndent(secret, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("failed to format output: %w", marshalErr)
				}
				cmd.Println(string(output))

			case format.YAML:
				output, marshalErr := yaml.Marshal(secret)
				if marshalErr != nil {
					return fmt.Errorf("failed to format output: %w", marshalErr)
				}
				cmd.Print(string(output))

			default: // format.Human
				printSecretResponse(cmd, secret)
			}

			return nil
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")
	format.AddFormatFlag(getCmd)

	cmd.AddCommand(getCmd)

	return cmd
}
