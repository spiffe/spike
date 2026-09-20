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

// newSecretGetCommand creates and returns a new cobra.Command for retrieving
// secrets. It configures a command that fetches and displays secret data from a
// specified path.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Arguments:
//   - path: Location of the secret to retrieve (required)
//   - key: Optional specific key to retrieve from the secret
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
//  2. Retrieve the secret from the specified path and version
//  3. Display all key-value pairs in the secret's data field
//
// Error cases:
//   - SPIKE not initialized: Prompts user to run 'spike init'
//   - Secret not found: Displays an appropriate message
//   - Read errors: Displays an error message
func newSecretGetCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get <path> [key]",
		Short: "Get secrets from the specified path",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			spiffeid.IsPilotOperatorOrDie(SPIFFEID)

			api := spike.NewWithSource(source)

			path := args[0]
			version, flagErr := flags.Int(cmd, "version")
			if flagErr != nil {
				return flagErr
			}

			outputFormat, formatErr := format.GetFormat(cmd)
			if formatErr != nil {
				return formatErr
			}

			if !validSecretPath(path) {
				return fmt.Errorf("invalid secret path: %s", path)
			}

			ctx := context.Background()

			secret, apiErr := api.GetSecretVersion(ctx, path, version)
			if apiErr != nil {
				return stdout.APIError(cmd, apiErr)
			}

			if secret == nil {
				return errors.New("secret not found")
			}

			if secret.Data == nil {
				return errors.New("secret has no data")
			}

			d := secret.Data
			key := ""
			if len(args) >= 2 {
				key = args[1]
			}

			// For human format, use plain key:value output
			if outputFormat == format.Human {
				found := false
				for k, v := range d {
					if key == "" {
						cmd.Printf("%s: %s\n", k, v)
						found = true
					} else if key == k {
						cmd.Printf("%s\n", v)
						found = true
						break
					}
				}
				if !found {
					return errors.New("key not found")
				}
				return nil
			}

			// For structured formats (JSON/YAML)
			var dataToFormat interface{}
			if key == "" {
				dataToFormat = d
			} else {
				val, exists := d[key]
				if !exists {
					return errors.New("key not found")
				}
				dataToFormat = val
			}

			switch outputFormat {
			case format.YAML:
				b, marshalErr := yaml.Marshal(dataToFormat)
				if marshalErr != nil {
					return fmt.Errorf("failed to format output: %w", marshalErr)
				}
				cmd.Printf("%s\n", string(b))

			case format.JSON:
				b, marshalErr := json.MarshalIndent(dataToFormat, "", "    ")
				if marshalErr != nil {
					return fmt.Errorf("failed to format output: %w", marshalErr)
				}
				cmd.Printf("%s\n", string(b))
			}

			return nil
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")
	format.AddFormatFlag(getCmd)

	return getCmd
}
