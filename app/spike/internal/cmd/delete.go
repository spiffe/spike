//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"github.com/spiffe/spike/app/spike/internal/state"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net"
)

// NewDeleteCommand creates and returns a new cobra.Command for deleting secrets.
// It configures a command that allows users to delete one or more versions of
// a secret at a specified path.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// The command accepts a single argument:
//   - path: Location of the secret to delete
//
// Flags:
//   - --versions, -v (string): Comma-separated list of version numbers to delete
//   - "0" or empty: Deletes current version only (default)
//   - "1,2,3": Deletes specific versions
//
// Returns:
//   - *cobra.Command: Configured delete command
//
// Example Usage:
//
//	spike delete secret/apocalyptica              # Deletes current version
//	spike delete secret/apocalyptica -v 1,2,3     # Deletes specific versions
//	spike delete secret/apocalyptica -v 0,1,2     # Deletes current version plus 1,2
//
// The command performs validation to ensure:
//   - Exactly one path argument is provided
//   - Version numbers are valid non-negative integers
//   - Version strings are properly formatted
func NewDeleteCommand(source *workloadapi.X509Source) *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   "delete <path>",
		Short: "Delete secrets at the specified path",
		Long: `Delete secrets at the specified path. 
Specify versions using -v or --versions flag with comma-separated values.
Version 0 refers to the current/latest version.
If no version is specified, defaults to deleting the current version.

Examples:
  spike delete secret/apocalyptica              # Deletes current version
  spike delete secret/apocalyptica -v 1,2,3     # Deletes specific versions
  spike delete secret/apocalyptica -v 0,1,2     # Deletes current version plus versions 1 and 2`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: new flow
			adminToken, err := state.AdminToken()
			if err != nil {
				fmt.Println("SPIKE is not initialized.")
				fmt.Println("Please run `spike init` to initialize SPIKE.")
				return
			}
			if adminToken == "" {
				fmt.Println("SPIKE is not initialized.")
				fmt.Println("Please run `spike init` to initialize SPIKE.")
				return
			}

			path := args[0]
			versions, _ := cmd.Flags().GetString("versions")

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
					fmt.Printf("Error: version numbers cannot be negative: %s\n", v)
					return
				}
			}

			err = net.DeleteSecret(source, path, versionList)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("OK")
		},
	}

	return deleteCmd
}
