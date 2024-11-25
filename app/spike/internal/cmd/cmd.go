//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"github.com/spiffe/spike/app/spike/internal/cmd/policy"
	"github.com/spiffe/spike/app/spike/internal/cmd/secret"
	"github.com/spiffe/spike/app/spike/internal/cmd/system"
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// Initialize sets up the CLI command structure with a workload API X.509
// source.
//
// It creates and configures the following commands:
//   - get: Retrieves secrets with optional version specification
//   - delete: Removes specified versions of secrets
//   - undelete: Restores specified versions of secrets
//   - init: Initializes the secret management system
//   - put: Stores new secrets
//   - list: Displays available secrets
//
// Parameters:
//   - source: An X.509 source for workload API authentication
//
// Each command is added to the root command with appropriate flags and options:
//   - get: --version, -v (int) for specific version retrieval
//   - delete: --versions, -v (string) for comma-separated version list
//   - undelete: --versions, -v (string) for comma-separated version list
//
// Example usage:
//
//	source := workloadapi.NewX509Source(...)
//	Initialize(source)
func Initialize(source *workloadapi.X509Source) {
	rootCmd.AddCommand(system.NewSystemInitCommand(source))
	rootCmd.AddCommand(secret.NewSecretGetCommand(source))
	rootCmd.AddCommand(secret.NewSecretDeleteCommand(source))
	rootCmd.AddCommand(secret.NewSecretUndeleteCommand(source))
	rootCmd.AddCommand(secret.NewSecretPutCommand(source))
	rootCmd.AddCommand(secret.NewSecretListCommand(source))
	rootCmd.AddCommand(policy.NewPolicyCreateCommand(source))
	rootCmd.AddCommand(policy.NewPolicyDeleteCommand(source))
	rootCmd.AddCommand(policy.NewPolicyGetCommand(source))
	rootCmd.AddCommand(policy.NewPolicyListCommand(source))
}

// Execute runs the root command and handles any errors that occur.
// If an error occurs during execution, it prints the error and exits
// with status code 1.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
