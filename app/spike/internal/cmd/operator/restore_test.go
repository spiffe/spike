//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestReadShardInput_NonInteractive verifies that readShardInput reads a
// shard from a non-terminal stdin (a pipe), which is what allows scripts
// to drive `spike operator restore`.
func TestReadShardInput_NonInteractive(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain shard line",
			input: "spike:1:" + strings.Repeat("ab", 32) + "\n",
			want:  "spike:1:" + strings.Repeat("ab", 32),
		},
		{
			name:  "surrounding whitespace is trimmed",
			input: "  spike:2:" + strings.Repeat("cd", 32) + "  \n\n",
			want:  "spike:2:" + strings.Repeat("cd", 32),
		},
		{
			name:  "no trailing newline",
			input: "spike:3:" + strings.Repeat("ef", 32),
			want:  "spike:3:" + strings.Repeat("ef", 32),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, pipeErr := os.Pipe()
			if pipeErr != nil {
				t.Fatalf("failed to create a pipe: %v", pipeErr)
				return
			}

			original := os.Stdin
			os.Stdin = r
			t.Cleanup(func() {
				os.Stdin = original
				_ = r.Close()
			})

			if _, writeErr := w.WriteString(tt.input); writeErr != nil {
				t.Fatalf("failed to write to the pipe: %v", writeErr)
				return
			}
			if closeErr := w.Close(); closeErr != nil {
				t.Fatalf("failed to close the pipe writer: %v", closeErr)
				return
			}

			cmd := &cobra.Command{Use: "test"}

			got, readErr := readShardInput(cmd)
			if readErr != nil {
				t.Fatalf("readShardInput() error = %v", readErr)
				return
			}

			if string(got) != tt.want {
				t.Errorf("readShardInput() = %q, want %q",
					string(got), tt.want)
			}
		})
	}
}
