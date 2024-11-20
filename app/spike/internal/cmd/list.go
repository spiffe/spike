//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/store"
)

//func adminToken(source *workloadapi.X509Source) string {
//	s, err := auth.CheckInitState(source)
//	if err != nil {
//		fmt.Println("Failed to check init state:")
//		fmt.Println(err.Error())
//		return ""
//	}
//
//	if s == data.NotInitialized {
//		fmt.Println("Please initialize SPIKE with `spike init` first.")
//		return ""
//	}
//
//	adminToken, err := state.AdminToken()
//	if err != nil {
//		fmt.Println("Please login first with `spike login`.")
//		return ""
//	}
//	if adminToken == "" {
//		fmt.Println("Please login first with `spike login`.")
//		return ""
//	}
//	return adminToken
//}

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
			//adminToken := adminToken(source)
			//if adminToken == "" {
			//	return
			//}

			keys, err := store.ListSecretKeys(source)
			if err != nil {
				fmt.Println("Error listing secret keys:", err)
				return
			}

			if len(keys) == 0 {
				fmt.Println("No secrets found")
				return
			}

			for _, key := range keys {
				fmt.Printf("- %s\n", key)
			}
		},
	}

	return listCmd
}
