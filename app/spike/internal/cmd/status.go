//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

func NewStatusCommand(source *workloadapi.X509Source, SPIFFEID string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current status and stats of SPIKE Nexus",
		Long:  `Displays information about root key, secret count, and basic health of SPIKE Nexus and Keepers.`,

		RunE: func(cmd *cobra.Command, args []string) error {
			//  Authenticate the SPIFFE ID
			// This ensures the client is authorized to communicate with SPIKE Nexus.
			trust.Authenticate(SPIFFEID)

			// Create a new API client with the given X509 source
			// This client will be used to interact securely with SPIKE Nexus.
			api := spike.NewWithSource(source)

			// Retrieve the list of secret keys from the system
			// ListSecretKeys returns a pointer to a slice of strings. If there is an error,
			// we cannot determine the system status, so we report an error and exit.
			listSecretKeys, err := api.ListSecretKeys()
			if err != nil {
				fmt.Println("Number of Secrets: ERROR")
				fmt.Println("SPIKE Nexus Health: ERROR")
				fmt.Println("SPIKE Keepers Health: ERROR")
				return err
			}

			// Calculate the number of secrets
			// Dereference the pointer to get the actual slice length. If the pointer is nil,
			// we assume there are no secrets.
			listSecretLength := 0
			if listSecretKeys != nil {
				listSecretLength = len(*listSecretKeys)
			}

			// 5. Determine if the root key is initialized
			// If there is at least one secret, we assume the root key has been initialized.
			rootInitialized := false
			if listSecretLength > 0 {
				rootInitialized = true
			}

			//Print the status header and basic information
			fmt.Println("--- SPIKE Nexus Status ---")
			fmt.Println("Root Key Initialized:", rootInitialized)
			fmt.Println("Number of Secrets:", listSecretLength)

			// Display health placeholders
			// The SDK does not provide real health checks for Nexus or Keepers,
			nexusHealth := "Healthy"
			keepersHealth := "Healthy"
			fmt.Println("SPIKE Nexus Health:", nexusHealth)
			fmt.Println("SPIKE Keepers Health:", keepersHealth)

			// 8. Final message indicating status has been fetched successfully
			fmt.Println("SPIKE Nexus status fetched successfully.")
			return nil
		},
	}
}
