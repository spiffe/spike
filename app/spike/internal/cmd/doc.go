//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package cmd provides the command-line interface for SPIKE Pilot, the CLI
// tool for interacting with SPIKE Nexus. It implements the root command and
// orchestrates all subcommand groups.
//
// The package exposes two primary functions:
//
//   - Initialize: Registers all command groups with the root command
//   - Execute: Runs the CLI and handles the command execution lifecycle
//
// # Command Groups
//
// The following command groups are available:
//
//   - secret: Create, read, update, and delete secrets with versioning support
//   - policy: Manage access control policies for workload authorization
//   - cipher: Encrypt and decrypt data using SPIKE's cryptographic services
//   - operator: Perform administrative operations such as recovery and restore
//
// # Authentication
//
// All commands authenticate using SPIFFE X.509 SVIDs obtained from the
// Workload API. The SPIFFE ID identifies the workload to SPIKE Nexus, which
// enforces access control based on configured policies.
//
// # Usage
//
// The typical initialization pattern:
//
//	func main() {
//	    source, SPIFFEID, err := spiffe.Source(ctx)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    cmd.Initialize(source, SPIFFEID)
//	    cmd.Execute()
//	}
//
// # Error Handling
//
// Execute terminates the process with exit code 1 on any command error.
// Errors are written to stderr; if stderr is unavailable, stdout is used
// as a fallback.
package cmd
