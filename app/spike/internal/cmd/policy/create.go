//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newPolicyCreateCommand creates a new Cobra command for policy creation.
// It allows users to create new policies via the command line by specifying
// the policy name, SPIFFE ID pattern, path pattern, and permissions.
//
// The command requires an X509Source for SPIFFE authentication and validates
// that the system is initialized before creating a policy.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication
//   - spiffeId: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy creation
//
// Command flags:
//   - --name: Name of the policy (required)
//   - --spiffeid: SPIFFE ID regex pattern for workload matching (required)
//   - --path: Path regex pattern for access control (required)
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
//	    --spiffeid "spiffe://example\.org/web-service/.*" \
//	    --path "tenants/acme/creds/.*" \
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
          --path="^db/" --spiffeid="spiffe://example\.org/service/" 
          --permissions="read,write"

        Valid permissions: read, write, list, super`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			// Check if all required flags are provided
			var missingFlags []string
			if name == "" {
				missingFlags = append(missingFlags, "name")
			}
			if pathPattern == "" {
				missingFlags = append(missingFlags, "path")
			}
			if SPIFFEIDPattern == "" {
				missingFlags = append(missingFlags, "spiffeid")
			}
			if permsStr == "" {
				missingFlags = append(missingFlags, "permissions")
			}

			if len(missingFlags) > 0 {
				fmt.Println("Error: all flags are required")
				for _, flag := range missingFlags {
					fmt.Printf("  --%s is missing\n", flag)
				}
				return
			}

			trust.Authenticate(SPIFFEID)
			api := spike.NewWithSource(source)

			// Validate permissions
			permissions, err := validatePermissions(permsStr)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			// Check if a policy with this name already exists
			exists, err := checkPolicyNameExists(api, name)
			if handleAPIError(err) {
				return
			}

			if exists {
				fmt.Printf("Error: A policy with name '%s' already exists\n", name)
				return
			}

			// Create policy
			err = api.CreatePolicy(name, SPIFFEIDPattern, pathPattern, permissions)
			if handleAPIError(err) {
				return
			}

			fmt.Println("Policy created successfully")
		},
	}

	// Define flags
	cmd.Flags().StringVar(&name, "name", "", "Policy name (required)")
	cmd.Flags().StringVar(&pathPattern, "path", "",
		"Resource path pattern, e.g., '^secrets/' (required)")
	cmd.Flags().StringVar(&SPIFFEIDPattern, "spiffeid", "",
		"SPIFFE ID pattern, e.g., '^spiffe://example\\.org/service/' (required)")
	cmd.Flags().StringVar(&permsStr, "permissions", "",
		"Comma-separated permissions: read, write, list, super (required)")

	return cmd
}
