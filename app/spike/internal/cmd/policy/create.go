//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/acl"
	"github.com/spiffe/spike/app/spike/internal/net/auth"
	"github.com/spiffe/spike/internal/entity/data"
)

func NewPolicyCreateCommand(source *workloadapi.X509Source) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy create",
		Short: "Create a new policy",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			spiffeIDPattern, _ := cmd.Flags().GetString("spiffeid")
			pathPattern, _ := cmd.Flags().GetString("path")
			permsStr, _ := cmd.Flags().GetString("permissions")

			if name == "" || spiffeIDPattern == "" ||
				pathPattern == "" || permsStr == "" {
				fmt.Println("Error: all flags are required")
				return
			}

			state, err := auth.CheckInitState(source)
			if err != nil {
				fmt.Println("Failed to check init state:", err)
				return
			}

			if state == data.NotInitialized {
				fmt.Println("Please initialize first by running 'spike init'.")
				return
			}

			permissions := strings.Split(permsStr, ",")
			err = acl.CreatePolicy(source, name, spiffeIDPattern, pathPattern, permissions)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Policy created successfully")
		},
	}

	cmd.Flags().String("name", "", "policy name")
	cmd.Flags().String("spiffeid", "", "SPIFFE ID pattern")
	cmd.Flags().String("path", "", "path pattern")
	cmd.Flags().String("permissions", "", "comma-separated permissions")

	return cmd
}
