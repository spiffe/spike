//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"context"
	"encoding/base64"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	sdk "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// decryptStream performs stream-based decryption by reading from a file or
// stdin and writing the decrypted plaintext to a file or stdout.
//
// Parameters:
//   - cmd: Cobra command for output
//   - api: The SPIKE SDK API client
//   - inFile: Input file path (empty string means stdin)
//   - outFile: Output file path (empty string means stdout)
//
// The function prints errors directly to stderr and returns without error
// propagation, following the CLI command pattern.
func decryptStream(cmd *cobra.Command, api *sdk.API, inFile, outFile string) {
	// Validate the input file exists before attempting decryption.
	if inFile != "" {
		if _, err := os.Stat(inFile); err != nil {
			if os.IsNotExist(err) {
				cmd.PrintErrf("Error: Input file does not exist: %s\n", inFile)
				return
			}
			cmd.PrintErrf("Error: Cannot access input file: %s\n", inFile)
			return
		}
	}

	in, cleanupIn, inputErr := openInput(inFile)
	if inputErr != nil {
		cmd.PrintErrf("Error: %v\n", inputErr)
		return
	}
	defer cleanupIn()

	out, cleanupOut, outputErr := openOutput(outFile)
	if outputErr != nil {
		cmd.PrintErrf("Error: %v\n", outputErr)
		return
	}
	defer cleanupOut()

	ctx := context.Background()

	plaintext, apiErr := api.CipherDecryptStream(ctx, in)
	if stdout.HandleAPIError(cmd, apiErr) {
		return
	}

	if _, writeErr := out.Write(plaintext); writeErr != nil {
		cmd.PrintErrf("Error: Failed to write output: %v\n", writeErr)
		return
	}
}

// decryptJSON performs JSON-based decryption using base64-encoded components
// (version, nonce, ciphertext) and writes the decrypted plaintext to a file
// or stdout.
//
// Parameters:
//   - cmd: Cobra command for output
//   - api: The SPIKE SDK API client
//   - versionStr: Version byte as a string (0-255)
//   - nonceB64: Base64-encoded nonce
//   - ciphertextB64: Base64-encoded ciphertext
//   - algorithm: Algorithm hint for decryption
//   - outFile: Output file path (empty string means stdout)
//
// The function prints errors directly to stderr and returns without error
// propagation, following the CLI command pattern.
func decryptJSON(cmd *cobra.Command, api *sdk.API, versionStr, nonceB64,
	ciphertextB64, algorithm, outFile string) {
	v, atoiErr := strconv.Atoi(versionStr)
	// version must be a valid byte value.
	if atoiErr != nil || v < 0 || v > 255 {
		cmd.PrintErrln("Error: Invalid --version, must be 0-255.")
		return
	}

	nonce, nonceErr := base64.StdEncoding.DecodeString(nonceB64)
	if nonceErr != nil {
		cmd.PrintErrln("Error: Invalid --nonce base64.")
		return
	}

	ciphertext, ciphertextErr := base64.StdEncoding.DecodeString(ciphertextB64)
	if ciphertextErr != nil {
		cmd.PrintErrln("Error: Invalid --ciphertext base64.")
		return
	}

	out, cleanupOut, openErr := openOutput(outFile)
	if openErr != nil {
		cmd.PrintErrf("Error: %v\n", openErr)
		return
	}
	defer cleanupOut()

	ctx := context.Background()

	plaintext, apiErr := api.CipherDecrypt(
		ctx, byte(v), nonce, ciphertext, algorithm,
	)
	if stdout.HandleAPIError(cmd, apiErr) {
		return
	}

	if _, writeErr := out.Write(plaintext); writeErr != nil {
		cmd.PrintErrf("Error: Failed to write plaintext: %v\n", writeErr)
		return
	}
}
