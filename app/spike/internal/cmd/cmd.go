//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/cmd/cipher"
	"github.com/spiffe/spike/app/spike/internal/cmd/operator"
	"github.com/spiffe/spike/app/spike/internal/cmd/policy"
	"github.com/spiffe/spike/app/spike/internal/cmd/secret"
)

// Initialize sets up the CLI command structure with a workload API X.509
// source.
//
// It creates and configures the following commands:
//   - get: Retrieves secrets with optional version specification
//   - delete: Removes specified versions of secrets
//   - undelete: Restores specified versions of secrets
//   - initialization: Initializes the secret management system
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
func Initialize(source *workloadapi.X509Source, SPIFFEID string) {
	rootCmd.AddCommand(policy.NewPolicyCommand(source, SPIFFEID))
	rootCmd.AddCommand(secret.NewSecretCommand(source, SPIFFEID))
	rootCmd.AddCommand(cipher.NewCipherCommand(source, SPIFFEID))
	rootCmd.AddCommand(operator.NewOperatorCommand(source, SPIFFEID))
}

// Execute runs the root command and handles any errors that occur.
// If an error occurs during execution, it prints the error and exits
// with status code 1.
func Execute() {
	var cmdErr error
	if cmdErr = rootCmd.Execute(); cmdErr == nil {
		return
	}

	if _, err := fmt.Fprintf(os.Stderr, "%v\n", cmdErr); err != nil {
		fmt.Println("failed to write to stderr: ", err.Error())
	}
	os.Exit(1)

}
