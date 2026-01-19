//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// newPolicyCreateCommand creates a new Cobra command for policy creation.
// It allows users to create new policies via the command line by specifying
// the policy name, SPIFFE ID pattern, path pattern, and permissions.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before creating a policy.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy creation
//
// Command flags:
//   - --name: Name of the policy (required)
//   - --spiffeid-pattern: SPIFFE ID regex pattern for workload matching
//     (required)
//   - --path-pattern: Path regex pattern for access control (required)
//   - --permissions: Comma-separated list of permissions (required)
//
// Valid permissions:
//   - read: Permission to read secrets
//   - write: Permission to create, update, or delete secrets
//   - list: Permission to list resources
//   - super: Administrative permissions
//
// Example usage:
//
//	spike policy create \
//	    --name "web-service-policy" \
//	    --spiffeid-pattern "^spiffe://example\.org/web-service/.*$" \
//	    --path-pattern "^tenants/acme/creds/.*$" \
//	    --permissions "read,write"
//
// The command will:
//  1. Validate that all required flags are provided
//  2. Check if the system is initialized
//  3. Validate permissions and convert to the expected format
//  4. Check if a policy with the same name already exists
//  5. Create the policy using the provided parameters
//
// Error conditions:
//   - Missing required flags
//   - Invalid permissions specified
//   - Policy with the same name already exists
//   - System not initialized (requires running 'spike init' first)
//   - Invalid SPIFFE ID pattern
//   - Policy creation failure
func newPolicyCreateCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var (
		name            string
		pathPattern     string
		SPIFFEIDPattern string
		permsStr        string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new policy",
		Long: `Create a new policy that grants specific permissions to workloads.

        Example:
        spike policy create --name=db-access
          --path-pattern="^db/.*$" --spiffeid-pattern="^spiffe://example\.org/service/.*$"
          --permissions="read,write"

        Valid permissions: read, write, list, super`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			spiffeid.IsPilotOperatorOrDie(SPIFFEID)

			api := spike.NewWithSource(source)

			// Check if all required flags are provided
			var missingFlags []string
			if name == "" {
				missingFlags = append(missingFlags, "name")
			}
			if pathPattern == "" {
				missingFlags = append(missingFlags, "path-pattern")
			}
			if SPIFFEIDPattern == "" {
				missingFlags = append(missingFlags, "spiffeid-pattern")
			}
			if permsStr == "" {
				missingFlags = append(missingFlags, "permissions")
			}

			if len(missingFlags) > 0 {
				c.PrintErrln("Error: All flags are required.")
				for _, flag := range missingFlags {
					c.PrintErrf("  --%s is missing\n", flag)
				}
				return
			}

			// Validate permissions
			permissions, err := validatePermissions(permsStr)
			if stdout.HandleAPIError(c, err) {
				return
			}

			// Check if a policy with this name already exists
			exists, apiErr := checkPolicyNameExists(api, name)
			if stdout.HandleAPIError(c, apiErr) {
				return
			}

			if exists {
				c.PrintErrf("Error: Policy '%s' already exists.\n", name)
				return
			}

			ctx := context.Background()

			// Create policy
			apiErr = api.CreatePolicy(ctx, name, SPIFFEIDPattern,
				pathPattern, permissions)
			if stdout.HandleAPIError(c, apiErr) {
				return
			}

			c.Println("Policy created successfully.")
		},
	}

	// Define flags
	cmd.Flags().StringVar(&name, "name", "", "Policy name (required)")
	cmd.Flags().StringVar(&pathPattern, "path-pattern", "",
		"Resource path regexp pattern, e.g., '^secrets/.*$' (required)")
	cmd.Flags().StringVar(&SPIFFEIDPattern, "spiffeid-pattern", "",
		"SPIFFE ID regexp pattern, e.g., '^spiffe://example\\.org/service/.*$' (required)")
	cmd.Flags().StringVar(&permsStr, "permissions", "",
		"Comma-separated permissions: read, write, list, super (required)")

	return cmd
}
