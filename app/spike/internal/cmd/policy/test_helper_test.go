//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package policy

import (
	"strings"

	"github.com/spf13/cobra"
)

// createTestCommandWithFormat creates a Cobra command that carries a
// --format flag preset to the given value, so that formatting tests can
// exercise every output format without parsing arguments.
//
// Parameters:
//   - format: The default value of the "format" flag, such as "human",
//     "json", or "yaml".
//
// Returns:
//   - *cobra.Command: A bare "test" command with the flag registered.
func createTestCommandWithFormat(format string) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("format", format, "Output format")
	return cmd
}

// normalizePolicyOutput normalizes the human-readable policy list output so
// that tests can compare it without depending on layout details.
//
// It trims each line, skips non-data lines such as "POLICIES", separators
// (for example "===="), and blank lines, and joins what remains with
// newlines.
//
// Parameters:
//   - output: The raw multi-line string produced by the policy list command
//
// Returns:
//   - string: The kept lines, each followed by a newline, or an empty
//     string when no data lines remain
func normalizePolicyOutput(output string) string {
	var kept []string

	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		// Skip the header section (POLICIES, separator, blank line, dashes)
		if line == "" || line == "POLICIES" ||
			strings.HasPrefix(line, "=") || strings.HasPrefix(line, "-") {
			continue
		}

		kept = append(kept, line)
	}

	if len(kept) == 0 {
		return ""
	}

	return strings.Join(kept, "\n") + "\n"
}
