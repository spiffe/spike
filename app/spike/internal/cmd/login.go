//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike/app/spike/internal/net"
	"golang.org/x/term"
	"os"
	"syscall"
)

func NewLoginCommand(source *workloadapi.X509Source) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to SPIKE Nexus",
		Long:  `Login to SPIKE Nexus.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print("Enter admin password: ")
			password, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				fmt.Println("\nFailed to read password:")
				fmt.Println(err.Error())
				return
			}
			fmt.Println()

			token, err := net.Login(source, string(password))
			if err != nil {
				fmt.Println("Failed to login:")
				fmt.Println(err.Error())
				return
			}

			// TODO: where the token is stored should be configurable.
			// Save token to file:
			err = os.WriteFile(".spike-admin-token", []byte(token), 0600)
			if err != nil {
				fmt.Println("Failed to save token to file:")
				fmt.Println(err.Error())
				return
			}

			fmt.Println("Login successful.")
		},
	}

	cmd.Flags().String("password", "", "Password to login to SPIKE Nexus")

	return cmd
}
