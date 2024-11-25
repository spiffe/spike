//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/acl"
	"github.com/spiffe/spike/app/spike/internal/net/auth"
	"github.com/spiffe/spike/internal/entity/data"
)

func NewPolicyListCommand(source *workloadapi.X509Source) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all policies",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			state, err := auth.CheckInitState(source)
			if err != nil {
				fmt.Println("Failed to check init state:", err)
				return
			}

			if state == data.NotInitialized {
				fmt.Println("Please initialize first by running 'spike init'.")
				return
			}

			policies, err := acl.ListPolicies(source)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			output, err := json.MarshalIndent(policies, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting output: %v\n", err)
				return
			}

			fmt.Println(string(output))
		},
	}
}
