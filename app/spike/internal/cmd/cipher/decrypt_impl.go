//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"

	sdk "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// checkInputFile verifies that the input file exists and is accessible.
// An empty path means stdin and needs no check.
//
// Parameters:
//   - inFile: Input file path (empty string means stdin)
//
// Returns:
//   - error: An error naming the file if it does not exist or cannot be
//     accessed; nil otherwise
func checkInputFile(inFile string) error {
	if inFile == "" {
		return nil
	}
	if _, err := os.Stat(inFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input file does not exist: %s", inFile)
		}
		return fmt.Errorf("cannot access input file: %s: %w", inFile, err)
	}
	return nil
}

// closeAll runs the cleanup functions in order and returns the first close
// error, if any. A failed close on an output file can mean that not all of
// the written data reached the disk, so it is a failure of the command.
//
// Parameters:
//   - cleanups: The cleanup functions to run, in order; every one of them
//     runs even when an earlier one fails
//
// Returns:
//   - error: The first close error encountered; nil when every cleanup
//     succeeded
func closeAll(cleanups ...func() *sdkErrors.SDKError) error {
	var first error
	for _, cleanup := range cleanups {
		if closeErr := cleanup(); closeErr != nil && first == nil {
			first = closeErr
		}
	}
	return first
}

// decryptStream performs stream-based decryption by reading from a file or
// stdin and writing the decrypted plaintext to a file or stdout.
//
// Parameters:
//   - api: The SPIKE SDK API client
//   - inFile: Input file path (empty string means stdin)
//   - outFile: Output file path (empty string means stdout)
//
// Returns:
//   - error: The first failure, including a failed close of the input or
//     output file; nil on success
func decryptStream(api *sdk.API, inFile, outFile string) (err error) {
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

	plaintext, apiErr := api.CipherDecryptStream(context.Background(), in)
	if apiErr != nil {
		return cipherAPIError(apiErr)
	}

	if _, writeErr := out.Write(plaintext); writeErr != nil {
		return fmt.Errorf("failed to write output: %w", writeErr)
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
//   - error: The first failure, including a failed close of the output
//     file; nil on success
func decryptJSON(api *sdk.API, versionStr, nonceB64, ciphertextB64,
	algorithm, outFile string) (err error) {
	v, atoiErr := strconv.Atoi(versionStr)
	// The version must be a valid byte value.
	if atoiErr != nil || v < 0 || v > 255 {
		return errors.New("invalid --version, must be 0-255")
	}

	nonce, nonceErr := base64.StdEncoding.DecodeString(nonceB64)
	if nonceErr != nil {
		return errors.New("invalid --nonce base64")
	}

	ciphertext, ciphertextErr := base64.StdEncoding.DecodeString(ciphertextB64)
	if ciphertextErr != nil {
		return errors.New("invalid --ciphertext base64")
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

	plaintext, apiErr := api.CipherDecrypt(
		context.Background(), byte(v), nonce, ciphertext, algorithm,
	)
	if apiErr != nil {
		return cipherAPIError(apiErr)
	}

	if _, writeErr := out.Write(plaintext); writeErr != nil {
		return fmt.Errorf("failed to write plaintext: %w", writeErr)
	}
	return nil
}
