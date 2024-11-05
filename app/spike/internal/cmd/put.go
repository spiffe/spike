//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net"
	"github.com/spiffe/spike/app/spike/internal/state"
)

func NewPutCommand(source *workloadapi.X509Source) *cobra.Command {
	var putCmd = &cobra.Command{
		Use:   "put <path> <key=value>...",
		Short: "Put secrets at the specified path",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("######### WILL CHECK TEMP TOKEN INSTEAD #####")

			adminToken, err := state.AdminToken()
			if err != nil {
				fmt.Println("")
				fmt.Println("SPIKE is not initialized.")
				fmt.Println("Please run `spike init` to initialize SPIKE.")
				return
			}

			if adminToken == "" {
				fmt.Println("")
				fmt.Println("SPIKE is not initialized.")
				fmt.Println("Please run `spike init` to initialize SPIKE.")
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
			}

			err = net.PutSecret(source, path, values)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			fmt.Println("OK")
		},
	}

	return putCmd
}
