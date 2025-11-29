//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package stdout

import (
	"strings"

	"github.com/spf13/cobra"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

// getCommandGroup extracts the command group from the Cobra command path.
// For example, "spike cipher encrypt" returns "cipher".
//
// Parameters:
//   - c: Cobra command to extract the group from
//
// Returns:
//   - string: The command group name (e.g., "cipher", "secret", "policy"),
//     or an empty string if the command path has fewer than 2 parts
func getCommandGroup(c *cobra.Command) string {
	parts := strings.Fields(c.CommandPath())
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// handlePolicyError handles policy-specific SDK errors.
//
// Parameters:
//   - c: Cobra command for error output
//   - err: The SDK error to check and handle
//
// Returns:
//   - bool: true if the error was a policy-specific error and was handled,
//     false otherwise
func handlePolicyError(c *cobra.Command, err *sdkErrors.SDKError) bool {
	switch {
	case err.Is(sdkErrors.ErrEntityNotFound):
		c.PrintErrln("Error: Entity not found.")
		return true
	case err.Is(sdkErrors.ErrEntityInvalid):
		c.PrintErrln("Error: Invalid entity.")
		return true
	case err.Is(sdkErrors.ErrAPIPostFailed):
		c.PrintErrln("Error: Operation failed.")
		return true
	case err.Is(sdkErrors.ErrEntityCreationFailed):
		c.PrintErrln("Error: Failed to create resource.")
		return true
	}
	return false
}

// handleCipherError handles cipher-specific SDK errors.
//
// Parameters:
//   - c: Cobra command for error output
//   - err: The SDK error to check and handle
//
// Returns:
//   - bool: true if the error was a cipher-specific error and was handled,
//     false otherwise
func handleCipherError(c *cobra.Command, err *sdkErrors.SDKError) bool {
	switch {
	case err.Is(sdkErrors.ErrCryptoEncryptionFailed):
		c.PrintErrln("Error: Encryption operation failed.")
		return true
	case err.Is(sdkErrors.ErrCryptoDecryptionFailed):
		c.PrintErrln("Error: Decryption operation failed.")
		return true
	case err.Is(sdkErrors.ErrCryptoCipherNotAvailable):
		c.PrintErrln("Error: Cipher not available.")
		return true
	case err.Is(sdkErrors.ErrCryptoInvalidEncryptionKeyLength):
		c.PrintErrln("Error: Invalid encryption key length.")
		return true
	}
	return false
}
