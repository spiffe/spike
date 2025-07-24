//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"golang.org/x/term"

	"github.com/spiffe/spike/app/spike/internal/env"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newOperatorRestoreCommand creates a new cobra command for restoration
// operations on SPIKE Nexus.
//
// This function creates a command that allows privileged operators with the
// 'restore' role to restore SPIKE Nexus after a system failure. The command
// accepts recovery shards interactively and initiates the restoration process.
//
// Parameters:
//   - source *workloadapi.X509Source: The X.509 source for SPIFFE
//     authentication.
//   - spiffeId string: The SPIFFE ID of the caller for role-based access
//     control.
//
// Returns:
//   - *cobra.Command: A cobra command that implements the restoration
//     functionality.
//
// The command performs the following operations:
//   - Verifies the caller has the 'restore' role, aborting otherwise.
//   - Authenticates the restoration request.
//   - Prompts the user to enter a recovery shard (input is hidden for
//     security).
//   - Sends the shard to the SPIKE API to contribute to restoration.
//   - Reports the status of the restoration process to the user.
//
// The command will abort with a fatal error if:
//   - The caller lacks the required 'restore' role.
//   - There's an error reading the recovery shard from input.
//   - The API call to restore using the shard fails.
//   - No status is returned from the restoration attempt.
//
// If restoration is incomplete (more shards are needed), the command displays
// the current count of collected shards and instructs the user to run the
// command again to provide additional shards.
func newOperatorRestoreCommand(
	source *workloadapi.X509Source, spiffeId string,
) *cobra.Command {
	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore SPIKE Nexus (do this if SPIKE Nexus cannot auto-recover)",
		Run: func(cmd *cobra.Command, args []string) {
			if !spiffeid.IsPilotRestore(env.TrustRoot(), spiffeId) {
				fmt.Println("")
				fmt.Println(
					"  You need to have a `restore` role to use this command.")
				fmt.Println(
					"  Please run " +
						"`./hack/bare-metal/entry/spire-server-entry-restore-register.sh`")
				fmt.Println("  with necessary privileges to assign this role.")
				fmt.Println("")
				os.Exit(1)
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

			var shardToRestore [32]byte

			// shard is in `spike:$id:$base64` format
			shardParts := strings.SplitN(string(shard), ":", 3)
			if len(shardParts) != 3 {
				fmt.Println("")
				fmt.Println(
					"Invalid shard format. Expected format: `spike:$id:$secret`.",
				)
				os.Exit(1)
			}

			index := shardParts[1]
			hexData := shardParts[2]

			// 32 bytes encoded in hex should be 64 characters
			if len(hexData) != 64 {
				fmt.Println("")
				fmt.Println(
					"Invalid hex shard length:", len(hexData),
					"(expected 64 characters).",
					"Did you miss some characters when pasting?",
				)
				os.Exit(1)
			}

			decodedShard, err := hex.DecodeString(hexData)

			// Security: Use `defer` for cleanup to ensure it happens even in
			// error paths
			defer func() {
				mem.ClearBytes(shard)
				mem.ClearBytes(decodedShard)
				mem.ClearRawBytes(&shardToRestore)
			}()

			// Security: reset shard immediately after use.
			mem.ClearBytes(shard)

			if err != nil {
				fmt.Println("")
				fmt.Println("Failed to decode recovery shard: ", err.Error())
				os.Exit(1)
			}

			if len(decodedShard) != 32 {
				// Security: reset decodedShard immediately after use.
				mem.ClearBytes(decodedShard)

				fmt.Println("")
				fmt.Println("Invalid recovery shard length: ", len(decodedShard))
				os.Exit(1)
			}

			for i := 0; i < 32; i++ {
				shardToRestore[i] = decodedShard[i]
			}

			// Security: reset decodedShard immediately after use.
			mem.ClearBytes(decodedShard)

			ix, err := strconv.Atoi(index)
			if err != nil {
				fmt.Println("")
				fmt.Println("Invalid shard index: ", err.Error())
				os.Exit(1)
			}

			status, err := api.Restore(ix, &shardToRestore)

			// Security: reset shardToRestore immediately after recovery.
			mem.ClearRawBytes(&shardToRestore)

			if err != nil {
				log.FatalLn(err.Error())
			}

			if status == nil {
				fmt.Println("")
				fmt.Println("Didn't get any status while trying to restore SPIKE.")
				os.Exit(1)
			}

			if status.Restored {
				fmt.Println("")
				fmt.Println("  SPIKE is now restored and ready to use.")
				fmt.Println(
					"  Please run " +
						"`./hack/bare-metal/entry/spire-server-entry-su-register.sh`")
				fmt.Println(
					"  with necessary privileges to start using SPIKE as a superuser.")
				fmt.Println("")
			} else {
				fmt.Println("")
				fmt.Println(" Shards collected: ", status.ShardsCollected)
				fmt.Println(" Shards remaining: ", status.ShardsRemaining)
				fmt.Println(
					" Please run `spike operator restore` " +
						"again to provide the remaining shards.")
				fmt.Println("")
			}
		},
	}

	return restoreCmd
}
