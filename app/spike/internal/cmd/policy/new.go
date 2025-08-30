package policy

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewPolicyCommand creates a new top-level command for working with policies.
// It acts as a parent for all policy-related subcommands: create, list, get,
// and delete.
//
// The policy commands allow for managing access control policies that define
// which workloads can access which resources based on SPIFFE ID patterns and
// path patterns.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication
//   - spiffeId: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured top-level Cobra command for policy management
//
// Available subcommands:
//   - create: Create a new policy
//   - list: List all existing policies
//   - get: Get details of a specific policy by ID or name
//   - delete: Delete a policy by ID or name
//
// Example usage:
//
//		spike policy list
//		spike policy get abc123
//		spike policy get --name=my-policy
//		spike policy create --name=new-policy --path="^secret/.*$" \
//	 	--spiffeid="^spiffe://example\.org/.*$" --permissions=read,write
//		spike policy delete abc123
//		spike policy delete --name=my-policy
//
// Each subcommand has its own set of flags and arguments. See the individual
// command documentation for details.
func NewPolicyCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage policies",
		Long: `Manage access control policies.

		Policies control which workloads can access which secrets.
		Each policy defines a set of permissions granted to workloads
		matching a SPIFFE ID pattern for resources matching a path pattern.

		Available subcommands:
		create    Create a new policy
		list      List all policies
		get       Get details of a specific policy
		delete    Delete a policy`,
	}

	// Add subcommands
	cmd.AddCommand(newPolicyListCommand(source, SPIFFEID))
	cmd.AddCommand(newPolicyGetCommand(source, SPIFFEID))
	cmd.AddCommand(newPolicyCreateCommand(source, SPIFFEID))
	cmd.AddCommand(newPolicyDeleteCommand(source, SPIFFEID))
	cmd.AddCommand(newPolicyApplyCommand(source, SPIFFEID))

	return cmd
}
