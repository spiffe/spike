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

// newDecryptCommand creates a Cobra command for decrypting data via SPIKE
// Nexus. The command supports two modes of operation:
//
// Stream Mode (default):
//   - Reads encrypted data from a file (--file) or stdin
//   - Writes decrypted plaintext to a file (--out) or stdout
//   - Handles binary data transparently
//
// JSON Mode (when --version, --nonce, or --ciphertext is provided):
//   - Accepts base64-encoded encryption components
//   - Requires version byte (0-255), nonce, and ciphertext
//   - Returns plaintext output
//   - Allows algorithm specification via --algorithm flag
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable. If nil, the command will display
//     a user-friendly error message and exit cleanly.
//   - SPIFFEID: The SPIFFE ID to authenticate with
//
// Returns:
//   - *cobra.Command: Configured Cobra command for decryption
//
// Flags:
//   - --file, -f: Input file path (defaults to stdin)
//   - --out, -o: Output file path (defaults to stdout)
//   - --version: Version byte (0-255) for JSON mode
//   - --nonce: Base64-encoded nonce for JSON mode
//   - --ciphertext: Base64-encoded ciphertext for JSON mode
//   - --algorithm: Algorithm hint for JSON mode
func newDecryptCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt file or stdin via SPIKE Nexus",
		Run: func(cmd *cobra.Command, args []string) {
			trust.AuthenticateForPilot(SPIFFEID)

			if source == nil {
				cmd.PrintErrln("Error: SPIFFE X509 source is unavailable.")
				return
			}

			api := sdk.NewWithSource(source)

			inFile, _ := cmd.Flags().GetString("file")
			outFile, _ := cmd.Flags().GetString("out")
			versionStr, _ := cmd.Flags().GetString("version")
			nonceB64, _ := cmd.Flags().GetString("nonce")
			ciphertextB64, _ := cmd.Flags().GetString("ciphertext")
			algorithm, _ := cmd.Flags().GetString("algorithm")

			jsonMode := versionStr != "" || nonceB64 != "" ||
				ciphertextB64 != ""

			if jsonMode {
				decryptJSON(cmd, api, versionStr, nonceB64,
					ciphertextB64, algorithm, outFile)
				return
			}

			decryptStream(cmd, api, inFile, outFile)
		},
	}

	cmd.Flags().StringP(
		"file", "f", "", "Input file (default: stdin)",
	)
	cmd.Flags().StringP(
		"out", "o", "", "Output file (default: stdout)",
	)
	cmd.Flags().String(
		"version", "", "Version byte (0-255) for JSON mode",
	)
	cmd.Flags().String(
		"nonce", "", "Nonce (base64) for JSON mode",
	)
	cmd.Flags().String(
		"ciphertext", "", "Ciphertext (base64) for JSON mode",
	)
	cmd.Flags().String(
		"algorithm", "", "Algorithm hint for JSON mode",
	)

	return cmd
}
