//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"fmt"
	spike "github.com/spiffe/spike-sdk-go/api"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
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
			fmt.Println()

			fmt.Println("Shard is:" + string(shard))

			api := spike.NewWithSource(source)

			status, err := api.Restore(string(shard))

			if err != nil {
				log.FatalLn(err.Error())
			}

			if status == nil {
				fmt.Println("no status")
			}

			fmt.Println("restored?", status.Restored)
			fmt.Println("shards collected", status.ShardsCollected)
			fmt.Println("shards neeeded", status.ShardsRemaining)
			fmt.Println("--------")
		},
	}

	return restoreCmd
}
