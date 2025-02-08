//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/trust"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
)

func newOperatorRecoverCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	var recoverCmd = &cobra.Command{
		Use:   "recover",
		Short: "Recover SPIKE Nexus (do this while SPIKE Nexus is healthy)",
		Run: func(cmd *cobra.Command, args []string) {
			if !auth.IsPilotRecover(spiffeId) {
				fmt.Println("")
				fmt.Println("  You need to have a `recover` role to use this command.")
				fmt.Println("  Please run `./hack/spire-server-entry-recover-register.sh`")
				fmt.Println("  with necessary privileges to assign this role.")
				fmt.Println("")
				log.FatalLn("Aborting.")
			}

			trust.AuthenticateRecover(spiffeId)
			fmt.Println("TODO:// WILL MAKE API REQUEST.")
			fmt.Println("Will save recovery shards.")

			// TODO: make an API request to SPIKE Nexus to get 2 shards.
		},
	}

	return recoverCmd
}
