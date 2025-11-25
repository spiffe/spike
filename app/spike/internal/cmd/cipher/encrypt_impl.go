//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"encoding/base64"
	"os"

	"github.com/spf13/cobra"
	sdk "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

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
	const fName = "encryptStream"

	// Validate input file exists before attempting encryption.
	if inFile != "" {
		if _, err := os.Stat(inFile); err != nil {
			if os.IsNotExist(err) {
				cmd.PrintErrf("Error: input file does not exist: %s\n",
					inFile)
				warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
				warnErr.Msg = "input file does not exist"
				log.WarnErr(fName, *warnErr)
				return
			}

			cmd.PrintErrf("Error: cannot access input file: %s\n", inFile)
			warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
			warnErr.Msg = "cannot access input file"
			log.WarnErr(fName, *warnErr)
			return
		}
	}

	in, cleanupIn, err := openInput(inFile)
	if err != nil {
		cmd.PrintErrf("Error: %v\n", err)
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "failed to open input"
		log.WarnErr(fName, *warnErr)
		return
	}
	defer cleanupIn()

	out, cleanupOut, err := openOutput(outFile)
	if err != nil {
		cmd.PrintErrf("Error: %v\n", err)
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "failed to open output"
		log.WarnErr(fName, *warnErr)
		return
	}
	defer cleanupOut()

	ciphertext, apiErr := api.CipherEncryptStream(in)
	if stdout.HandleAPIError(cmd, apiErr) {
		return
	}

	if _, err := out.Write(ciphertext); err != nil {
		cmd.PrintErrf("Error: failed to write ciphertext: %v\n", err)
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "failed to write ciphertext"
		log.WarnErr(fName, *warnErr)
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
	const fName = "encryptJSON"

	plaintext, err := base64.StdEncoding.DecodeString(plaintextB64)
	if err != nil {
		cmd.PrintErrln("Error: invalid --plaintext base64")
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "invalid --plaintext base64"
		log.WarnErr(fName, *warnErr)
		return
	}

	out, cleanupOut, err := openOutput(outFile)
	if err != nil {
		cmd.PrintErrf("Error: %v\n", err)
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "failed to open output"
		log.WarnErr(fName, *warnErr)
		return
	}
	defer cleanupOut()

	ciphertext, apiErr := api.CipherEncrypt(plaintext, algorithm)
	if stdout.HandleAPIError(cmd, apiErr) {
		return
	}

	if _, err := out.Write(ciphertext); err != nil {
		cmd.PrintErrf("Error: failed to write ciphertext: %v\n", err)
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "failed to write ciphertext"
		log.WarnErr(fName, *warnErr)
		return
	}
}
