//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"bytes"

	"github.com/spf13/cobra"
)

// createTestCommandWithBuffer creates a Cobra command whose standard output
// is captured in a buffer, so that tests can assert on what a command
// prints.
//
// Returns:
//   - *cobra.Command: A bare "test" command with its output redirected.
//   - *bytes.Buffer: The buffer that receives the command's output.
func createTestCommandWithBuffer() (*cobra.Command, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(buf)
	return cmd, buf
}
