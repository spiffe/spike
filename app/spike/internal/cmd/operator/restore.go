//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			spiffeid.IsPilotRestoreOrDie(SPIFFEID)

			shard, readErr := readShardInput(cmd)
			if readErr != nil {
				return readErr
			}

			api := spike.NewWithSource(source)

			var shardToRestore [crypto.AES256KeySize]byte

			// shard is in `spike:$id:$hex` format
			shardParts := strings.SplitN(string(shard), ":", 3)
			if len(shardParts) != 3 {
				return errors.New("invalid shard format")
			}

			index := shardParts[1]
			hexData := shardParts[2]

			// 32 bytes encoded in hex should be 64 characters
			if len(hexData) != 64 {
				return fmt.Errorf(
					"invalid hex shard length: %d (expected 64)", len(hexData),
				)
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
				return errors.New("failed to decode recovery shard")
			}

			if len(decodedShard) != crypto.AES256KeySize {
				// Security: reset decodedShard immediately after use.
				mem.ClearBytes(decodedShard)
				return fmt.Errorf("invalid shard length: %d (expected %d)",
					len(decodedShard), crypto.AES256KeySize)
			}

			for i := 0; i < crypto.AES256KeySize; i++ {
				shardToRestore[i] = decodedShard[i]
			}

			// Security: reset decodedShard immediately after use.
			mem.ClearBytes(decodedShard)

			ix, atoiErr := strconv.Atoi(index)
			if atoiErr != nil {
				return fmt.Errorf("invalid shard index: %s", index)
			}

			ctx := context.Background()

			status, restoreErr := api.Restore(ctx, ix, &shardToRestore)
			// Security: reset shardToRestore immediately after recovery.
			mem.ClearRawBytes(&shardToRestore)
			if restoreErr != nil {
				return fmt.Errorf(
					"failed to communicate with SPIKE Nexus: %w", restoreErr,
				)
			}

			if status == nil {
				return errors.New("no status returned from SPIKE Nexus")
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

			return nil
		},
	}

	return restoreCmd
}

// readShardInput reads a recovery shard from standard input. When stdin
// is a terminal, the input is hidden while typed. Otherwise, the shard
// is read until EOF, which lets scripts (such as the bare-metal recovery
// drill) drive `spike operator restore` non-interactively.
//
// In the non-interactive mode, the process supplying the shard holds a
// copy of it too, so scripted restore should be reserved for development
// environments and recovery drills.
//
// Parameters:
//   - cmd: The Cobra command used for prompting.
//
// Returns:
//   - []byte: The shard bytes with surrounding whitespace removed.
//   - error: An error if reading standard input fails.
func readShardInput(cmd *cobra.Command) ([]byte, error) {
	fd := int(os.Stdin.Fd())

	if term.IsTerminal(fd) {
		cmd.Println("(your input will be hidden as you paste/type it)")
		cmd.Print("Enter recovery shard: ")
		shard, readErr := term.ReadPassword(fd)
		cmd.Println("") // newline after hidden input
		return shard, readErr
	}

	// A shard line is well under 4KB; the limit only guards against
	// unbounded input.
	data, readErr := io.ReadAll(io.LimitReader(os.Stdin, 4096))
	if readErr != nil {
		return nil, readErr
	}

	return bytes.TrimSpace(data), nil
}
