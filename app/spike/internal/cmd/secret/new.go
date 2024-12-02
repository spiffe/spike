package secret

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

// NewSecretCommand creates a new Cobra command for managing secrets.
func NewSecretCommand(source *workloadapi.X509Source) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage secrets",
	}

	// Add subcommands to the policy command
	cmd.AddCommand(newSecretDeleteCommand(source))
	cmd.AddCommand(newSecretUndeleteCommand(source))
	cmd.AddCommand(newSecretListCommand(source))
	cmd.AddCommand(newSecretGetCommand(source))
	cmd.AddCommand(newSecretMetadataGetCommand(source))
	cmd.AddCommand(newSecretPutCommand(source))

	return cmd
}
