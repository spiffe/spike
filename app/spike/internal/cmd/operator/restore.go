//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"golang.org/x/term"

	"github.com/spiffe/spike/app/spike/internal/trust"
	"github.com/spiffe/spike/internal/auth"
	"github.com/spiffe/spike/internal/log"
)

func newOperatorRestoreCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore SPIKE Nexus (do this if SPIKE Nexus cannot auto-recover)",
		Run: func(cmd *cobra.Command, args []string) {
			if !auth.IsPilotRestore(spiffeId) {
				fmt.Println("")
				fmt.Println("  You need to have a `restore` role to use this command.")
				fmt.Println("  Please run `./hack/spire-server-entry-restore-register.sh`")
				fmt.Println("  with necessary privileges to assign this role.")
				fmt.Println("")
				log.FatalLn("Aborting.")
			}

			trust.AuthenticateRestore(spiffeId)

			fmt.Println("(your input will be hidden as you paste/type it)")
			fmt.Print("Enter recovery shard: ")
			shard, err := term.ReadPassword(syscall.Stdin)
			if err != nil {
				_, e := fmt.Fprintf(os.Stderr, "Error reading shard: %v\n", err)
				if e != nil {
					return
				}
				os.Exit(1)
			}

			api := spike.NewWithSource(source)

			status, err := api.Restore(string(shard))

			if err != nil {
				log.FatalLn(err.Error())
			}

			if status == nil {
				log.FatalLn("no status")
			}

			if status.Restored {
				fmt.Println("")
				fmt.Println("  SPIKE is now restored and ready to use.")
				fmt.Println("  Please run `./hack/spire-server-entry-su-register.sh`")
				fmt.Println("  with necessary privileges to start using SPIKE as a superuser.")
				fmt.Println("")
			} else {
				fmt.Println("")
				fmt.Println(" Shards collected: ", status.ShardsCollected)
				fmt.Println(" Shards remaining: ", status.ShardsRemaining)
				fmt.Println(" Please run `spike restore` again to provide the remaining shards.")
				fmt.Println("")
			}
		},
	}

	return restoreCmd
}
