package health

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func NewOperatorCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Manage health operations",
	}

	cmd.AddCommand(newStatusCommand(source, SPIFFEID))
	return cmd
}
