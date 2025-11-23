//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike/internal/net"
)

// handleGetSecretError processes errors that occur during secret retrieval
// operations and sends appropriate HTTP responses.
//
// The function distinguishes between two types of errors:
//   - sdkErrors.ErrEntityNotFound: Returns HTTP 404 Not Found when the
//     requested secret does not exist at the specified path or version
//   - Other errors: Returns HTTP 500 Internal Server Error for backend or
//     server-side failures during secret retrieval
//
// Parameters:
//   - err: The error that occurred during secret retrieval
//   - w: The HTTP response writer for sending error responses
//
// Returns:
//   - *sdkErrors.SDKError: The error that was sent to the client
func handleGetSecretError(
	err *sdkErrors.SDKError, w http.ResponseWriter,
) *sdkErrors.SDKError {
	if err == nil {
		return nil
	}
	if err.Is(sdkErrors.ErrEntityNotFound) {
		net.Fail(reqres.SecretGetNotFound, w, http.StatusNotFound)
		return err
	}
	// Backend or other server-side failure
	net.Fail(reqres.SecretGetInternal, w, http.StatusInternalServerError)
	return err
}

// handleGetSecretMetadataError processes errors that occur during secret
// metadata retrieval operations and sends appropriate HTTP responses.
//
// The function distinguishes between two types of errors:
//   - sdkErrors.ErrEntityNotFound: Returns HTTP 404 Not Found when the
//     requested secret metadata does not exist at the specified path or version
//   - Other errors: Returns HTTP 500 Internal Server Error for backend or
//     server-side failures during metadata retrieval
//
// Parameters:
//   - err: The error that occurred during secret metadata retrieval
//   - w: The HTTP response writer for sending error responses
//
// Returns:
//   - *sdkErrors.SDKError: The error that was sent to the client
func handleGetSecretMetadataError(
	err *sdkErrors.SDKError, w http.ResponseWriter,
) *sdkErrors.SDKError {
	if err == nil {
		return nil
	}
	if err.Is(sdkErrors.ErrEntityNotFound) {
		net.Fail(reqres.SecretMetadataNotFound, w, http.StatusNotFound)
		return err
	}
	// Backend or other server-side failure
	net.Fail(reqres.SecretMetadataInternal, w, http.StatusInternalServerError)
	return err
}
