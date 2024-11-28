//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
)

// newSecretListCommand creates and returns a new cobra.Command for listing all
// secret paths. It configures a command that retrieves and displays all
// available secret paths from the system.
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
func newSecretListCommand(source *workloadapi.X509Source) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all secret paths",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := spike.CheckInitState(source)
			if err != nil {
				fmt.Println("Failed to check initialization state:")
				fmt.Println(err.Error())
				return
			}

			// Maybe have this as an SDK method instead of exposing the entities.
			if state == data.NotInitialized {
				fmt.Println("Please initialize SPIKE first by running 'spike init'.")
				return
			}

			keys, err := spike.ListSecretKeys(source)
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
