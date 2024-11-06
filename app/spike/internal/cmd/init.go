//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/state"
	"github.com/spiffe/spike/internal/crypto"
)

// NewInitCommand creates and returns a new cobra.Command for initializing the
// SPIKE system. It configures a command that handles first-time setup by
// generating and storing authentication credentials.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// Returns:
//   - *cobra.Command: Configured init command
//
// The command will:
//  1. Check if SPIKE is already initialized
//  2. If not initialized:
//     - Generate a new admin token
//     - Save the token using the provided X.509 source
//     - Store the token in ./.spike-token file
//
// Error cases:
//   - Already initialized: Notifies user and exits
//   - Token save failure: Displays error message
//
// Note: This implementation is transitional. Future versions will:
//   - Replace admin token storage with temporary token exchange
//   - Integrate with 'pilot login' flow
//   - Include database configuration
func NewInitCommand(source *workloadapi.X509Source) *cobra.Command {
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize spike configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("######## ASK FOR PASSWORD AND DB DETAILS #####")
			fmt.Println("--this flow will change; no need to save or check admin token--")
			fmt.Println("--`pilot login` will exchange a temp token instead.")

			if state.AdminTokenExists() {
				fmt.Println("SPIKE is already initialized.")
				fmt.Println("Nothing to do.")
				return
			}

			// Generate and set the token
			token := crypto.Token()
			err := state.SaveAdminToken(source, token)
			if err != nil {
				fmt.Println("Failed to save admin token:")
				fmt.Println(err.Error())
				return
			}

			fmt.Println("")
			fmt.Println("    SPIKE system initialization completed.")
			fmt.Println("      Generated admin token and saved it to")
			fmt.Println("        ./.spike-token for future use.")
			fmt.Println("")
		},
	}

	return initCmd
}
