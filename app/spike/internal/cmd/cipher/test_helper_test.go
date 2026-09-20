//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/spiffeid"
)

// pilotSPIFFEID returns a SPIFFE ID that passes the Pilot operator gate for
// the trust root configured in the test environment.
//
// Returns:
//   - string: The Pilot SPIFFE ID for the first configured Pilot trust root
func pilotSPIFFEID() string {
	trustRoots := env.TrustRootFromEnv(env.TrustRootPilot)
	first := strings.TrimSpace(strings.Split(trustRoots, ",")[0])
	return spiffeid.Pilot(first)
}

// newSilencedRoot mirrors the Pilot root command: handlers return errors,
// and Cobra's own error and usage output is silenced so that the root
// command can print the error once and exit non-zero.
//
// Parameters:
//   - sub: The command under test, attached as a child of the root
//
// Returns:
//   - *cobra.Command: A "spike" root command with SilenceErrors and
//     SilenceUsage set and sub registered as its subcommand
func newSilencedRoot(sub *cobra.Command) *cobra.Command {
	root := &cobra.Command{
		Use:           "spike",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(sub)
	return root
}
