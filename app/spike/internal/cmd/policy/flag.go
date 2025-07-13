//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import "github.com/spf13/cobra"

// addFormatFlag adds a format flag to the given command to allow specifying
// the output format (human or JSON).
//
// Parameters:
//   - cmd: The Cobra command to add the flag to
func addFormatFlag(cmd *cobra.Command) {
	cmd.Flags().String("format", "human",
		"Output format: 'human' or 'json'")
}

// addNameFlag adds a name flag to the given command to allow specifying
// a policy by name instead of by ID.
//
// Parameters:
//   - cmd: The Cobra command to add the flag to
func addNameFlag(cmd *cobra.Command) {
	cmd.Flags().String("name", "",
		"Policy name to look up (alternative to policy ID)")
}
