//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// normalizePath removes trailing slashes and ensures a consistent path format.
func normalizePath(path string) string {
	if path == "" {
		return path
	}

	// Remove all trailing slashes except for the root path:
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}

	return path
}

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
//   - source: SPIFFE X.509 SVID source for authentication
//   - spiffeId: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy application
//
// Command flags:
//   - --name: Name of the policy (required if not using --file)
//   - --spiffeid: SPIFFE ID pattern for workload matching
//     (required if not using --file)
//   - --path: Path pattern for access control (required if not using --file)
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
//	    --spiffeid "spiffe://example.org/web-service/*" \
//	    --path "secrets/web/database" \
//	    --permissions "read,write"
//
// Example usage with YAML file:
//
//	spike policy apply --file policy.yaml
//
// Example YAML file structure:
//
//	name: web-service-policy
//	spiffeid: ^spiffe://example\.org/web-service/
//	path: ^secrets/web/database$
//	permissions:
//	  - read
//	  - write
//
// The command will:
//  1. Validate that all required parameters are provided (either via
//     flags or file)
//  2. Normalize the path pattern (remove trailing slashes)
//  3. Check if the system is initialized
//  4. Validate permissions and convert to the expected format
//  5. Apply the policy using upsert semantics (create if new, update if exists)
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
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	var (
		name            string
		pathPattern     string
		spiffeIdPattern string
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
        spiffeid: spiffe://example\.org/service/
        path: ^secrets/database/production
        permissions:
          - read
          - write

        Valid permissions: read, write, list, super`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var policy Spec
			var err error

			// Determine if we're using file-based or flag-based input
			if filePath != "" {
				// Read policy from YAML file
				policy, err = readPolicyFromFile(filePath)
				if err != nil {
					fmt.Printf("Error reading policy file: %v\n", err)
					return
				}
			} else {
				// Use flag-based input
				policy, err = getPolicyFromFlags(name, spiffeIdPattern,
					pathPattern, permsStr)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}
			}

			// Normalize the path pattern
			policy.Path = normalizePath(policy.Path)

			// Convert permissions slice to comma-separated string
			// for validation
			ps := ""
			if len(policy.Permissions) > 0 {
				for i, perm := range policy.Permissions {
					if i > 0 {
						ps += ","
					}
					ps += perm
				}
			}

			trust.Authenticate(spiffeId)
			api := spike.NewWithSource(source)

			// Validate permissions
			permissions, err := validatePermissions(ps)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Apply policy using upsert semantics
			err = api.CreatePolicy(policy.Name, policy.SpiffeID,
				policy.Path, permissions)
			if handleAPIError(err) {
				return
			}

			fmt.Printf("Policy '%s' applied successfully\n", policy.Name)
		},
	}

	// Define flags
	cmd.Flags().StringVar(&name, "name", "",
		"Policy name (required if not using --file)")
	cmd.Flags().StringVar(&pathPattern, "path", "",
		"Resource path pattern, e.g., "+
			"'^secrets/database/production' (required if not using --file)")
	cmd.Flags().StringVar(&spiffeIdPattern, "spiffeid", "",
		"SPIFFE ID pattern, e.g., "+
			"'^spiffe://example\\.org/service/' (required if not using --file)")
	cmd.Flags().StringVar(&permsStr, "permissions", "",
		"Comma-separated permissions: read, write, list, "+
			"super (required if not using --file)")
	cmd.Flags().StringVar(&filePath, "file", "",
		"Path to YAML file containing policy configuration")

	return cmd
}
