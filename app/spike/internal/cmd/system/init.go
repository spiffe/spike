//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package system

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	spike "github.com/spiffe/spike-sdk-go/api"
	"github.com/spiffe/spike-sdk-go/api/entity/data"

	"github.com/spiffe/spike/internal/config"
	"github.com/spiffe/spike/pkg/retry"
)

// NewSystemInitCommand creates and returns a new cobra.Command for initializing
// the SPIKE system. It configures a command that handles first-time setup by
// generating and storing authentication credentials.
//
// Parameters:
//   - source: X.509 source for workload API authentication
//
// Returns:
//   - *cobra.Command: Configured initialization command
//
// The command will:
//  1. Check if SPIKE is already initialized
//  2. If not initialized:
//     - Generate a new admin token
//     - Save the token using the provided X.509 source
//     - Store the token in SpikeAdminTokenFile()
//
// Error cases:
//   - Already initialized: Notifies user and exits
//   - Token save failure: Displays error message
//
// Note: This implementation is transitional. Future versions will:
//   - Replace admin token storage with temporary token exchange
//   - Integrate with 'pilot login' flow
//   - Include database configuration
func NewSystemInitCommand(source *workloadapi.X509Source) *cobra.Command {
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize spike configuration",
		Run: func(cmd *cobra.Command, args []string) {
			retrier := retry.NewExponentialRetrier()
			typedRetrier := retry.NewTypedRetrier[data.InitState](retrier)

			ctx := cmd.Context()
			state, err := typedRetrier.RetryWithBackoff(ctx,
				func() (data.InitState, error) {
					return spike.CheckInitState(source)
				})

			if err != nil {
				fmt.Println("Failed to check initialization state:")
				fmt.Println(err.Error())
				return
			}

			if state == data.AlreadyInitialized {
				fmt.Println("SPIKE is already initialized.")
				fmt.Println("Nothing to do.")
				return
			}

			err = retrier.RetryWithBackoff(ctx, func() error {
				return spike.Init(source)
			})

			if err != nil {
				fmt.Println("Failed to save admin token:")
				fmt.Println(err.Error())
				return
			}

			fmt.Println("")
			fmt.Println("    SPIKE system initialization completed.")
			fmt.Println("")
			fmt.Println("    >>> Recovery token saved to:")
			fmt.Printf("    >>> %s\n", config.SpikePilotRootKeyRecoveryFile())
			fmt.Println("")
			fmt.Println("    The recovery token is the only way to")
			fmt.Println("    recover SPIKE after a total system crash.")
			fmt.Println("    * Keep it SAFE and do not share it.")
			fmt.Println("    * ENCRYPT it and store it in a SECURE location.")
			fmt.Println("    * DO NOT KEEP IT ON THIS MACHINE.")
			fmt.Println("")
		},
	}

	return initCmd
}
