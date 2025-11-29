//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewCommand creates a new cobra.Command for managing SPIKE admin
// operations. It initializes an "operator" command with subcommands for
// recovery and restore operations.
//
// Parameters:
//   - source: An X509Source used for SPIFFE authentication. Can be nil if the
//     Workload API connection is unavailable. Subcommands will check for nil
//     and display user-friendly error messages instead of crashing.
//   - SPIFFEID: The SPIFFE ID associated with the operator
//
// Returns:
//   - *cobra.Command: A configured cobra command for operator management
func NewCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "Manage admin operations",
	}

	cmd.AddCommand(newOperatorRecoverCommand(source, SPIFFEID))
	cmd.AddCommand(newOperatorRestoreCommand(source, SPIFFEID))

	return cmd
}
