//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/config/fs"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/spike/internal/trust"
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
		Run: func(cmd *cobra.Command, args []string) {
			if !spiffeid.IsPilotRecover(SPIFFEID) {
				cmd.PrintErrln("Error: You need the 'recover' role.")
				cmd.PrintErrln("See https://spike.ist/operations/recovery/")
				return
			}

			trust.AuthenticateForPilotRecover(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable.")
				return
			}

			api := spike.NewWithSource(source)

			shards, apiErr := api.Recover()
			// Security: clean the shards when we no longer need them.
			defer func() {
				for _, shard := range shards {
					mem.ClearRawBytes(shard)
				}
			}()

			if apiErr != nil {
				cmd.PrintErrln("Error: Failed to retrieve recovery shards.")
				return
			}

			if shards == nil {
				cmd.PrintErrln("Error: No shards found.")
				return
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
					cmd.PrintErrln("Error: Empty shard found.")
					return
				}
			}

			// Creates the folder if it does not exist.
			recoverDir := fs.PilotRecoveryFolder()

			// Clean the path to normalize it
			cleanPath, absErr := filepath.Abs(filepath.Clean(recoverDir))
			if absErr != nil {
				cmd.PrintErrf("Error: %v\n", absErr)
				return
			}

			// Verify the path exists and is a directory
			fileInfo, statErr := os.Stat(cleanPath)
			if statErr != nil || !fileInfo.IsDir() {
				cmd.PrintErrln("Error: Invalid recovery directory path.")
				return
			}

			// Ensure the cleaned path doesn't contain suspicious components
			if strings.Contains(cleanPath, "..") ||
				strings.Contains(cleanPath, "./") ||
				strings.Contains(cleanPath, "//") {
				cmd.PrintErrln("Error: Invalid recovery directory path.")
				return
			}

			// Ensure the recover directory is clean by
			// deleting any existing recovery files.
			if _, dirStatErr := os.Stat(recoverDir); dirStatErr == nil {
				files, readErr := os.ReadDir(recoverDir)
				if readErr != nil {
					cmd.PrintErrf("Error: Failed to read recover directory: %v\n",
						readErr)
					return
				}

				for _, file := range files {
					if file.Name() != "" && filepath.Ext(file.Name()) == ".txt" &&
						strings.HasPrefix(file.Name(), "spike.recovery") {
						filePath := filepath.Join(recoverDir, file.Name())
						_ = os.Remove(filePath)
					}
				}
			}

			// Save each shard to a file
			for i, shard := range shards {
				filePath := fmt.Sprintf("%s/spike.recovery.%d.txt", recoverDir, i)

				encodedShard := hex.EncodeToString(shard[:])

				out := fmt.Sprintf("spike:%d:%s", i, encodedShard)

				// 0600 to be more restrictive.
				writeErr := os.WriteFile(filePath, []byte(out), 0600)

				// Security: Hint gc to reclaim memory.
				encodedShard = "" // nolint:ineffassign
				out = ""          // nolint:ineffassign
				runtime.GC()

				if writeErr != nil {
					cmd.PrintErrf("Error: Failed to save shard %d: %v\n",
						i, writeErr)
					return
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
		},
	}

	return recoverCmd
}
