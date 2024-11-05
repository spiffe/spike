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
