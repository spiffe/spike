//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewSecretCommand creates a new top-level command for managing secrets
// within the SPIKE ecosystem. It acts as a parent for all secret-related
// subcommands that provide CRUD operations on secrets.
//
// Secrets in SPIKE are versioned key-value pairs stored securely in SPIKE
// Nexus. Access to secrets is controlled by policies that match SPIFFE IDs
// and resource paths. All secret operations use SPIFFE-based authentication
// to ensure only authorized workloads can access sensitive data.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable. Subcommands will check for nil
//     and display user-friendly error messages instead of crashing.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured top-level Cobra command for secret management
//
// Available subcommands:
//   - put: Create or update a secret
//   - get: Retrieve a secret value
//   - list: List all secrets (or filtered by path)
//   - delete: Soft-delete a secret (can be recovered)
//   - undelete: Restore a soft-deleted secret
//   - metadata-get: Retrieve secret metadata without the value
//
// Example usage:
//
//	spike secret put secrets/db/password value=mypassword
//	spike secret get secrets/db/password
//	spike secret list secrets/db
//	spike secret delete secrets/db/password
//	spike secret undelete secrets/db/password
//	spike secret metadata-get secrets/db/password
//
// Note: Secret paths are namespace identifiers (e.g., "secrets/db/password"),
// not filesystem paths. They should not start with a forward slash.
//
// Each subcommand has its own set of flags and arguments. See the individual
// command documentation for details.
func NewSecretCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage secrets",
	}

	// Add subcommands to the secret command
	cmd.AddCommand(newSecretDeleteCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretUndeleteCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretListCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretGetCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretMetadataGetCommand(source, SPIFFEID))
	cmd.AddCommand(newSecretPutCommand(source, SPIFFEID))

	return cmd
}
