//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"encoding/hex"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffeid"
	"golang.org/x/term"

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
	const fName = "newOperatorRestoreCommand"

	var restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "Restore SPIKE Nexus (do this if SPIKE Nexus cannot auto-recover)",
		Run: func(cmd *cobra.Command, args []string) {
			if !spiffeid.IsPilotRestore(SPIFFEID) {
				cmd.PrintErrln("")
				cmd.PrintErrln(
					"  You need to have a `restore` role to use this command.")
				cmd.PrintErrln(
					"  Please refer https://spike.ist/operations/recovery/ " +
						"for more info.",
				)
				cmd.PrintErrln(
					"  with necessary privileges to assign this role.")
				cmd.PrintErrln("")
				failErr := *sdkErrors.ErrAccessUnauthorized // copy
				failErr.Msg = "you do not have the required 'restore' role"
				log.FatalErr(fName, failErr)
			}

			trust.AuthenticateForPilotRestore(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable")
				cmd.PrintErrln("The workload API may have lost connection.")
				cmd.PrintErrln("Please check your SPIFFE agent and try again.")
				return
			}

			cmd.Println("(your input will be hidden as you paste/type it)")
			cmd.Print("Enter recovery shard: ")
			shard, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				cmd.Println("") // newline after hidden input
				cmd.PrintErrf("Error reading shard: %v\n", err)
				failErr := sdkErrors.ErrAPIBadRequest.Wrap(err)
				log.FatalErr(fName, *failErr)
			}

			api := spike.NewWithSource(source)

			var shardToRestore [crypto.AES256KeySize]byte

			// shard is in `spike:$id:$hex` format
			shardParts := strings.SplitN(string(shard), ":", 3)
			if len(shardParts) != 3 {
				cmd.PrintErrln("")
				cmd.PrintErrln(
					"Invalid shard format. Expected format: 'spike:$id:$secret'.",
				)
				cmd.PrintErrln("")
				failErr := *sdkErrors.ErrAPIBadRequest // copy
				failErr.Msg = "invalid shard format"
				log.FatalErr(fName, failErr)
			}

			index := shardParts[1]
			hexData := shardParts[2]

			// 32 bytes encoded in hex should be 64 characters
			if len(hexData) != 64 {
				cmd.PrintErrln("")
				cmd.PrintErrln(
					"Invalid hex shard length:", len(hexData),
					"(expected 64 characters).",
					"Did you miss some characters when pasting?",
				)
				cmd.PrintErrln("")
				failErr := *sdkErrors.ErrAPIBadRequest // copy
				failErr.Msg = "invalid hex shard length"
				log.FatalErr(fName, failErr)
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
				cmd.PrintErrln("Failed to decode the recovery shard.")
				cmd.PrintErrln("")
				failErr := sdkErrors.ErrAPIBadRequest.Wrap(err)
				failErr.Msg = "failed to decode recovery shard"
				log.FatalErr(fName, *failErr)
			}

			if len(decodedShard) != crypto.AES256KeySize {
				// Security: reset decodedShard immediately after use.
				mem.ClearBytes(decodedShard)

				cmd.PrintErrln("")
				cmd.PrintErrf(
					"Invalid recovery shard length. Got: %d. Expected: %d.\n",
					len(decodedShard), crypto.AES256KeySize)
				cmd.PrintErrln("")
				failErr := *sdkErrors.ErrCryptoInvalidEncryptionKeyLength // copy
				failErr.Msg = "invalid recovery shard length"
				log.FatalErr(fName, failErr)
			}

			for i := 0; i < crypto.AES256KeySize; i++ {
				shardToRestore[i] = decodedShard[i]
			}

			// Security: reset decodedShard immediately after use.
			mem.ClearBytes(decodedShard)

			ix, err := strconv.Atoi(index)
			if err != nil {
				cmd.PrintErrln("")
				cmd.PrintErrln("Invalid shard index:", index)
				cmd.PrintErrln("")
				failErr := sdkErrors.ErrAPIBadRequest.Wrap(err)
				failErr.Msg = "invalid shard index"
				log.FatalErr(fName, *failErr)
			}

			status, err := api.Restore(ix, &shardToRestore)
			// Security: reset shardToRestore immediately after recovery.
			mem.ClearRawBytes(&shardToRestore)
			if err != nil {
				cmd.PrintErrln("")
				cmd.PrintErrln("There was a problem talking to SPIKE Nexus.")
				cmd.PrintErrln("")
				failErr := sdkErrors.ErrAPIPostFailed.Wrap(err)
				failErr.Msg = "there was a problem talking to SPIKE Nexus"
				log.FatalErr(fName, *failErr)
			}

			if status == nil {
				cmd.PrintErrln("")
				cmd.PrintErrln(
					"Didn't get any status trying to restore SPIKE Nexus.")
				cmd.PrintErrln("Please check SPIKE Nexus logs for more info.")
				cmd.PrintErrln("")
				failErr := *sdkErrors.ErrAPIPostFailed // copy
				failErr.Msg = "bad status response from SPIKE Nexus"
				log.FatalErr(fName, failErr)
			}

			if status.Restored {
				cmd.Println("")
				cmd.Println("  SPIKE is now restored and ready to use.")
				cmd.Println(
					"  Please check out " +
						" https://spike.ist/operations/recovery/ for what to do next.")
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
