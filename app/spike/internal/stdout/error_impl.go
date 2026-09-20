//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package stdout

import (
	"errors"
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

// policyError maps policy-specific SDK errors to user-facing errors.
//
// Parameters:
//   - err: The SDK error to check
//
// Returns:
//   - error: The user-facing error if err is a policy-specific error, nil
//     otherwise
func policyError(err *sdkErrors.SDKError) error {
	switch {
	case err.Is(sdkErrors.ErrEntityNotFound):
		return errors.New("entity not found")
	case err.Is(sdkErrors.ErrEntityInvalid):
		return errors.New("invalid entity")
	case err.Is(sdkErrors.ErrAPIPostFailed):
		return errors.New("operation failed")
	case err.Is(sdkErrors.ErrEntityCreationFailed):
		return errors.New("failed to create resource")
	}
	return nil
}

// cipherError maps cipher-specific SDK errors to user-facing errors.
//
// Parameters:
//   - err: The SDK error to check
//
// Returns:
//   - error: The user-facing error if err is a cipher-specific error, nil
//     otherwise
func cipherError(err *sdkErrors.SDKError) error {
	switch {
	case err.Is(sdkErrors.ErrCryptoEncryptionFailed):
		return errors.New("encryption operation failed")
	case err.Is(sdkErrors.ErrCryptoDecryptionFailed):
		return errors.New("decryption operation failed")
	case err.Is(sdkErrors.ErrCryptoCipherNotAvailable):
		return errors.New("cipher not available")
	case err.Is(sdkErrors.ErrCryptoInvalidEncryptionKeyLength):
		return errors.New("invalid encryption key length")
	}
	return nil
}
