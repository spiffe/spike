package policy

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewPolicyCommand creates a new Cobra command for managing policies.
func NewPolicyCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	// trust.Authenticate(spiffeId)

	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage policies",
	}

	// Add subcommands to the policy command
	cmd.AddCommand(newPolicyDeleteCommand(source, spiffeId))
	cmd.AddCommand(newPolicyCreateCommand(source, spiffeId))
	cmd.AddCommand(newPolicyListCommand(source, spiffeId))
	cmd.AddCommand(newPolicyGetCommand(source, spiffeId))

	return cmd
}
