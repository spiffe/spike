package policy

import (
	"strings"
)

// normalizePolicyOutput is a helper function that normalizes policy output
// for testing.
//
// Parameters:
//   - output: the raw multi-line string produced by the policy list command
//
// Returns:
//   - A normalized string with trailing whitespace removed from each line
//
// It skips non-data lines such as "POLICIES", separators (e.g. "===="),
// and blank lines.
func normalizePolicyOutput(output string) string {
	lines := strings.Split(output, "\n")

	var b strings.Builder

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		// Skip the header section (POLICIES, separator, blank line, dashes)
		if line == "" || line == "POLICIES" ||
			strings.HasPrefix(line, "=") || strings.HasPrefix(line, "-") {
			continue
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
