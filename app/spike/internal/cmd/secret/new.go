//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewSecretCommand creates a new Cobra command for managing secrets.
func NewSecretCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	// trust.Authenticate(spiffeId)

	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage secrets",
	}

	// Add subcommands to the policy command
	cmd.AddCommand(newSecretDeleteCommand(source, spiffeId))
	cmd.AddCommand(newSecretUndeleteCommand(source, spiffeId))
	cmd.AddCommand(newSecretListCommand(source, spiffeId))
	cmd.AddCommand(newSecretGetCommand(source, spiffeId))
	cmd.AddCommand(newSecretMetadataGetCommand(source, spiffeId))
	cmd.AddCommand(newSecretPutCommand(source, spiffeId))

	return cmd
}
