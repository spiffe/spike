//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package stdout

import (
	"github.com/spf13/cobra"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

const commandGroupPolicy = "policy"
const commandGroupCipher = "cipher"

// HandleAPIError processes API errors and prints appropriate user-friendly
// messages. It detects the command group (policy, secret, cipher) from the
// Cobra command path and handles group-specific errors accordingly.
//
// Parameters:
//   - c: Cobra command for output and command group detection
//   - err: The error returned from an API call
//
// Returns:
//   - bool: true if an error was handled, false if no error
//
// Common error types handled for all command groups:
//   - ErrStateNotReady: System not initialized
//   - ErrDataMarshalFailure: Request serialization failure
//   - ErrDataUnmarshalFailure: Response parsing failure
//   - ErrAPINotFound: Resource not found
//   - ErrAPIBadRequest: Invalid request parameters
//   - ErrDataInvalidInput: Input validation failure
//   - ErrNetPeerConnection: Network connection failure
//   - ErrAccessUnauthorized: Permission denied
//   - ErrNetReadingResponseBody: Response read failure
//
// Policy-specific errors:
//   - ErrEntityNotFound, ErrEntityInvalid, ErrAPIPostFailed,
//     ErrEntityCreationFailed
//
// Cipher-specific errors:
//   - ErrCryptoEncryptionFailed, ErrCryptoDecryptionFailed,
//     ErrCryptoCipherNotAvailable, ErrCryptoInvalidEncryptionKeyLength
//
// For any unhandled error types, the function falls back to displaying
// the SDK error message directly.
//
// Usage example:
//
//	secret, err := api.GetSecretVersion(path, version)
//	if stdout.HandleAPIError(cmd, err) {
//	    return
//	}
func HandleAPIError(c *cobra.Command, err *sdkErrors.SDKError) bool {
	if err == nil {
		return false
	}

	// Common errors (all command groups)
	switch {
	case err.Is(sdkErrors.ErrStateNotReady):
		PrintNotReady()
		return true
	case err.Is(sdkErrors.ErrDataMarshalFailure):
		c.PrintErrln("Error: Malformed request.")
		return true
	case err.Is(sdkErrors.ErrDataUnmarshalFailure):
		c.PrintErrln("Error: Failed to parse API response.")
		return true
	case err.Is(sdkErrors.ErrAPINotFound):
		c.PrintErrln("Error: Resource not found.")
		return true
	case err.Is(sdkErrors.ErrAPIBadRequest):
		c.PrintErrln("Error: Invalid request.")
		return true
	case err.Is(sdkErrors.ErrDataInvalidInput):
		c.PrintErrln("Error: Invalid input provided.")
		return true
	case err.Is(sdkErrors.ErrNetPeerConnection):
		c.PrintErrln("Error: Failed to connect to SPIKE Nexus.")
		return true
	case err.Is(sdkErrors.ErrAccessUnauthorized):
		c.PrintErrln("Error: Unauthorized access.")
		return true
	case err.Is(sdkErrors.ErrNetReadingResponseBody):
		c.PrintErrln("Error: Failed to read response body.")
		return true
	}

	// Command-group-specific errors
	group := getCommandGroup(c)
	switch group {
	case commandGroupPolicy:
		if handlePolicyError(c, err) {
			return true
		}
	case commandGroupCipher:
		if handleCipherError(c, err) {
			return true
		}
	}

	// Fallback for any unhandled errors
	c.PrintErrf("Error: %v\n", err)
	return true
}
