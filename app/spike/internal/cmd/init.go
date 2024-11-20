//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/auth"
	"github.com/spiffe/spike/internal/entity/data"
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
//     - Store the token in SpikeAdminTokenFile()
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
			state, err := auth.CheckInitState(source)

			if err != nil {
				fmt.Println("Failed to check init state:")
				fmt.Println(err.Error())
				return
			}

			if state == data.AlreadyInitialized {
				fmt.Println("SPIKE is already initialized.")
				fmt.Println("Nothing to do.")
				return
			}

			//fmt.Println("SPIKE is not initialized.")
			//fmt.Println("As the first user, you will be the admin.")
			//fmt.Println("Choose a strong password:")
			//fmt.Println("* The password should be at least 16 characters long.")
			//fmt.Println("* Make sure the password is a mix of letters, numbers, and symbols.")
			//fmt.Println("")

			//fmt.Print("Enter admin password: ")
			//password, err := term.ReadPassword(syscall.Stdin)
			//if err != nil {
			//	fmt.Println("\nFailed to read password:")
			//	fmt.Println(err.Error())
			//	return
			//}
			//fmt.Println()
			//
			//if len(password) < 16 {
			//	fmt.Println("Password is too short.")
			//	fmt.Println("Please try again.")
			//	return
			//}

			//fmt.Print("Confirm admin password: ")
			//confirm, err := term.ReadPassword(syscall.Stdin)
			//if err != nil {
			//	fmt.Println("\nFailed to read password:")
			//	fmt.Println(err.Error())
			//	return
			//}
			//fmt.Println()
			//
			//if string(password) != string(confirm) {
			//	fmt.Println("Passwords do not match.")
			//	fmt.Println("Please try again.")
			//	return
			//}
			//
			//passwordStr := string(password)

			// TODO:
			// 1. get recovery token from response and save it.
			// 2. make sure current user cannot use spike
			// 3. make sure leonardo can use spike.
			// 4. create a demo recording about it.
			// 5. code cleanup.
			// 6. make sure root key is encrypted with the token and backed up.
			// 7. implement a /recover endpoint to recover the root key.
			// 8. Let keeper and nexus crash; and use /recover to recover the root key
			//    you'll need to enable sqlite too.
			// 9. address TODO items in the source code.

			err = auth.Init(source)
			//err = auth.Init(source, passwordStr)
			if err != nil {
				fmt.Println("Failed to save admin token:")
				fmt.Println(err.Error())
				return
			}

			fmt.Println("")
			fmt.Println("    SPIKE system initialization completed.")
			fmt.Println("    Use `spike login` to authenticate.")
			fmt.Println("")
		},
	}

	return initCmd
}
