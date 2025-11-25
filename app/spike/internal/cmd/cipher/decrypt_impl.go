//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"encoding/base64"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	sdk "github.com/spiffe/spike-sdk-go/api"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/log"

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
	const fName = "decryptStream"

	// Validate input file exists before attempting decryption.
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

	plaintext, apiErr := api.CipherDecryptStream(in)
	if stdout.HandleAPIError(cmd, apiErr) {
		return
	}

	if _, err := out.Write(plaintext); err != nil {
		cmd.PrintErrf("Error: failed to write output: %v\n", err)
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "failed to write output"
		log.WarnErr(fName, *warnErr)
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
	const fName = "decryptJSON"

	v, err := strconv.Atoi(versionStr)
	// version must be a valid byte value.
	if err != nil || v < 0 || v > 255 {
		cmd.PrintErrln("Error: invalid --version, must be 0-255")
		warnErr := *sdkErrors.ErrDataInvalidInput.Clone()
		warnErr.Msg = "invalid --version, must be 0-255"
		log.WarnErr(fName, warnErr)
		return
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		cmd.PrintErrln("Error: invalid --nonce base64")
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "invalid --nonce base64"
		log.WarnErr(fName, *warnErr)
		return
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		cmd.PrintErrln("Error: invalid --ciphertext base64")
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "invalid --ciphertext base64"
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

	plaintext, apiErr := api.CipherDecrypt(
		byte(v), nonce, ciphertext, algorithm,
	)
	if stdout.HandleAPIError(cmd, apiErr) {
		return
	}

	if _, err := out.Write(plaintext); err != nil {
		cmd.PrintErrf("Error: failed to write plaintext: %v\n", err)
		warnErr := sdkErrors.ErrDataInvalidInput.Wrap(err)
		warnErr.Msg = "failed to write plaintext"
		log.WarnErr(fName, *warnErr)
		return
	}
}
