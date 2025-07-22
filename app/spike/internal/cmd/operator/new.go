//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/spiffe/spike/internal/lock"
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        if lock.IsLocked() { 
            return fmt.Errorf("SPIKE is locked — please unlock before running this command")
        }
        return nil
    },
	}

	cmd.AddCommand(newOperatorRecoverCommand(source, spiffeId))
	cmd.AddCommand(newOperatorRestoreCommand(source, spiffeId))
	cmd.AddCommand(newOperatorLockCommand(spiffeId))
cmd.AddCommand(newOperatorUnlockCommand(spiffeId))

	return cmd
}
