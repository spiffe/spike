//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	sdk "github.com/spiffe/spike-sdk-go/api"
)

// encryptStream performs stream-based encryption by reading from a file or
// stdin and writing the encrypted ciphertext to a file or stdout.
//
// Parameters:
//   - api: The SPIKE SDK API client
//   - inFile: Input file path (empty string means stdin)
//   - outFile: Output file path (empty string means stdout)
//
// Returns:
//   - error: The first failure, including a failed close of the input or
//     output file; nil on success
func encryptStream(api *sdk.API, inFile, outFile string) (err error) {
	if checkErr := checkInputFile(inFile); checkErr != nil {
		return checkErr
	}

	in, cleanupIn, inputErr := openInput(inFile)
	if inputErr != nil {
		return inputErr
	}

	out, cleanupOut, outputErr := openOutput(outFile)
	if outputErr != nil {
		return errors.Join(outputErr, closeAll(cleanupIn))
	}
	defer func() {
		if closeErr := closeAll(cleanupOut, cleanupIn); closeErr != nil &&
			err == nil {
			err = closeErr
		}
	}()

	ciphertext, apiErr := api.CipherEncryptStream(context.Background(), in)
	if apiErr != nil {
		return cipherAPIError(apiErr)
	}

	if _, writeErr := out.Write(ciphertext); writeErr != nil {
		return fmt.Errorf("failed to write ciphertext: %w", writeErr)
	}
	return nil
}

// encryptJSON performs JSON-based encryption using base64-encoded plaintext
// and writes the encrypted result to a file or stdout.
//
// Parameters:
//   - api: The SPIKE SDK API client
//   - plaintextB64: Base64-encoded plaintext
//   - algorithm: Algorithm hint for encryption
//   - outFile: Output file path (empty string means stdout)
//
// Returns:
//   - error: The first failure, including a failed close of the output
//     file; nil on success
func encryptJSON(api *sdk.API, plaintextB64, algorithm,
	outFile string) (err error) {
	plaintext, decodeErr := base64.StdEncoding.DecodeString(plaintextB64)
	if decodeErr != nil {
		return errors.New("invalid --plaintext base64")
	}

	out, cleanupOut, openErr := openOutput(outFile)
	if openErr != nil {
		return openErr
	}
	defer func() {
		if closeErr := closeAll(cleanupOut); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	ciphertext, apiErr := api.CipherEncrypt(
		context.Background(), plaintext, algorithm,
	)
	if apiErr != nil {
		return cipherAPIError(apiErr)
	}

	if _, writeErr := out.Write(ciphertext); writeErr != nil {
		return fmt.Errorf("failed to write ciphertext: %w", writeErr)
	}
	return nil
}
