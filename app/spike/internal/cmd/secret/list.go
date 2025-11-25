//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

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
	const fName = "newSecretListCommand"
	const notFoundMessage = "No secrets found."

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all secret paths",
		Run: func(cmd *cobra.Command, args []string) {
			trust.AuthenticateForPilot(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable")
				cmd.PrintErrln("The workload API may have lost connection.")
				cmd.PrintErrln("Please check your SPIFFE agent and try again.")
				warnErr := *sdkErrors.ErrSPIFFENilX509Source
				warnErr.Msg = "SPIFFE X509 source is unavailable"
				log.WarnErr(fName, warnErr)
				return
			}

			api := spike.NewWithSource(source)

			keys, err := api.ListSecretKeys()
			if stdout.HandleAPIError(cmd, err) {
				return
			}
			if keys == nil {
				cmd.Println(notFoundMessage)
				return
			}

			if len(*keys) == 0 {
				cmd.Println(notFoundMessage)
				return
			}

			for _, key := range *keys {
				cmd.Printf("- %s\n", key)
			}
		},
	}

	return listCmd
}
