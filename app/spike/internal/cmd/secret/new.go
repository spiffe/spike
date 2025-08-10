//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewSecretCommand creates a new Cobra command for managing secrets.
func NewSecretCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	// trust.Authenticate(SPIFFEID)

	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage secrets",
	}

	// Add subcommands to the policy command
	cmd.AddCommand(newSecretDeleteCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretUndeleteCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretListCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretGetCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretMetadataGetCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretPutCommand(source, SPIFFEID))

	return cmd
}
