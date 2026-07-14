//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"context"
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"golang.org/x/term"
)

// newOperatorRestoreCommand creates a new cobra command for restoration
// operations on SPIKE Nexus.
//
// This function creates a command that allows privileged operators with the
// 'restore' role to restore SPIKE Nexus after a system failure. The command
// accepts recovery shards interactively and initiates the restoration process.
//
// Parameters:
//   - source: X.509 source for SPIFFE authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID of the caller for role-based access control.
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
//   - Sends the shard to SPIKE Nexus to contribute to restoration.
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
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore SPIKE Nexus (do this if SPIKE Nexus cannot auto-recover)",
		Run: func(cmd *cobra.Command, args []string) {
			spiffeid.IsPilotRestoreOrDie(SPIFFEID)

			cmd.Println("(your input will be hidden as you paste/type it)")
			cmd.Print("Enter recovery shard: ")
			shard, readErr := term.ReadPassword(int(os.Stdin.Fd()))
			if readErr != nil {
				cmd.Println("") // newline after hidden input
				cmd.PrintErrf("Error: %v\n", readErr)
				return
			}

			api := spike.NewWithSource(source)

			var shardToRestore [crypto.AES256KeySize]byte

			// shard is in `spike:$id:$hex` format
			shardParts := strings.SplitN(string(shard), ":", 3)
			if len(shardParts) != 3 {
				cmd.PrintErrln("Error: Invalid shard format.")
				return
			}

			index := shardParts[1]
			hexData := shardParts[2]

			// 32 bytes encoded in hex should be 64 characters
			if len(hexData) != 64 {
				cmd.PrintErrf("Error: Invalid hex shard length: %d (expected 64).\n",
					len(hexData))
				return
			}

			decodedShard, decodeErr := hex.DecodeString(hexData)

			// Security: Use `defer` for cleanup to ensure it happens even in
			// error paths
			defer func() {
				mem.ClearBytes(shard)
				mem.ClearBytes(decodedShard)
				mem.ClearRawBytes(&shardToRestore)
			}()

			// Security: reset shard immediately after use.
			mem.ClearBytes(shard)

			if decodeErr != nil {
				cmd.PrintErrln("Error: Failed to decode recovery shard.")
				return
			}

			if len(decodedShard) != crypto.AES256KeySize {
				// Security: reset decodedShard immediately after use.
				mem.ClearBytes(decodedShard)
				cmd.PrintErrf("Error: Invalid shard length: %d (expected %d).\n",
					len(decodedShard), crypto.AES256KeySize)
				return
			}

			for i := 0; i < crypto.AES256KeySize; i++ {
				shardToRestore[i] = decodedShard[i]
			}

			// Security: reset decodedShard immediately after use.
			mem.ClearBytes(decodedShard)

			ix, atoiErr := strconv.Atoi(index)
			if atoiErr != nil {
				cmd.PrintErrf("Error: Invalid shard index: %s\n", index)
				return
			}

			ctx := context.Background()

			status, restoreErr := api.Restore(ctx, ix, &shardToRestore)
			// Security: reset shardToRestore immediately after recovery.
			mem.ClearRawBytes(&shardToRestore)
			if restoreErr != nil {
				cmd.PrintErrln("Error: Failed to communicate with SPIKE Nexus.")
				return
			}

			if status == nil {
				cmd.PrintErrln("Error: No status returned from SPIKE Nexus.")
				return
			}

			if status.Restored {
				cmd.Println("")
				cmd.Println("  SPIKE is now restored and ready to use.")
				cmd.Println(
					"  See https://spike.ist/operations/recovery/ for next steps.")
				cmd.Println("")
			} else {
				cmd.Println("")
				cmd.Println(" Shards collected: ", status.ShardsCollected)
				cmd.Println(" Shards remaining: ", status.ShardsRemaining)
				cmd.Println(
					" Please run `spike operator restore` " +
						"again to provide the remaining shards.")
				cmd.Println("")
			}
		},
	}

	return restoreCmd
}
