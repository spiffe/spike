//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"github.com/spf13/cobra"

	"github.com/spiffe/spike/app/spike/internal/cmd/format"
)

// addFormatFlag adds a standardized format flag to the given command.
// Supports human/h/plain/p, json/j, and yaml/y formats.
//
// Parameters:
//   - cmd: The Cobra command to add the flag to
func addFormatFlag(cmd *cobra.Command) {
	format.AddFormatFlag(cmd)
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
