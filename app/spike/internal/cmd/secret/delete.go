//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/errors"
	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

const validPath = `^[a-zA-Z0-9._\-/()?+*|[\]{}\\]+$`

// Helper function to validate secret paths
func validSecretPath(path string) bool {
	r := regexp.MustCompile(validPath)
	return r.MatchString(path)
}

// newSecretDeleteCommand creates and returns a new cobra.Command for deleting
// secrets. It configures a command that allows users to delete one or more
// versions of a secret at a specified path.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// The command accepts a single argument:
//   - path: Location of the secret to delete
//
// Flags:
//   - --versions, -v (string): Comma-separated list of version numbers to
//     delete
//   - "0" or empty: Deletes current version only (default)
//   - "1,2,3": Deletes specific versions
//
// Returns:
//   - *cobra.Command: Configured delete command
//
// Example Usage:
//
//	spike secret delete secret/pass           # Deletes current version
//	spike secret delete secret/pass -v 1,2,3  # Deletes specific versions
//	spike secret delete secret/pass -v 0,1,2  # Deletes current version plus 1,2
//
// The command performs trust to ensure:
//   - Exactly one path argument is provided
//   - Version numbers are valid non-negative integers
//   - Version strings are properly formatted
func newSecretDeleteCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   "delete <path>",
		Short: "Delete secrets at the specified path",
		Long: `Delete secrets at the specified path. 
Specify versions using -v or --versions flag with comma-separated values.
Version 0 refers to the current/latest version.
If no version is specified, defaults to deleting the current version.

Examples:
  spike secret delete secret/apocalyptica          # Deletes current version
  spike secret delete secret/apocalyptica -v 1,2,3 # Deletes specific versions
  spike secret delete secret/apocalyptica -v 0,1,2 
  # Deletes current version plus versions 1 and 2`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			trust.AuthenticateForPilot(SPIFFEID)

			api := spike.NewWithSource(source)

			path := args[0]
			versions, _ := cmd.Flags().GetString("versions")

			if !validSecretPath(path) {
				fmt.Printf("Error: invalid secret path: %s\n", path)
				return
			}

			if versions == "" {
				versions = "0"
			}

			// Parse and validate versions
			versionList := strings.Split(versions, ",")
			for _, v := range versionList {
				version, err := strconv.Atoi(strings.TrimSpace(v))

				if err != nil {
					fmt.Printf("Error: invalid version number: %s\n", v)
					return
				}

				if version < 0 {
					fmt.Printf(
						"Error: version numbers cannot be negative: %s\n", v,
					)
					return
				}
			}

			var vv []int
			for _, v := range versionList {
				iv, err := strconv.Atoi(v)
				if err == nil {
					vv = append(vv, iv)
				}
			}
			if vv == nil {
				vv = []int{}
			}

			err := api.DeleteSecretVersions(path, vv)
			if err != nil {
				if errors.NotReadyError(err) {
					stdout.PrintNotReady()
					return
				}

				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("OK")
		},
	}

	deleteCmd.Flags().StringP("versions", "v", "0",
		"Comma-separated list of versions to delete")

	return deleteCmd
}
