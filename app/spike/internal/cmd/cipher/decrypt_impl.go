//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	sdk "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/errors"
	"github.com/spiffe/spike/app/spike/internal/stdout"
)

// decryptStream performs stream-based decryption by reading from a file or
// stdin and writing the decrypted plaintext to a file or stdout.
//
// Parameters:
//   - api: The SPIKE SDK API client
//   - inFile: Input file path (empty string means stdin)
//   - outFile: Output file path (empty string means stdout)
//
// Returns:
//   - error if decryption or I/O operations fail
func decryptStream(api *sdk.API, inFile, outFile string) error {
	// Validate input file exists before attempting decryption.
	if inFile != "" {
		if _, err := os.Stat(inFile); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("input file does not exist: %s",
					inFile)
			}
			return fmt.Errorf("cannot access input file: %w", err)
		}
	}

	in, cleanupIn, err := openInput(inFile)
	if err != nil {
		return err
	}
	defer cleanupIn()

	out, cleanupOut, err := openOutput(outFile)
	if err != nil {
		return err
	}
	defer cleanupOut()

	plaintext, err := api.CipherDecryptStream(in,
		"application/octet-stream")
	if err != nil {
		if errors.NotReadyError(err) {
			stdout.PrintNotReady()
		}
		return fmt.Errorf("failed to call decrypt endpoint: %w", err)
	}

	if _, err := out.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// decryptJSON performs JSON-based decryption using base64-encoded components
// (version, nonce, ciphertext) and writes the decrypted plaintext to a file
// or stdout.
//
// Parameters:
//   - api: The SPIKE SDK API client
//   - versionStr: Version byte as a string (0-255)
//   - nonceB64: Base64-encoded nonce
//   - ciphertextB64: Base64-encoded ciphertext
//   - algorithm: Algorithm hint for decryption
//   - outFile: Output file path (empty string means stdout)
//
// Returns:
//   - error if validation, decryption, or I/O operations fail
func decryptJSON(api *sdk.API, versionStr, nonceB64, ciphertextB64,
	algorithm, outFile string) error {
	v, err := strconv.Atoi(versionStr)
	// version must be a valid byte value.
	if err != nil || v < 0 || v > 255 {
		return fmt.Errorf("invalid --version, must be 0-255")
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		return fmt.Errorf("invalid --nonce base64: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return fmt.Errorf("invalid --ciphertext base64: %w", err)
	}

	out, cleanupOut, err := openOutput(outFile)
	if err != nil {
		return err
	}
	defer cleanupOut()

	plaintext, err := api.CipherDecryptJSON(
		byte(v), nonce, ciphertext, algorithm,
	)
	if err != nil {
		if errors.NotReadyError(err) {
			stdout.PrintNotReady()
		}
		return fmt.Errorf(
			"failed to call decrypt endpoint (json): %w", err,
		)
	}

	if _, err := out.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write plaintext: %w", err)
	}

	return nil
}
