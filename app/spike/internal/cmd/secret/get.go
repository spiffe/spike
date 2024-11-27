//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"
	"github.com/spiffe/spike/internal/entity/v1/reqres"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/auth"
	"github.com/spiffe/spike/app/spike/internal/net/store"
	"github.com/spiffe/spike/internal/entity/data"
)

// newSecretGetCommand creates and returns a new cobra.Command for retrieving
// secrets. It configures a command that fetches and displays secret data from a
// specified path.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// The command accepts a single argument:
//   - path: Location of the secret to retrieve
//
// Flags:
//   - --version, -v (int): Specific version of the secret to retrieve
//     (default 0) where 0 represents the current version
//
// Returns:
//   - *cobra.Command: Configured get command
//
// The command will:
//  1. Verify SPIKE initialization status via admin token
//  2. Retrieve the secret from the specified path and version
//  3. Display all key-value pairs in the secret's data field
//
// Error cases:
//   - SPIKE not initialized: Prompts user to run 'spike init'
//   - Secret not found: Displays appropriate message
//   - Read errors: Displays error message
func newSecretGetCommand(source *workloadapi.X509Source) *cobra.Command {
	var getCmd = &cobra.Command{
		Use:   "get <path>",
		Short: "Get secrets from the specified path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			state, err := auth.CheckInitState(source)
			if err != nil {
				fmt.Println("Failed to check initialization state:")
				fmt.Println(err.Error())
				return
			}

			if state == data.NotInitialized {
				fmt.Println("Please initialize SPIKE first by running 'spike init'.")
				return
			}

			path := args[0]
			version, _ := cmd.Flags().GetInt("version")

			secret, err := store.GetSecret(source, path, version)
			if err != nil {
				fmt.Println("Error reading secret:", err.Error())
				return
			}

			if secret == nil {
				fmt.Println("Secret not found.")
				return
			}

			d := secret.Data
			for k, v := range d {
				fmt.Printf("%s: %s\n", k, v)
			}
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")

	return getCmd
}

// newSecretMetadataGetCommand creates and returns a new cobra.Command for retrieving
// secrets. It configures a command that fetches and displays secret data from a
// specified path.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// The command accepts a single argument:
//   - path: Location of the secret to retrieve
//
// Flags:
//   - --version, -v (int): Specific version of the secret to retrieve
//     (default 0) where 0 represents the current version
//
// Returns:
//   - *cobra.Command: Configured get command
//
// The command will:
//  1. Verify SPIKE initialization status via admin token
//  2. Retrieve the secret metadata from the specified path and version
//  3. Display all metadata fields and secret versions
//
// Error cases:
//   - SPIKE not initialized: Prompts user to run 'spike init'
//   - Secret not found: Displays appropriate message
//   - Read errors: Displays error message
func newSecretMetadataGetCommand(source *workloadapi.X509Source) *cobra.Command {
	getMetadataCmd := &cobra.Command{
		Use:   "metadata",
		Short: "Manage secrets",
	}
	var getCmd = &cobra.Command{
		Use:   "get <path>",
		Short: "Get secrets metadata from the specified path",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			state, err := auth.CheckInitState(source)
			if err != nil {
				fmt.Println("Failed to check initialization state:")
				fmt.Println(err.Error())
				return
			}

			if state == data.NotInitialized {
				fmt.Println("Please initialize SPIKE first by running 'spike init'.")
				return
			}

			path := args[0]
			version, _ := cmd.Flags().GetInt("version")

			secret, err := store.GetSecretMetadata(source, path, version)
			if err != nil {
				fmt.Println("Error reading secret:", err.Error())
				return
			}

			if secret == nil {
				fmt.Println("Secret not found.")
				return
			}

			printSecretResponse(secret)
		},
	}

	getCmd.Flags().IntP("version", "v", 0, "Specific version to retrieve")

	getMetadataCmd.AddCommand(getCmd)
	return getMetadataCmd
}

// printSecretResponse prints secret metadata
func printSecretResponse(response *reqres.SecretMetadataResponse) {
	printSeparator := func() {
		fmt.Println(strings.Repeat("-", 50))
	}

	formatTime := func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05 MST")
	}

	if response.Metadata != (reqres.SecretRawMetadataResponse{}) {
		fmt.Println("\nMetadata:")
		printSeparator()
		fmt.Printf("Current Version    : %d\n", response.Metadata.CurrentVersion)
		fmt.Printf("Oldest Version     : %d\n", response.Metadata.OldestVersion)
		fmt.Printf("Created Time       : %s\n", formatTime(response.Metadata.CreatedTime))
		fmt.Printf("Last Updated       : %s\n", formatTime(response.Metadata.UpdatedTime))
		fmt.Printf("Max Versions       : %d\n", response.Metadata.MaxVersions)
		printSeparator()
	}

	if len(response.Versions) > 0 {
		fmt.Println("\nSecret Versions:")
		printSeparator()

		for version, versionData := range response.Versions {
			fmt.Printf("Version %d:\n", version)
			fmt.Printf("  Created: %s\n", formatTime(versionData.CreatedTime))
			if versionData.DeletedTime != nil {
				fmt.Printf("  Deleted: %s\n", formatTime(*versionData.DeletedTime))
			}
			printSeparator()
		}
	}

	if response.Err != "" {
		fmt.Printf("\nError: %s\n", response.Err)
		printSeparator()
	}
}
