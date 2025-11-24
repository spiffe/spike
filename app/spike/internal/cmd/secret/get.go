//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"encoding/json"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"gopkg.in/yaml.v3"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
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
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get <path> [key]",
		Short: "Get secrets from the specified path",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			trust.AuthenticateForPilot(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable")
				cmd.PrintErrln("The workload API may have lost connection.")
				cmd.PrintErrln("Please check your SPIFFE agent and try again.")
				return
			}

			api := spike.NewWithSource(source)

			path := args[0]
			version, _ := cmd.Flags().GetInt("version")
			format, _ := cmd.Flags().GetString("format")

			if !slices.Contains([]string{"plain",
				"yaml", "json", "y", "p", "j"}, format) {
				cmd.PrintErrf("Error: invalid format specified: %s\n", format)
				return
			}

			if !validSecretPath(path) {
				cmd.PrintErrf("Error: invalid secret path: %s\n", path)
				return
			}

			secret, err := api.GetSecretVersion(path, version)
			if err != nil {
				if err.Is(sdkErrors.ErrStateNotReady) {
					stdout.PrintNotReady()
					return
				}

				cmd.PrintErrf("Error: failure reading secret: %v\n", err.Error())
				return
			}

			if secret == nil {
				cmd.PrintErrln("Error: secret not found")
				return
			}

			if secret.Data == nil {
				cmd.PrintErrln("Error: secret has no data")
				return
			}

			// TODO: this part may benefit from some refactoring; similar code can be extracted into helper functions.

			d := secret.Data

			if format == "plain" || format == "p" {
				found := false
				for k, v := range d {
					if len(args) < 2 || args[1] == "" {
						cmd.Printf("%s: %s\n", k, v)
						found = true
					} else if args[1] == k {
						cmd.Printf("%s\n", v)
						found = true
						break
					}
				}
				if !found {
					cmd.PrintErrln("Error: key not found")
					return
				}

				return
			}

			if len(args) < 2 || args[1] == "" {
				if format == "yaml" || format == "y" {
					b, marshalErr := yaml.Marshal(d)
					if marshalErr != nil {
						cmd.PrintErrf("Error: failed to marshal data: %v\n",
							marshalErr)
						return
					}

					cmd.Printf("%s\n", string(b))
					return
				}

				b, marshalErr := json.MarshalIndent(d, "", "    ")
				if marshalErr != nil {
					cmd.PrintErrf("Error: failed to marshal data: %v\n",
						marshalErr)
					return
				}

				cmd.Printf("%s\n", string(b))
				return
			}

			for k, v := range d {
				if args[1] == k {
					if format == "yaml" || format == "y" {
						b, marshalErr := yaml.Marshal(v)
						if marshalErr != nil {
							cmd.PrintErrf("Error: failed to marshal data: %v\n",
								marshalErr)
							return
						}

						cmd.Printf("%s\n", string(b))
						return
					}

					b, marshalErr := json.Marshal(v)
					if marshalErr != nil {
						cmd.PrintErrf("Error: failed to marshal data: %v\n",
							marshalErr)
						return
					}

					cmd.Printf("%s\n", string(b))
					return
				}
			}

			cmd.PrintErrln("Error: key not found")
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")
	getCmd.Flags().StringP("format", "f", "plain",
		"Format to use. Valid options: plain, p, yaml, y, json, j")

	return getCmd
}
