package policy

import (
	"fmt"
	"strings"
)

// normalizePolicyOutput is a helper function that converts raw CLI/table output into a normalized, test-friendly format.
//
// Parameters:
//   - output: the raw multi-line string produced by the policy list command (includes headers,
//     separators, and tabwriter-aligned columns).
//
// Returns:
//   - A string where each data row is rewritten into a predictable representation:
//     "ID: <id>, Name: <name>\n"
//
// It skips non-data lines such as "POLICIES", separators (e.g. "===="), blank lines,
// and the tabwriter column header ("ID NAME"),then parses each remaining data row to extract ID and Name.
// If a data row has only one field (because one column is visually empty), it uses the
// original line’s leading whitespace to infer whether the value belongs to the Name
// or the ID column
func normalizePolicyOutput(output string) string {
	lines := strings.Split(output, "\n")

	var b strings.Builder

	for _, raw := range lines {
		var id, name string

		line := strings.TrimSpace(raw)
		// Skip the header section (POLICIES, separator, blank line)
		if line == "" || line == "POLICIES" || strings.HasPrefix(line, "=") {
			continue
		}

		fields := strings.Fields(line)
		// Skip the tabwriter column header (ID NAME)
		if len(fields) == 2 && fields[0] == "ID" {
			continue
		}

		if len(fields) == 1 {
			// Determine if the field belongs to ID (no leading whitespace) or Name column (leading whitespace).
			// Check if the original raw line had leading whitespace (meaning ID column was empty).
			if len(raw) > len(strings.TrimLeft(raw, " ")) {
				name = fields[0]
			} else {
				id = fields[0]
			}
		} else {
			// The expected two-column format.
			id = fields[0]
			name = fields[1]
		}

		fmt.Fprintf(&b, "ID: %s, Name: %s\n", id, name)
	}

	return b.String()
}
