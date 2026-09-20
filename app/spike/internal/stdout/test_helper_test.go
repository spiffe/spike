//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package stdout

import (
	"github.com/spf13/cobra"
)

// createTestCommand creates a bare Cobra command with no parent.
//
// Parameters:
//   - commandPath: The value used as the command's Use string
//
// Returns:
//   - *cobra.Command: A command whose path has a single element, so that
//     no command group is detected
func createTestCommand(commandPath string) *cobra.Command {
	return &cobra.Command{Use: commandPath}
}

// createTestCommandWithParent creates a Cobra command hierarchy for testing
// command group detection (for example, "spike cipher encrypt").
//
// Parameters:
//   - group: The command group name, such as "cipher" or "policy"
//   - subcommand: The leaf command name under the group
//
// Returns:
//   - *cobra.Command: The leaf command, attached to the group under a
//     "spike" root
func createTestCommandWithParent(group, subcommand string) *cobra.Command {
	root := &cobra.Command{Use: "spike"}
	groupCmd := &cobra.Command{Use: group}
	subCmd := &cobra.Command{Use: subcommand}

	root.AddCommand(groupCmd)
	groupCmd.AddCommand(subCmd)

	return subCmd
}
