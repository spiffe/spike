//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package cmd provides the entry point for the CLI application, including the
// initialization and execution mechanisms for SPIKE's secret and policy
// management commands. It integrates with the SPIFFE Workload API to enable
// secure and authenticated communication for secret operations.
//
// The package includes the following functionalities:
//   - Command initialization and configuration for 'policy' and 'secret'
//     related operations
//   - Execution of the root command with error handling
//
// Each command uses SPIFFE's X.509 source for authentication and builds
// a secure environment for managing secrets and policies.
package cmd
