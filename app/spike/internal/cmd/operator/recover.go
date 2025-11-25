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
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/spike/internal/trust"
	"github.com/spiffe/spike/internal/config"
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
	const fName = "newOperatorRecoverCommand"

	var recoverCmd = &cobra.Command{
		Use:   "recover",
		Short: "Recover SPIKE Nexus (do this while SPIKE Nexus is healthy)",
		Run: func(cmd *cobra.Command, args []string) {
			if !spiffeid.IsPilotRecover(SPIFFEID) {
				cmd.PrintErrln("")
				cmd.PrintErrln(
					"  You need to have a `recover` role to use this command.")
				cmd.PrintErrln(
					"  Please refer https://spike.ist/operations/recovery/ " +
						"for more info.")
				cmd.PrintErrln(
					"  with necessary privileges to assign this role.")
				cmd.PrintErrln("")
				failErr := *sdkErrors.ErrAccessUnauthorized.Clone()
				failErr.Msg = "you do not have the required 'recover' role"
				log.FatalErr(fName, failErr)
			}

			trust.AuthenticateForPilotRecover(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable")
				cmd.PrintErrln("The workload API may have lost connection.")
				cmd.PrintErrln("Please check your SPIFFE agent and try again.")
				warnErr := *sdkErrors.ErrSPIFFENilX509Source.Clone()
				warnErr.Msg = "SPIFFE X509 source is unavailable"
				log.WarnErr(fName, warnErr)
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
				cmd.PrintErrln("")
				cmd.PrintErrln("  Failed to retrieve recovery shards.")
				cmd.PrintErrln("")
				log.FatalErr(fName, *apiErr)
			}

			if shards == nil {
				cmd.PrintErrln("")
				cmd.PrintErrln("  No shards found.")
				cmd.PrintErrln("  Cannot save recovery shards.")
				cmd.PrintErrln("  Please try again later.")
				cmd.PrintErrln("  If the problem persists, check SPIKE logs.")
				cmd.PrintErrln("")
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
					cmd.PrintErrln("")
					cmd.PrintErrln("  Empty shard found.")
					cmd.PrintErrln("  Cannot save recovery shards.")
					cmd.PrintErrln("  Please try again later.")
					cmd.PrintErrln("  If the problem persists, check SPIKE logs.")
					cmd.PrintErrln("")
					warnErr := *sdkErrors.ErrDataInvalidInput.Clone()
					warnErr.Msg = "empty shard found"
					log.WarnErr(fName, warnErr)
					return
				}
			}

			// Creates the folder if it does not exist.
			recoverDir := config.PilotRecoveryFolder()

			// Clean the path to normalize it
			cleanPath, err := filepath.Abs(filepath.Clean(recoverDir))
			if err != nil {
				cmd.PrintErrln("")
				cmd.PrintErrln("    Error resolving recovery directory path.")
				cmd.PrintErrln("    " + err.Error())
				cmd.PrintErrln("")
				failErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
				failErr.Msg = "error resolving recovery directory path"
				log.FatalErr(fName, *failErr)
			}

			// Verify the path exists and is a directory
			fileInfo, err := os.Stat(cleanPath)
			if err != nil || !fileInfo.IsDir() {
				cmd.PrintErrln("")
				cmd.PrintErrln("    Invalid recovery directory path.")
				cmd.PrintErrln(
					"    Path does not exist or is not a directory.")
				cmd.PrintErrln("")
				failErr := *sdkErrors.ErrDataInvalidInput.Clone()
				failErr.Msg = "invalid recovery directory path"
				log.FatalErr(fName, failErr)
			}

			// Ensure the cleaned path doesn't contain suspicious components
			// This helps catch any attempts at path traversal that survived
			// cleaning
			if strings.Contains(cleanPath, "..") ||
				strings.Contains(cleanPath, "./") ||
				strings.Contains(cleanPath, "//") {
				cmd.PrintErrln("")
				cmd.PrintErrln("    Invalid recovery directory path.")
				cmd.PrintErrln("    Path contains suspicious components.")
				cmd.PrintErrln("")
				failErr := *sdkErrors.ErrDataInvalidInput.Clone()
				failErr.Msg = "path contains suspicious components"
				log.FatalErr(fName, failErr)
			}

			// Ensure the recover directory is clean by
			// deleting any existing recovery files.
			// We are NOT warning the user about this operation because
			// the admin ought to have securely backed up the shards and
			// deleted them from the recover directory anyway.
			if _, err := os.Stat(recoverDir); err == nil {
				files, err := os.ReadDir(recoverDir)
				if err != nil {
					cmd.PrintErrf("Failed to read recover directory %s: %s\n",
						recoverDir, err.Error())
					failErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
					failErr.Msg = "failed to read recover directory"
					log.FatalErr(fName, *failErr)
				}

				for _, file := range files {
					if file.Name() != "" && filepath.Ext(file.Name()) == ".txt" &&
						strings.HasPrefix(file.Name(), "spike.recovery") {
						filePath := filepath.Join(recoverDir, file.Name())
						err := os.Remove(filePath)
						if err != nil {
							cmd.PrintErrf(
								"Failed to delete old recovery file %s: %s\n",
								filePath, err.Error())
						}
					}
				}
			}

			// Save each shard to a file
			for i, shard := range shards {
				filePath := fmt.Sprintf("%s/spike.recovery.%d.txt", recoverDir, i)

				encodedShard := hex.EncodeToString(shard[:])

				out := fmt.Sprintf("spike:%d:%s", i, encodedShard)

				// 0600 to be more restrictive.
				err := os.WriteFile(filePath, []byte(out), 0600)

				// Security: Hint gc to reclaim memory.
				encodedShard = "" // nolint:ineffassign
				out = ""          // nolint:ineffassign
				runtime.GC()

				if err != nil {
					cmd.PrintErrf("Failed to save shard %d: %s\n", i, err.Error())
					failErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
					failErr.Msg = "failed to save recovery shard"
					log.FatalErr(fName, *failErr)
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
