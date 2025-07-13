//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewOperatorCommand creates a new cobra.Command for managing SPIKE admin
// operations. It initializes an "operator" command with subcommands for
// recovery and restore operations.
//
// Parameters:
//   - source: An X509Source used for SPIFFE authentication
//   - spiffeId: The SPIFFE ID associated with the operator
//
// Returns:
//   - *cobra.Command: A configured cobra command for operator management
func NewOperatorCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "Manage admin operations",
	}

	cmd.AddCommand(newOperatorRecoverCommand(source, spiffeId))
	cmd.AddCommand(newOperatorRestoreCommand(source, spiffeId))

	return cmd
}
