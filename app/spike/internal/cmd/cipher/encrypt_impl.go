//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	sdk "github.com/spiffe/spike-sdk-go/api"

	"github.com/spiffe/spike/app/spike/internal/errors"
	"github.com/spiffe/spike/app/spike/internal/stdout"
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
//   - error if encryption or I/O operations fail
func encryptStream(api *sdk.API, inFile, outFile string) error {
	// Validate input file exists before attempting encryption.
	if inFile != "" {
		if _, err := os.Stat(inFile); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("input file does not exist: %s",
					inFile)
			}
			return fmt.Errorf("cannot access input file: %w", err)
		}
	}

	var in io.ReadCloser
	if inFile != "" {
		f, err := os.Open(inFile)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		in = f
	} else {
		in = os.Stdin
	}
	defer func() {
		if in != os.Stdin {
			err := in.Close()
			if err != nil {
				fmt.Printf("Failed to close input file: %s\n",
					err.Error())
			}
		}
	}()

	var out io.Writer
	var outCloser io.Closer
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		out = f
		outCloser = f
	} else {
		out = os.Stdout
	}
	if outCloser != nil {
		defer func(outCloser io.Closer) {
			err := outCloser.Close()
			if err != nil {
				fmt.Printf("Failed to close output file: %s\n",
					err.Error())
			}
		}(outCloser)
	}

	ciphertext, err := api.CipherEncryptStream(in,
		"application/octet-stream")
	if err != nil {
		if errors.NotReadyError(err) {
			stdout.PrintNotReady()
		}
		return fmt.Errorf("failed to call encrypt endpoint: %w", err)
	}

	if _, err := out.Write(ciphertext); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
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
//   - error if validation, encryption, or I/O operations fail
func encryptJSON(api *sdk.API, plaintextB64, algorithm,
	outFile string) error {
	plaintext, err := base64.StdEncoding.DecodeString(plaintextB64)
	if err != nil {
		return fmt.Errorf("invalid --plaintext base64: %w", err)
	}

	var out io.Writer
	var outCloser io.Closer
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		out = f
		outCloser = f
	} else {
		out = os.Stdout
	}
	if outCloser != nil {
		defer func(outCloser io.Closer) {
			err := outCloser.Close()
			if err != nil {
				fmt.Printf("Failed to close output file: %s\n",
					err.Error())
			}
		}(outCloser)
	}

	ciphertext, err := api.CipherEncryptJSON(plaintext, algorithm)
	if err != nil {
		if errors.NotReadyError(err) {
			stdout.PrintNotReady()
		}
		return fmt.Errorf("failed to call encrypt endpoint (json): %w",
			err)
	}

	if _, err := out.Write(ciphertext); err != nil {
		return fmt.Errorf("failed to write ciphertext: %w", err)
	}

	return nil
}
