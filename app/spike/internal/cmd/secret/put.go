//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net/auth"
	"github.com/spiffe/spike/app/spike/internal/net/store"
	"github.com/spiffe/spike/internal/entity/data"
)

// NewSecretPutCommand creates and returns a new cobra.Command for storing secrets.
// It configures a command that stores key-value pairs as a secret at a
// specified path.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// Returns:
//   - *cobra.Command: Configured put command
//
// Arguments:
//  1. path: Location where the secret will be stored
//  2. key=value pairs: One or more key-value pairs in the format "key=value"
//
// Example Usage:
//
//	spike put secret/myapp username=admin password=secret
//	spike put secret/config host=localhost port=8080
//
// The command will:
//  1. Verify SPIKE initialization status via admin token
//  2. Parse all key-value pairs from arguments
//  3. Store the collected key-value pairs at the specified path
//
// Error cases:
//   - SPIKE not initialized: Prompts user to run 'spike initialization'
//   - Invalid key-value format: Reports the malformed pair
//   - Network/storage errors: Displays error message
//
// Note: Current admin token verification will be replaced with
// temporary token authentication in future versions
func NewSecretPutCommand(source *workloadapi.X509Source) *cobra.Command {
	var putCmd = &cobra.Command{
		Use:   "secret put <path> <key=value>...",
		Short: "Put secrets at the specified path",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			state, err := auth.CheckInitState(source)
			if err != nil {
				fmt.Println("Failed to check initialization state:")
				fmt.Println(err.Error())
				return
			}

			if state == data.NotInitialized {
				fmt.Println("Please initialize SPIKE first by running 'spike initialization'.")
				return
			}

			path := args[0]
			kvPairs := args[1:]
			values := make(map[string]string)
			for _, kv := range kvPairs {
				if !strings.Contains(kv, "=") {
					fmt.Printf("Error: invalid key-value pair format: %s\n", kv)
					continue
				}
				kvs := strings.Split(kv, "=")
				values[kvs[0]] = kvs[1]
			}

			if len(values) == 0 {
				fmt.Println("OK")
				return
			}

			err = store.PutSecret(source, path, values)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("OK")
		},
	}

	return putCmd
}