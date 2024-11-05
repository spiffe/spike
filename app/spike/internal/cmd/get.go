//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net"
	"github.com/spiffe/spike/app/spike/internal/state"
)

func NewGetCommand(source *workloadapi.X509Source) *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get <path>",
		Short: "Get secrets from the specified path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			version, _ := cmd.Flags().GetInt("version")

			fmt.Println("######## GET SESSION TOKENS INSTEAD #######")
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

			secret, err := net.GetSecret(source, path, version)
			if err != nil {
				fmt.Println("Error reading secret:", err.Error())
				return
			}

			data := secret.Data
			for k, v := range data {
				fmt.Printf("%s: %s\n", k, v)
			}
		},
	}

	return getCmd
}
