//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/errors"
	"github.com/spiffe/spike/app/spike/internal/stdout"
	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newSecretPutCommand creates and returns a new cobra.Command for storing
// secrets. It configures a command that stores key-value pairs as a secret at a
// specified path.
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable, in which case the command will
//     display an error message and return.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured put command
//
// Arguments:
//  1. path: Location where the secret will be stored (namespace format, no
//     leading slash)
//  2. key=value pairs: One or more key-value pairs in the format "key=value"
//
// Example Usage:
//
//	spike secret put secret/myapp username=admin password=secret
//	spike secret put secret/config host=localhost port=8080
//
// The command execution flow:
//  1. Verify X509 source is available (workload API connection active)
//  2. Authenticate the pilot using SPIFFE ID
//  3. Validate the secret path format
//  4. Parse all key-value pairs from arguments
//  5. Store the key-value pairs at the specified path via SPIKE API
//
// Error cases:
//   - X509 source unavailable: Workload API connection lost
//   - Invalid secret path: Path format validation failed
//   - Invalid key-value format: Malformed pair (continues with other pairs)
//   - SPIKE not ready: Backend not initialized, prompts to wait
//   - Network/API errors: Connection or storage failures
func newSecretPutCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	var putCmd = &cobra.Command{
		Use:   "put <path> <key=value>...",
		Short: "Put secrets at the specified path",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable")
				cmd.PrintErrln("The workload API may have lost connection.")
				cmd.PrintErrln("Please check your SPIFFE agent and try again.")
				return
			}

			trust.AuthenticateForPilot(SPIFFEID)

			api := spike.NewWithSource(source)

			path := args[0]

			if !validSecretPath(path) {
				cmd.PrintErrf("Error: invalid secret path: %s\n", path)
				return
			}

			kvPairs := args[1:]
			values := make(map[string]string)
			for _, kv := range kvPairs {
				if !strings.Contains(kv, "=") {
					cmd.PrintErrf("Error: invalid key-value pair format: %s\n", kv)
					continue
				}
				kvs := strings.Split(kv, "=")
				values[kvs[0]] = kvs[1]
			}

			if len(values) == 0 {
				cmd.Println("OK")
				return
			}

			err := api.PutSecret(path, values)
			if err != nil {
				if errors.NotReadyError(err) {
					stdout.PrintNotReady()
					return
				}

				cmd.PrintErrf("Error: %v\n", err)
				return
			}

			cmd.Println("OK")
		},
	}

	return putCmd
}
