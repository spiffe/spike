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

// Initialize sets up the complete SPIKE CLI command structure by registering
// all top-level command groups with the root command. This function must be
// called before Execute to establish the command hierarchy.
//
// The following command groups are registered:
//   - policy: Manage access control policies
//   - secret: Manage secrets (CRUD operations)
//   - cipher: Encrypt and decrypt data
//   - operator: Operator functions (recover, restore)
//
// Each command group provides its own subcommands and flags. See the
// individual command documentation for details.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for workload authentication
//   - SPIFFEID: The SPIFFE ID used to authenticate with SPIKE Nexus
//
// Example usage:
//
//	source, err := workloadapi.NewX509Source(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	Initialize(source, "spiffe://example.org/pilot")
//	Execute()
func Initialize(source *workloadapi.X509Source, SPIFFEID string) {
	rootCmd.AddCommand(policy.NewPolicyCommand(source, SPIFFEID))
	rootCmd.AddCommand(secret.NewSecretCommand(source, SPIFFEID))
	rootCmd.AddCommand(cipher.NewCipherCommand(source, SPIFFEID))
	rootCmd.AddCommand(operator.NewOperatorCommand(source, SPIFFEID))
}

// Execute runs the root command and processes the entire command execution
// lifecycle. This function should be called after Initialize to start the CLI
// application.
//
// The function handles command execution and error reporting:
//   - Executes the root command (and any subcommands)
//   - Returns successfully (exit code 0) if no errors occur
//   - Prints errors to stderr and exits with code 1 on failure
//
// Error handling:
//   - Command errors are written to stderr
//   - If stderr write fails, error is printed to stdout as fallback
//   - Process exits with status code 1 on any error
//
// This function does not return on error; it terminates the process.
//
// Example usage:
//
//	func main() {
//	    source, SPIFFEID, err := spiffe.Source(ctx)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    Initialize(source, SPIFFEID)
//	    Execute()  // Does not return on error
//	}
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
