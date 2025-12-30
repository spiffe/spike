//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package format

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestOutputFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
		want   string
	}{
		{
			name:   "Human format",
			format: Human,
			want:   "human",
		},
		{
			name:   "JSON format",
			format: JSON,
			want:   "json",
		},
		{
			name:   "YAML format",
			format: YAML,
			want:   "yaml",
		},
		{
			name:   "Unknown format",
			format: OutputFormat(999),
			want:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.format.String(); got != tt.want {
				t.Errorf("OutputFormat.String() = %v, want %v",
					got, tt.want)
			}
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name      string
		formatStr string
		want      OutputFormat
		wantErr   bool
	}{
		// Human format aliases
		{
			name:      "Format 'human'",
			formatStr: "human",
			want:      Human,
			wantErr:   false,
		},
		{
			name:      "Format 'h'",
			formatStr: "h",
			want:      Human,
			wantErr:   false,
		},
		{
			name:      "Format 'plain'",
			formatStr: "plain",
			want:      Human,
			wantErr:   false,
		},
		{
			name:      "Format 'p'",
			formatStr: "p",
			want:      Human,
			wantErr:   false,
		},
		// JSON format aliases
		{
			name:      "Format 'json'",
			formatStr: "json",
			want:      JSON,
			wantErr:   false,
		},
		{
			name:      "Format 'j'",
			formatStr: "j",
			want:      JSON,
			wantErr:   false,
		},
		// YAML format aliases
		{
			name:      "Format 'yaml'",
			formatStr: "yaml",
			want:      YAML,
			wantErr:   false,
		},
		{
			name:      "Format 'y'",
			formatStr: "y",
			want:      YAML,
			wantErr:   false,
		},
		{
			name:      "Empty format defaults to human",
			formatStr: "",
			want:      Human,
			wantErr:   false,
		},
		// Invalid formats
		{
			name:      "Invalid format",
			formatStr: "invalid",
			want:      Human,
			wantErr:   true,
		},
		{
			name:      "Case sensitive - JSON uppercase",
			formatStr: "JSON",
			want:      Human,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormat(tt.formatStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat() error = %v, wantErr %v",
					err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddFormatFlag(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	AddFormatFlag(cmd)

	// Test that the flag was added
	flag := cmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatal("AddFormatFlag() did not add format flag")
	}

	// Test default value
	if flag.DefValue != "human" {
		t.Errorf("AddFormatFlag() default = %v, want %v",
			flag.DefValue, "human")
	}

	// Test shorthand
	shortFlag := cmd.Flags().ShorthandLookup("f")
	if shortFlag == nil {
		t.Fatal("AddFormatFlag() did not add shorthand flag")
	}
}

func TestGetFormat(t *testing.T) {
	tests := []struct {
		name      string
		flagValue string
		want      OutputFormat
		wantErr   bool
	}{
		{
			name:      "Get human format",
			flagValue: "human",
			want:      Human,
			wantErr:   false,
		},
		{
			name:      "Get json format",
			flagValue: "json",
			want:      JSON,
			wantErr:   false,
		},
		{
			name:      "Get yaml format",
			flagValue: "yaml",
			want:      YAML,
			wantErr:   false,
		},
		{
			name:      "Get format with alias 'p'",
			flagValue: "p",
			want:      Human,
			wantErr:   false,
		},
		{
			name:      "Get invalid format",
			flagValue: "invalid",
			want:      Human,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "test",
				Short: "Test command",
			}
			AddFormatFlag(cmd)
			_ = cmd.Flags().Set("format", tt.flagValue)

			got, err := GetFormat(cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFormat() error = %v, wantErr %v",
					err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
