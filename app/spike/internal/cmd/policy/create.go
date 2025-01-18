//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"
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
//
// Returns:
//   - *cobra.Command: Configured Cobra command for policy creation
//
// Command flags:
//   - --name: Name of the policy (required)
//   - --spiffeid: SPIFFE ID pattern for workload matching (required)
//   - --path: Path pattern for access control (required)
//   - --permissions: Comma-separated list of permissions (required)
//
// Example usage:
//
//	spike policy create \
//	    --name "web-service-policy" \
//	    --spiffeid "spiffe://example.org/web-service/*" \
//	    --path "/api/v1/*" \
//	    --permissions "read,write"
//
// The command will:
//  1. Validate that all required flags are provided
//  2. Check if the system is initialized
//  3. Convert the comma-separated permissions into policy permissions
//  4. Create the policy using the provided parameters
//
// Error conditions:
//   - Missing required flags
//   - System not initialized (requires running 'spike init' first)
//   - Invalid SPIFFE ID pattern
//   - Policy creation failure
func newPolicyCreateCommand(source *workloadapi.X509Source) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new policy",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			spiffeIddPattern, _ := cmd.Flags().GetString("spiffeid")
			pathPattern, _ := cmd.Flags().GetString("path")
			permsStr, _ := cmd.Flags().GetString("permissions")

			if name == "" || spiffeIddPattern == "" ||
				pathPattern == "" || permsStr == "" {
				fmt.Println("Error: all flags are required")
				return
			}

			api := spike.NewWithSource(source)

			permissions := strings.Split(permsStr, ",")
			perms := make([]data.PolicyPermission, 0, len(permissions))
			for _, perm := range permissions {
				perms = append(perms, data.PolicyPermission(perm))
			}

			err := api.CreatePolicy(name, spiffeIddPattern, pathPattern, perms)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("Policy created successfully")
		},
	}

	cmd.Flags().String("name", "", "policy name")
	cmd.Flags().String("spiffeid", "", "SPIFFE ID pattern")
	cmd.Flags().String("path", "", "path pattern")
	cmd.Flags().String("permissions", "", "comma-separated permissions")

	return cmd
}
