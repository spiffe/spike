//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newSecretUndeleteCommand creates and returns a new cobra.Command for
// restoring deleted secrets. It configures a command that allows users to
// restore one or more previously deleted versions of a secret at a specified
// path.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// The command accepts a single argument:
//   - path: Location of the secret to restore
//
// Flags:
//   - --versions, -v (string): Comma-separated list of version numbers to
//     restore
//   - "0" or empty: Restores current version only (default)
//   - "1,2,3": Restores specific versions
//
// Returns:
//   - *cobra.Command: Configured undelete command
//
// Example Usage:
//
//	spike secret undelete db/pwd           # Restores current version
//	spike secret undelete db/pwd -v 1,2,3  # Restores specific versions
//	spike secret undelete db/pwd -v 0,1,2  # Restores current version plus 1,2
//
// The command performs validation to ensure:
//   - Exactly one path argument is provided
//   - Version numbers are valid non-negative integers
//   - Version strings are properly formatted
func newSecretUndeleteCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	const fName = "newSecretUndeleteCommand"

	var undeleteCmd = &cobra.Command{
		Use:   "undelete <path>",
		Short: "Undelete secrets at the specified path",
		Long: `Undelete secrets at the specified path.
Specify versions using -v or --versions flag with comma-separated values.
Version 0 refers to the current/latest version.
If no version is specified, defaults to undeleting the current version.

Examples:
  spike secret undelete secret/ella           # Undeletes current version
  spike secret undelete secret/ella -v 1,2,3  # Undeletes specific versions
  spike secret undelete secret/ella -v 0,1,2
  # Undeletes current version plus versions 1 and 2`,
		Args: cobra.ExactArgs(1),
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

			path := args[0]

			if !validSecretPath(path) {
				cmd.PrintErrf("Error: invalid secret path: %s\n", path)
				warnErr := *sdkErrors.ErrDataInvalidInput.Clone()
				warnErr.Msg = "invalid secret path"
				log.WarnErr(fName, warnErr)
				return
			}

			versions, _ := cmd.Flags().GetString("versions")
			if versions == "" {
				versions = "0"
			}

			// Parse and validate versions
			versionStrs := strings.Split(versions, ",")
			vv := make([]int, 0, len(versionStrs))
			for _, v := range versionStrs {
				version, err := strconv.Atoi(strings.TrimSpace(v))

				if err != nil {
					cmd.PrintErrf("Error: invalid version number: %s\n", v)
					warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
					warnErr.Msg = "invalid version number"
					log.WarnErr(fName, *warnErr)
					return
				}

				if version < 0 {
					cmd.PrintErrf(
						"Error: version numbers cannot be negative: %s\n", v,
					)
					warnErr := *sdkErrors.ErrDataInvalidInput.Clone()
					warnErr.Msg = "version numbers cannot be negative"
					log.WarnErr(fName, warnErr)
					return
				}

				vv = append(vv, version)
			}

			err := api.UndeleteSecret(path, vv)
			if stdout.HandleAPIError(cmd, err) {
				return
			}

			cmd.Println("OK")
		},
	}

	undeleteCmd.Flags().StringP("versions", "v", "0",
		"Comma-separated list of versions to undelete")

	return undeleteCmd
}
