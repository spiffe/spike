//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/config/fs"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffeid"
)

// newOperatorRecoverCommand creates a new cobra command for recovery operations
// on SPIKE Nexus.
//
// This function creates a command that allows privileged operators with the
// 'recover' role to retrieve recovery shards from a healthy SPIKE Nexus system.
// The retrieved shards are saved to the configured recovery directory and can
// be used to restore the system in case of a catastrophic failure.
//
// Parameters:
//   - source: X.509 source for SPIFFE authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID of the caller for role-based access control.
//
// Returns:
//   - *cobra.Command: A cobra command that implements the recovery
//     functionality.
//
// The command performs the following operations:
//   - Verifies the caller has the 'recover' role, aborting otherwise.
//   - Authenticates the recovery request.
//   - Retrieves recovery shards from the SPIKE API.
//   - Cleans the recovery directory of any previous recovery files.
//   - Saves the retrieved shards as text files in the recovery directory.
//   - Provides instructions to the operator about securing the recovery shards.
//
// The command will abort with a fatal error if:
//   - The caller lacks the required 'recover' role.
//   - The API call to retrieve shards fails.
//   - Fewer than 2 shards are retrieved.
//   - It fails to read or clean the recovery directory.
func newOperatorRecoverCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var recoverCmd = &cobra.Command{
		Use:   "recover",
		Short: "Recover SPIKE Nexus (do this while SPIKE Nexus is healthy)",
		RunE: func(cmd *cobra.Command, args []string) error {
			spiffeid.IsPilotRecoverOrDie(SPIFFEID)

			if source == nil {
				return errors.New("SPIFFE X509 source is unavailable")
			}

			api := spike.NewWithSource(source)

			ctx := context.Background()

			shards, apiErr := api.Recover(ctx)
			// Security: clean the shards when we no longer need them.
			defer func() {
				for _, shard := range shards {
					mem.ClearRawBytes(shard)
				}
			}()

			if apiErr != nil {
				return fmt.Errorf(
					"failed to retrieve recovery shards: %w", apiErr,
				)
			}

			if shards == nil {
				return errors.New("no shards found")
			}

			for _, shard := range shards {
				emptyShard := true
				for _, v := range shard {
					if v != 0 {
						emptyShard = false
						break
					}
				}
				if emptyShard {
					return errors.New("empty shard found")
				}
			}

			// Creates the folder if it does not exist.
			recoverDir := fs.PilotRecoveryFolder()

			// Clean the path to normalize it
			cleanPath, absErr := filepath.Abs(filepath.Clean(recoverDir))
			if absErr != nil {
				return absErr
			}

			// Verify the path exists and is a directory
			fileInfo, statErr := os.Stat(cleanPath)
			if statErr != nil || !fileInfo.IsDir() {
				return errors.New("invalid recovery directory path")
			}

			// Ensure the cleaned path doesn't contain suspicious components
			if strings.Contains(cleanPath, "..") ||
				strings.Contains(cleanPath, "./") ||
				strings.Contains(cleanPath, "//") {
				return errors.New("invalid recovery directory path")
			}

			// Ensure the recover directory is clean by
			// deleting any existing recovery files.
			if _, dirStatErr := os.Stat(recoverDir); dirStatErr == nil {
				files, readErr := os.ReadDir(recoverDir)
				if readErr != nil {
					return fmt.Errorf(
						"failed to read recover directory: %w", readErr,
					)
				}

				for _, file := range files {
					if file.Name() != "" && filepath.Ext(file.Name()) == ".txt" &&
						strings.HasPrefix(file.Name(), "spike.recovery") {
						filePath := filepath.Join(recoverDir, file.Name())
						if rmErr := os.Remove(filePath); rmErr != nil {
							// Leaving a stale shard behind next to fresh ones
							// is a security problem, not a cosmetic one.
							return fmt.Errorf(
								"failed to remove stale recovery file %s: %w",
								filePath, rmErr,
							)
						}
					}
				}
			}

			// Save each shard to a file.
			for i, shard := range shards {
				filePath := filepath.Join(
					recoverDir, fmt.Sprintf("spike.recovery.%d.txt", i),
				)

				// Security: The shard is hex-encoded into a byte slice, never
				// into a string. A string is immutable, so a string copy of
				// the shard could not be zeroed after use.
				encoded := make([]byte, hex.EncodedLen(len(shard)))
				hex.Encode(encoded, shard[:])
				out := append([]byte(fmt.Sprintf("spike:%d:", i)), encoded...)

				// 0600 to be more restrictive.
				writeErr := os.WriteFile(filePath, out, 0600)

				// Security: Erase the shard copies as soon as they are written.
				mem.ClearBytes(encoded)
				mem.ClearBytes(out)

				if writeErr != nil {
					return fmt.Errorf("failed to save shard %d: %w", i, writeErr)
				}
			}

			cmd.Println("")
			cmd.Println(
				"  SPIKE Recovery shards saved to the recovery directory:")
			cmd.Println("  " + recoverDir)
			cmd.Println("")
			cmd.Println("  Please make sure that:")
			cmd.Println("    1. You encrypt these shards and keep them safe.")
			cmd.Println("    2. Securely erase the shards from the")
			cmd.Println("       recovery directory after you encrypt them")
			cmd.Println("       and save them to a safe location.")
			cmd.Println("")
			cmd.Println(
				"  If you lose these shards, you will not be able to recover")
			cmd.Println(
				"  SPIKE Nexus in the unlikely event of a total system crash.")
			cmd.Println("")

			return nil
		},
	}

	return recoverCmd
}
