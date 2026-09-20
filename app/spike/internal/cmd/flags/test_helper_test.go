//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package flags

import (
	"testing"

	"github.com/spf13/cobra"
)

// newCommand creates a Cobra command with one string flag and one integer
// flag registered, so that the tests can exercise both helpers.
//
// Parameters:
//   - t: The test, used to mark this function as a helper
//
// Returns:
//   - *cobra.Command: A command with the "name" string flag (default
//     "default") and the "count" integer flag (default 7)
func newCommand(t *testing.T) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "default", "a string flag")
	cmd.Flags().Int("count", 7, "an integer flag")
	return cmd
}
