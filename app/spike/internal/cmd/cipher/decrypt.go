//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	sdk "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/spiffeid"

	"github.com/spiffe/spike/app/spike/internal/cmd/flags"
)

// newDecryptCommand creates a Cobra command for decrypting data via SPIKE
// Nexus. The command supports two modes of operation:
//
// Stream Mode (default):
//   - Reads encrypted data from a file (--file) or stdin
//   - Writes decrypted plaintext to a file (--out) or stdout
//   - Handles binary data transparently
//
// JSON Mode (when --version, --nonce, --ciphertext, or --input is provided):
//   - Accepts base64-encoded encryption components, or the JSON document
//     printed by "cipher encrypt" via --input
//   - Requires version byte (0-255), nonce, and ciphertext; the component
//     flags and --input are alternative ways to supply them and cannot be
//     combined with each other or with --file
//   - Returns plaintext output
//   - Allows algorithm specification via --algorithm flag, which pairs
//     with either form (--input plus --algorithm is valid)
//
// Parameters:
//   - source: SPIFFE X.509 SVID source for authentication. Can be nil if the
//     Workload API connection is unavailable. If nil, the command returns a
//     user-friendly error and the CLI exits non-zero.
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
//   - --input: JSON document printed by "cipher encrypt" for JSON mode;
//     cannot be combined with --file, --version, --nonce, or --ciphertext
//   - --algorithm: Algorithm hint for JSON mode
func newDecryptCommand(
	source *workloadapi.X509Source, SPIFFEID string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt file or stdin via SPIKE Nexus",
		RunE: func(cmd *cobra.Command, args []string) error {
			spiffeid.IsPilotOperatorOrDie(SPIFFEID)

			var (
				inFile, outFile, versionStr, nonceB64 string
				ciphertextB64, algorithm, inputJSON   string
				flagErr                               *sdkErrors.SDKError
			)
			if inFile, flagErr = flags.String(cmd, "file"); flagErr != nil {
				return flagErr
			}
			if outFile, flagErr = flags.String(cmd, "out"); flagErr != nil {
				return flagErr
			}
			versionStr, flagErr = flags.String(cmd, "version")
			if flagErr != nil {
				return flagErr
			}
			if nonceB64, flagErr = flags.String(cmd, "nonce"); flagErr != nil {
				return flagErr
			}
			ciphertextB64, flagErr = flags.String(cmd, "ciphertext")
			if flagErr != nil {
				return flagErr
			}
			inputJSON, flagErr = flags.String(cmd, "input")
			if flagErr != nil {
				return flagErr
			}
			algorithm, flagErr = flags.String(cmd, "algorithm")
			if flagErr != nil {
				return flagErr
			}

			// --input carries the whole document and is exclusive with
			// the component flags and --file: mixing them would resolve
			// per-field in ways the user cannot see. The conflicts are
			// rejected in a fixed order and before the X509 source check
			// so a bad invocation reports the actual mistake rather than
			// the missing source. --algorithm is not a conflict: the
			// document does not carry it, so it pairs with --input.
			if inputJSON != "" {
				conflicts := []struct {
					flag  string
					value string
				}{
					{"--file", inFile},
					{"--version", versionStr},
					{"--nonce", nonceB64},
					{"--ciphertext", ciphertextB64},
				}
				for _, conflict := range conflicts {
					if conflict.value != "" {
						return fmt.Errorf(
							"--input and %s cannot be combined",
							conflict.flag,
						)
					}
				}
				var parseErr error
				versionStr, nonceB64, ciphertextB64, parseErr =
					parseEncryptedJSON(inputJSON)
				if parseErr != nil {
					return parseErr
				}
			}

			if source == nil {
				return errors.New("SPIFFE X509 source is unavailable")
			}

			api := sdk.NewWithSource(source)

			jsonMode := versionStr != "" || nonceB64 != "" ||
				ciphertextB64 != ""

			if jsonMode {
				return decryptJSON(api, versionStr, nonceB64,
					ciphertextB64, algorithm, outFile)
			}

			return decryptStream(api, inFile, outFile)
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
		"input", "",
		"JSON document printed by 'encrypt' for JSON mode; exclusive "+
			"with --file, --version, --nonce and --ciphertext",
	)
	cmd.Flags().String(
		"algorithm", "", "Algorithm hint for JSON mode",
	)

	return cmd
}
