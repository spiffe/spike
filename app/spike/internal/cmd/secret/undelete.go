//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/spike/internal/cmd/flags"
	"github.com/spiffe/spike/app/spike/internal/stdout"
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
//   - "0" or empty: Restores the current version only (default)
//   - "1,2,3": Restores specific versions
//
// Returns:
//   - *cobra.Command: Configured undelete command
//
// Example Usage:
//
//	spike secret undelete db/pwd           # Restores current version
//	spike secret undelete db/pwd -v 1,2,3  # Restores specific versions
//	spike secret undelete db/pwd -v 0,1,2  # Restores current plus 1 and 2
//
// The command performs validation to ensure:
//   - Exactly one path argument is provided
//   - Version numbers are valid non-negative integers
//   - Version strings are properly formatted
func newSecretUndeleteCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			spiffeid.IsPilotOperatorOrDie(SPIFFEID)

			api := spike.NewWithSource(source)

			path := args[0]

			if !validSecretPath(path) {
				return fmt.Errorf("invalid secret path: %s", path)
			}

			versions, flagErr := flags.String(cmd, "versions")
			if flagErr != nil {
				return flagErr
			}
			if versions == "" {
				versions = "0"
			}

			// Parse and validate versions
			versionStrs := strings.Split(versions, ",")
			vv := make([]int, 0, len(versionStrs))
			for _, v := range versionStrs {
				version, err := strconv.Atoi(strings.TrimSpace(v))

				if err != nil {
					return fmt.Errorf("invalid version number: %s", v)
				}

				if version < 0 {
					return fmt.Errorf("negative version number: %s", v)
				}

				vv = append(vv, version)
			}

			ctx := context.Background()

			if apiErr := api.UndeleteSecret(ctx, path, vv); apiErr != nil {
				return stdout.APIError(cmd, apiErr)
			}

			cmd.Println("OK")
			return nil
		},
	}

	undeleteCmd.Flags().StringP("versions", "v", "0",
		"Comma-separated list of versions to undelete")

	return undeleteCmd
}
