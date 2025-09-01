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
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/spike/internal/env"
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
//   - source *workloadapi.X509Source: The X.509 source for SPIFFE
//     authentication.
//   - spiffeId string: The SPIFFE ID of the caller for role-based access
//     control.
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
			if !spiffeid.IsPilotRecover(env.TrustRoot(), SPIFFEID) {
				fmt.Println("")
				fmt.Println("  You need to have a `recover` role to use this command.")
				fmt.Println(
					"  Please run `./hack/bare-metal/entry/spire-server-entry-recover-register.sh`")
				fmt.Println("  with necessary privileges to assign this role.")
				fmt.Println("")
				log.FatalLn("Aborting.")
			}

			trust.AuthenticateRecover(SPIFFEID)

			api := spike.NewWithSource(source)

			shards, err := api.Recover()
			// Security: clean the shards when we no longer need them.
			defer func() {
				for _, shard := range shards {
					mem.ClearRawBytes(shard)
				}
			}()

			if err != nil {
				log.FatalLn(err.Error())
			}

			if shards == nil {
				fmt.Println("")
				fmt.Println("  No shards found.")
				fmt.Println("  Cannot save recovery shards.")
				fmt.Println("  Please try again later.")
				fmt.Println("  If the problem persists, check SPIKE logs.")
				fmt.Println("")

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
					fmt.Println("")
					fmt.Println("  Empty shard found.")
					fmt.Println("  Cannot save recovery shards.")
					fmt.Println("  Please try again later.")
					fmt.Println("  If the problem persists, check SPIKE logs.")
				}
			}

			// Creates the folder if it does not exist.
			recoverDir := config.PilotRecoveryFolder()

			// Clean the path to normalize it
			cleanPath, err := filepath.Abs(filepath.Clean(recoverDir))
			if err != nil {
				fmt.Println("")
				fmt.Println("    Error resolving recovery directory path.")
				fmt.Println("    " + err.Error())
				fmt.Println("")
				log.FatalLn("Aborting.")
			}

			// Verify the path exists and is a directory
			fileInfo, err := os.Stat(cleanPath)
			if err != nil || !fileInfo.IsDir() {
				fmt.Println("")
				fmt.Println("    Invalid recovery directory path.")
				fmt.Println("    Path does not exist or is not a directory.")
				fmt.Println("")
				log.FatalLn("Aborting.")
			}

			// Ensure the cleaned path doesn't contain suspicious components
			// This helps catch any attempts at path traversal that survived cleaning
			if strings.Contains(cleanPath, "..") ||
				strings.Contains(cleanPath, "./") ||
				strings.Contains(cleanPath, "//") {
				fmt.Println("")
				fmt.Println("    Invalid recovery directory path.")
				fmt.Println("    Path contains suspicious components.")
				fmt.Println("")
				log.FatalLn("Aborting.")
			}

			// Ensure the recover directory is clean by
			// deleting any existing recovery files.
			// We are NOT warning the user about this operation because
			// the admin ought to have securely backed up the shards and
			// deleted them from the recover directory anyway.
			if _, err := os.Stat(recoverDir); err == nil {
				files, err := os.ReadDir(recoverDir)
				if err != nil {
					fmt.Printf("Failed to read recover directory %s: %s\n",
						recoverDir, err.Error())
					log.FatalLn(err.Error())
				}

				for _, file := range files {
					if file.Name() != "" && filepath.Ext(file.Name()) == ".txt" &&
						strings.HasPrefix(file.Name(), "spike.recovery") {
						filePath := filepath.Join(recoverDir, file.Name())
						err := os.Remove(filePath)
						if err != nil {
							fmt.Printf("Failed to delete old recovery file %s: %s\n",
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
					fmt.Printf("Failed to save shard %d: %s\n", i, err.Error())
				}
			}

			fmt.Println("")
			fmt.Println("  SPIKE Recovery shards saved to the recovery directory:")
			fmt.Println("  " + recoverDir)
			fmt.Println("")
			fmt.Println("  Please make sure that:")
			fmt.Println("    1. You encrypt these shards and keep them safe.")
			fmt.Println("    2. Securely erase the shards from the")
			fmt.Println("       recovery directory after you encrypt them")
			fmt.Println("       and save them to a safe location.")
			fmt.Println("")
			fmt.Println(
				"  If you lose these shards, you will not be able to recover")
			fmt.Println(
				"  SPIKE Nexus in the unlikely event of a total system crash.")
			fmt.Println("")
		},
	}

	return recoverCmd
}
