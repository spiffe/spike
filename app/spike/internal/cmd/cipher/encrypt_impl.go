//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"context"
	"encoding/base64"
	"os"

	"github.com/spf13/cobra"
	sdk "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// encryptStream performs stream-based encryption by reading from a file or
// stdin and writing the encrypted ciphertext to a file or stdout.
//
// Parameters:
//   - cmd: Cobra command for output
//   - api: The SPIKE SDK API client
//   - inFile: Input file path (empty string means stdin)
//   - outFile: Output file path (empty string means stdout)
//
// The function prints errors directly to stderr and returns without error
// propagation, following the CLI command pattern.
func encryptStream(cmd *cobra.Command, api *sdk.API, inFile, outFile string) {
	// Validate the input file exists before attempting encryption.
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
	defer cleanupIn() // safe: openInput returns noop on error.
	if inputErr != nil {
		cmd.PrintErrf("Error: %v\n", inputErr)
		return
	}

	out, cleanupOut, outputErr := openOutput(outFile)
	if outputErr != nil {
		cmd.PrintErrf("Error: %v\n", outputErr)
		return
	}
	defer cleanupOut()

	ctx := context.Background()

	ciphertext, apiErr := api.CipherEncryptStream(ctx, in)
	if stdout.HandleAPIError(cmd, apiErr) {
		return
	}

	if _, writeErr := out.Write(ciphertext); writeErr != nil {
		cmd.PrintErrf("Error: Failed to write ciphertext: %v\n", writeErr)
		return
	}
}

// encryptJSON performs JSON-based encryption using base64-encoded plaintext
// and writes the encrypted result to a file or stdout.
//
// Parameters:
//   - cmd: Cobra command for output
//   - api: The SPIKE SDK API client
//   - plaintextB64: Base64-encoded plaintext
//   - algorithm: Algorithm hint for encryption
//   - outFile: Output file path (empty string means stdout)
//
// The function prints errors directly to stderr and returns without error
// propagation, following the CLI command pattern.
func encryptJSON(cmd *cobra.Command, api *sdk.API, plaintextB64, algorithm,
		outFile string) {
	plaintext, err := base64.StdEncoding.DecodeString(plaintextB64)
	if err != nil {
		cmd.PrintErrln("Error: Invalid --plaintext base64.")
		return
	}

	out, cleanupOut, openErr := openOutput(outFile)
	if openErr != nil {
		cmd.PrintErrf("Error: %v\n", openErr)
		return
	}
	defer cleanupOut()

	ctx := context.Background()

	ciphertext, apiErr := api.CipherEncrypt(ctx, plaintext, algorithm)
	if stdout.HandleAPIError(cmd, apiErr) {
		return
	}

	if _, writeErr := out.Write(ciphertext); writeErr != nil {
		cmd.PrintErrf("Error: Failed to write ciphertext: %v\n", writeErr)
		return
	}
}
