//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	sdk "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/trust"
)

// newEncryptCommand creates a Cobra command for encrypting data via SPIKE
// Nexus. The command supports two modes of operation:
//
// Stream Mode (default):
//   - Reads data from a file (--file) or stdin
//   - Writes encrypted data to a file (--out) or stdout
//   - Handles binary data transparently
//
// JSON Mode (when --plaintext is provided):
//   - Accepts base64-encoded plaintext
//   - Returns JSON-formatted encryption result
//   - Allows algorithm specification via --algorithm flag
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable. If nil, the command will display
//     a user-friendly error message and exit cleanly.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for encryption
//
// Flags:
//   - --file, -f: Input file path (defaults to stdin)
//   - --out, -o: Output file path (defaults to stdout)
//   - --plaintext: Base64-encoded plaintext for JSON mode
//   - --algorithm: Algorithm hint for JSON mode
func newEncryptCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt file or stdin via SPIKE Nexus",
		RunE: func(cmd *cobra.Command, args []string) error {
			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable")
				cmd.PrintErrln("The workload API may have lost connection.")
				cmd.PrintErrln("Please check your SPIFFE agent and try again.")
				return nil
			}

			trust.AuthenticateForPilot(SPIFFEID)

			api := sdk.NewWithSource(source)

			inFile, _ := cmd.Flags().GetString("file")
			outFile, _ := cmd.Flags().GetString("out")
			plaintextB64, _ := cmd.Flags().GetString("plaintext")
			algorithm, _ := cmd.Flags().GetString("algorithm")

			if plaintextB64 != "" {
				return encryptJSON(api, plaintextB64, algorithm, outFile)
			}

			return encryptStream(api, inFile, outFile)
		},
	}

	cmd.Flags().StringP(
		"file", "f", "", "Input file (default: stdin)",
	)
	cmd.Flags().StringP(
		"out", "o", "", "Output file (default: stdout)",
	)
	cmd.Flags().String(
		"plaintext", "",
		"Base64 plaintext for JSON mode; if set, uses JSON API",
	)
	cmd.Flags().String(
		"algorithm", "", "Algorithm hint for JSON mode",
	)

	return cmd
}
