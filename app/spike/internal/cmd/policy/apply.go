//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"

	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newPolicyApplyCommand creates a new Cobra command for policy application.
// It allows users to apply policies via the command line by specifying
// the policy name, SPIFFE ID pattern, path pattern, and permissions either
// through command line flags or by reading from a YAML file.
//
// The command uses upsert semantics - it will update an existing policy if one
// exists with the same name or create a new policy if it doesn't exist.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before applying a policy.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy application
//
// Command flags:
//   - --name: Name of the policy (required if not using --file)
//   - --spiffeid-pattern: SPIFFE ID regex pattern for workload matching
//     (required if not using --file)
//   - --path-pattern: Path regex pattern for access control (required
//     if not using --file)
//   - --permissions: Comma-separated list of permissions
//     (required if not using --file)
//   - --file: Path to YAML file containing policy configuration
//
// Valid permissions:
//   - read: Permission to read secrets
//   - write: Permission to create, update, or delete secrets
//   - list: Permission to list resources
//   - super: Administrative permissions
//
// Example usage with flags:
//
//	spike policy apply \
//	    --name "web-service-policy" \
//	    --spiffeid-pattern "spiffe://example\.org/web-service/.*" \
//	    --path-pattern "^secrets/web/database$" \
//	    --permissions "read,write"
//
// Example usage with YAML file:
//
//	spike policy apply --file policy.yaml
//
// Example YAML file structure:
//
//	name: web-service-policy
//	spiffeidPattern: ^spiffe://example\.org/web-service/.*$
//	pathPattern: ^secrets/web/database$
//	permissions:
//	  - read
//	  - write
//
// The command will:
//  1. Validate that all required parameters are provided (either via
//     flags or file)
//  2. Validate permissions and convert to the expected format
//  3. Apply the policy using upsert semantics (create if new, update if exists)
//
// Error conditions:
//   - Missing required flags (when not using --file)
//   - Invalid permissions specified
//   - System not initialized (requires running 'spike init' first)
//   - Invalid SPIFFE ID pattern
//   - Policy application failure
//   - File reading errors (when using --file)
//   - Invalid YAML format
func newPolicyApplyCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var (
		name            string
		pathPattern     string
		SPIFFEIDPattern string
		permsStr        string
		filePath        string
	)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a policy configuration",
		Long: `Apply a policy that grants specific permissions to workloads.

        Example using YAML file:
        spike policy apply --file=policy.yaml

        Example YAML file structure:
        name: db-access
        spiffeidPattern: ^spiffe://example\.org/service/.*$
        pathPattern: ^secrets/database/production/.*$
        permissions:
          - read
          - write

        Valid permissions: read, write, list, super`,
		Args: cobra.NoArgs,
		Run: func(c *cobra.Command, args []string) {
			trust.AuthenticateForPilot(SPIFFEID)

			if source == nil {
				c.PrintErrln("Error: SPIFFE X509 source is unavailable.")
				return
			}

			api := spike.NewWithSource(source)

			var policy data.PolicySpec

			// Determine if we're using file-based or flag-based input
			if filePath != "" {
				// Read policy from the YAML file
				p, err := readPolicyFromFile(filePath)
				if err != nil {
					c.PrintErrf("Error reading policy file: %v\n", err)
					return
				}
				policy = p
			} else {
				// Use flag-based input
				p, flagErr := getPolicyFromFlags(name, SPIFFEIDPattern,
					pathPattern, permsStr)
				if stdout.HandleAPIError(c, flagErr) {
					return
				}
				policy = p
			}

			// Convert permissions slice to comma-separated string
			// for validation
			ps := ""
			if len(policy.Permissions) > 0 {
				for i, perm := range policy.Permissions {
					if i > 0 {
						ps += ","
					}
					ps += string(perm)
				}
			}

			// Validate permissions
			permissions, permErr := validatePermissions(ps)
			if stdout.HandleAPIError(c, permErr) {
				return
			}

			ctx := context.Background()

			// Apply policy using upsert semantics
			policyErr := api.CreatePolicy(ctx, policy.Name, policy.SpiffeIDPattern,
				policy.PathPattern, permissions)
			if stdout.HandleAPIError(c, policyErr) {
				return
			}

			c.Printf("Policy '%s' applied successfully\n", policy.Name)
		},
	}

	// Define flags
	cmd.Flags().StringVar(&name, "name", "",
		"Policy name (required if not using --file)")
	cmd.Flags().StringVar(&pathPattern, "path-pattern", "",
		"Resource path regex pattern, e.g., "+
			"'^secrets/database/production/.*$' (required if not using --file)")
	cmd.Flags().StringVar(&SPIFFEIDPattern, "spiffeid-pattern", "",
		"SPIFFE ID regex pattern, e.g., "+
			"'^spiffe://example\\.org/service/.*$' (required if not using --file)")
	cmd.Flags().StringVar(&permsStr, "permissions", "",
		"Comma-separated permissions: read, write, list, "+
			"super (required if not using --file)")
	cmd.Flags().StringVar(&filePath, "file", "",
		"Path to YAML file containing policy configuration")

	return cmd
}
