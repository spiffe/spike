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
//  2. Display the results based on the --format flag
//  3. Show "No secrets found" or "[]" if the system is empty
//
// Flags:
//   - --format, -f (string): Output format. Valid options: human/h/plain/p,
//     json/j, yaml/y (default "human")
//
// Example output (human format):
//
//   - secret/path1
//   - secret/path2
//   - secret/path3
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

			outputFormat, formatErr := format.GetFormat(cmd)
			if formatErr != nil {
				cmd.PrintErrf("Error: %v\n", formatErr)
				return
			}

			api := spike.NewWithSource(source)

			keys, err := api.ListSecretKeys()
			if stdout.HandleAPIError(cmd, err) {
				return
			}

			isEmptyList := keys == nil || len(*keys) == 0

			switch outputFormat {
			case format.JSON:
				if isEmptyList {
					cmd.Println("[]")
					return
				}
				output, marshalErr := json.MarshalIndent(keys, "", "  ")
				if marshalErr != nil {
					cmd.PrintErrf("Error formatting output: %v\n", marshalErr)
					return
				}
				cmd.Println(string(output))

			case format.YAML:
				if isEmptyList {
					cmd.Println("[]")
					return
				}
				output, marshalErr := yaml.Marshal(keys)
				if marshalErr != nil {
					cmd.PrintErrf("Error formatting output: %v\n", marshalErr)
					return
				}
				cmd.Print(string(output))

			default: // format.Human
				if isEmptyList {
					cmd.Println("No secrets found.")
					return
				}
				for _, key := range *keys {
					cmd.Printf("- %s\n", key)
				}
			}
		},
	}

	format.AddFormatFlag(listCmd)

	return listCmd
}
