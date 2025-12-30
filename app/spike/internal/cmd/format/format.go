//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package format

import (
	"fmt"

	"github.com/spf13/cobra"
)

// OutputFormat represents the supported output formats.
type OutputFormat int

const (
	// Human represents human-readable output format.
	Human OutputFormat = iota
	// JSON represents JSON output format.
	JSON
	// YAML represents YAML output format.
	YAML
)

// String returns the canonical string representation of the format.
func (f OutputFormat) String() string {
	switch f {
	case Human:
		return "human"
	case JSON:
		return "json"
	case YAML:
		return "yaml"
	default:
		return "unknown"
	}
}

// AddFormatFlag adds a standardized format flag to the given command.
// The format flag supports the following options:
//   - human, h, plain, p: Human-readable, friendly output (default)
//   - json, j: Valid JSON output (for scripting/parsing)
//   - yaml, y: Valid YAML output (for scripting/parsing)
//
// Parameters:
//   - cmd: The Cobra command to add the flag to
func AddFormatFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("format", "f", "human",
		"Output format: human/h/plain/p, json/j, or yaml/y")
}

// GetFormat retrieves and validates the format flag from the command.
// It supports multiple aliases for each format:
//   - human, h, plain, p -> Human format
//   - json, j -> JSON format
//   - yaml, y -> YAML format
//
// Parameters:
//   - cmd: The Cobra command containing the format flag
//
// Returns:
//   - OutputFormat: The parsed output format
//   - error: An error if the format is invalid
func GetFormat(cmd *cobra.Command) (OutputFormat, error) {
	formatStr, _ := cmd.Flags().GetString("format")
	return ParseFormat(formatStr)
}

// ParseFormat parses a format string into an OutputFormat.
// It supports multiple aliases for each format:
//   - human, h, plain, p, "" (empty) -> Human format
//   - json, j -> JSON format
//   - yaml, y -> YAML format
//
// Parameters:
//   - formatStr: The format string to parse
//
// Returns:
//   - OutputFormat: The parsed output format
//   - error: An error if the format is invalid
func ParseFormat(formatStr string) (OutputFormat, error) {
	switch formatStr {
	case "human", "h", "plain", "p", "":
		return Human, nil
	case "json", "j":
		return JSON, nil
	case "yaml", "y":
		return YAML, nil
	default:
		return Human, fmt.Errorf(
			"invalid format '%s'. Valid formats are: "+
				"human/h/plain/p, json/j, yaml/y",
			formatStr)
	}
}
