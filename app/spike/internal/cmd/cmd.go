//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike-sdk-go/log"

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
//   - source: SPIFFE X.509 SVID source for workload authentication. Can be nil
//     if the Workload API connection is unavailable. Individual subcommands
//     will check for nil and display user-friendly error messages.
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
	// Cobra's Print family writes to OutOrStderr, so without an explicit
	// output writer every piece of normal output (secret values included)
	// lands on stderr, and shell redirection such as
	// `spike secret get db/creds > creds.txt` yields an empty file.
	// Route data output to stdout; the PrintErr family keeps writing to
	// stderr, so errors stay separable. Subcommands inherit this writer.
	rootCmd.SetOut(os.Stdout)

	rootCmd.AddCommand(policy.NewCommand(source, SPIFFEID))
	rootCmd.AddCommand(secret.NewCommand(source, SPIFFEID))
	rootCmd.AddCommand(cipher.NewCommand(source, SPIFFEID))
	rootCmd.AddCommand(operator.NewCommand(source, SPIFFEID))
}

// Execute runs the root command and processes the entire command execution
// lifecycle. This function should be called after Initialize to start the CLI
// application.
//
// The function handles command execution and error reporting:
//   - Executes the root command (and any subcommands)
//   - Returns successfully (exit code 0) if no errors occur
//   - Prints the error to stderr as "Error: <message>" and exits with code
//     1 on failure. Every command handler returns its failure as an error,
//     so a failed command never exits 0.
//
// Error handling:
//   - The command error is written to stderr exactly once; the root command
//     silences Cobra's own error and usage output
//   - If the stderr write fails, the process terminates through the SDK
//     logger with both the command error and the write error
//   - Process exits with status code 1 on any error
//
// This function does not return on error; it terminates the process
// through the SDK fatal helper. main must have routed the SDK logger to
// the Pilot diagnostics log first (logfile.Route), so that the structured
// failure record never reaches stdout.
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
	const fName = "Execute"

	var cmdErr error
	if cmdErr = rootCmd.Execute(); cmdErr == nil {
		return
	}

	if _, err := fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr); err != nil {
		// The Pilot cannot reach its own stderr. There is no other channel
		// to the user, so terminate with both errors on record.
		log.FatalLn(fName, "message", "failed to write the error to stderr",
			"err", err.Error(), "cmdErr", cmdErr.Error())
	}

	// The human-readable message is already on stderr. The structured
	// record goes to the Pilot diagnostics log (see the logfile package),
	// never to stdout, and the process exits with status 1.
	log.FatalLn(fName, "message", "command failed", "err", cmdErr.Error())
}
