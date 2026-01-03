//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"gopkg.in/yaml.v3"

	"github.com/spiffe/spike/app/spike/internal/cmd/format"
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
		Run: func(cmd *cobra.Command, args []string) {
			trust.AuthenticateForPilot(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable.")
				return
			}

			api := spike.NewWithSource(source)

			path := args[0]
			version, _ := cmd.Flags().GetInt("version")

			outputFormat, formatErr := format.GetFormat(cmd)
			if formatErr != nil {
				cmd.PrintErrf("Error: %v\n", formatErr)
				return
			}

			if !validSecretPath(path) {
				cmd.PrintErrf("Error: Invalid secret path: %s\n", path)
				return
			}

			secret, err := api.GetSecretVersion(path, version)
			if stdout.HandleAPIError(cmd, err) {
				return
			}

			if secret == nil {
				cmd.PrintErrln("Error: Secret not found.")
				return
			}

			if secret.Data == nil {
				cmd.PrintErrln("Error: Secret has no data.")
				return
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
					cmd.PrintErrln("Error: Key not found.")
				}
				return
			}

			// For structured formats (JSON/YAML)
			var dataToFormat interface{}
			if key == "" {
				dataToFormat = d
			} else {
				val, exists := d[key]
				if !exists {
					cmd.PrintErrln("Error: Key not found.")
					return
				}
				dataToFormat = val
			}

			switch outputFormat {
			case format.YAML:
				b, marshalErr := yaml.Marshal(dataToFormat)
				if marshalErr != nil {
					cmd.PrintErrf("Error: %v\n", marshalErr)
					return
				}
				cmd.Printf("%s\n", string(b))

			case format.JSON:
				b, marshalErr := json.MarshalIndent(dataToFormat, "", "    ")
				if marshalErr != nil {
					cmd.PrintErrf("Error: %v\n", marshalErr)
					return
				}
				cmd.Printf("%s\n", string(b))
			}
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")
	format.AddFormatFlag(getCmd)

	return getCmd
}
