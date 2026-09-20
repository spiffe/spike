//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package stdout

import (
	"errors"

	"github.com/spf13/cobra"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

const commandGroupPolicy = "policy"
const commandGroupCipher = "cipher"

// APIError translates an SDK error returned by an API call into the error a
// command handler returns to Cobra. The text is the user-facing message; the
// root command prints it to stderr and the process exits non-zero. It detects
// the command group (policy, secret, cipher) from the Cobra command path and
// maps group-specific errors accordingly.
//
// Parameters:
//   - c: Cobra command used for command group detection
//   - err: The error returned from an API call
//
// Returns:
//   - error: nil when err is nil; otherwise a non-nil error whose text is the
//     message to show the user
//
// Common error types handled for all command groups:
//   - ErrStateNotReady: System not initialized (the progressive not-ready
//     notice is printed, and a short error is returned)
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
// Any other error is returned as is, so its SDK message reaches the user.
//
// Usage example:
//
//	secret, apiErr := api.GetSecretVersion(ctx, path, version)
//	if apiErr != nil {
//	    return stdout.APIError(cmd, apiErr)
//	}
func APIError(c *cobra.Command, err *sdkErrors.SDKError) error {
	if err == nil {
		return nil
	}

	// Common errors (all command groups)
	switch {
	case err.Is(sdkErrors.ErrStateNotReady):
		PrintNotReady()
		return errors.New("SPIKE Nexus is not ready")
	case err.Is(sdkErrors.ErrDataMarshalFailure):
		return errors.New("malformed request")
	case err.Is(sdkErrors.ErrDataUnmarshalFailure):
		return errors.New("failed to parse API response")
	case err.Is(sdkErrors.ErrAPINotFound):
		return errors.New("resource not found")
	case err.Is(sdkErrors.ErrAPIBadRequest):
		return errors.New("invalid request")
	case err.Is(sdkErrors.ErrDataInvalidInput):
		return errors.New("invalid input provided")
	case err.Is(sdkErrors.ErrNetPeerConnection):
		return errors.New("failed to connect to SPIKE Nexus")
	case err.Is(sdkErrors.ErrAccessUnauthorized):
		return errors.New("unauthorized access")
	case err.Is(sdkErrors.ErrNetReadingResponseBody):
		return errors.New("failed to read response body")
	}

	// Command-group-specific errors
	switch getCommandGroup(c) {
	case commandGroupPolicy:
		if groupErr := policyError(err); groupErr != nil {
			return groupErr
		}
	case commandGroupCipher:
		if groupErr := cipherError(err); groupErr != nil {
			return groupErr
		}
	}

	// Fallback for any unhandled errors: the SDK message reaches the user.
	return err
}
