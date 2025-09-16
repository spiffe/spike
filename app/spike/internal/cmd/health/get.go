package health

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

func newStatusCommand(source *workloadapi.X509Source, SPIFFEID string) *cobra.Command {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show system status",
		Run: func(cmd *cobra.Command, args []string) {
			trust.Authenticate(SPIFFEID)

			api := spike.NewWithSource(source)
			status, err := api.GetSystemStatus()
			if err != nil {
				fmt.Println("Error fetching system status:", err)
				return
			}

			fmt.Printf("Health: %s\n", status.Health)
			fmt.Printf("Secrets count: %v\n", *status.SecretsCount)
			fmt.Printf("Backing store: %s\n", status.BackingStore.Type)
			fmt.Printf("Root key status: %s\n", status.RootKey.Status)
			fmt.Printf("FIPS mode: %v\n", status.FIPSMode)
			fmt.Printf("Uptime: %d seconds\n", status.UptimeSeconds)
		},
	}

	return statusCmd
}
