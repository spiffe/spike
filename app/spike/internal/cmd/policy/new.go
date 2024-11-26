package policy

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewPolicyCommand creates a new Cobra command for managing policies.
func NewPolicyCommand(source *workloadapi.X509Source) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage policies",
	}

	// Add subcommands to the policy command
	cmd.AddCommand(newPolicyDeleteCommand(source))
	cmd.AddCommand(newPolicyCreateCommand(source))
	cmd.AddCommand(newPolicyListCommand(source))
	cmd.AddCommand(newPolicyGetCommand(source))

	return cmd
}
