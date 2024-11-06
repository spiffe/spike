//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewListCommand creates and returns a new cobra.Command for listing all secret
// paths. It configures a command that retrieves and displays all available
// secret paths from the system.
//
// Parameters:
//   - source: X.509 source for workload API authentication
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
func NewListCommand(source *workloadapi.X509Source) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all secret paths",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Listing all secret paths...")
			fmt.Println("I mean... It WILL list them once someone implements me.")

			//		keys := make([]string, 0, len(store.data))
			//		for k := range store.data {
			//			keys = append(keys, k)
			//		}
			//		if len(keys) == 0 {
			//			fmt.Println("No secrets found")
			//			return
			//		}
			//		fmt.Println("Secrets:")
			//		for _, key := range keys {
			//			fmt.Printf("- %s\n", key)
			//		}

		},
	}

	return listCmd
}
