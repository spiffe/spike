//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"net/http"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/net"
)

// getCipherOrFailStreaming retrieves the system cipher from the backend
// and handles errors for streaming mode requests.
//
// If the cipher is unavailable, sends a plain HTTP error response.
//
// Parameters:
//   - w: The HTTP response writer for sending error responses
//
// Returns:
//   - cipher.AEAD: The system cipher if available, nil otherwise
//   - error: An error if the cipher is unavailable, nil otherwise
func getCipherOrFailStreaming(
	w http.ResponseWriter,
) (cipher.AEAD, error) {
	c := persist.Backend().GetCipher()

	if c == nil {
		http.Error(
			w, string(sdkErrors.ErrCodeCryptoCipherNotAvailable),
			http.StatusInternalServerError,
		)
		return nil, sdkErrors.ErrCryptoCipherNotAvailable
	}

	return c, nil
}

// getCipherOrFailJSON retrieves the system cipher from the backend and
// handles errors for JSON mode requests.
//
// If the cipher is unavailable, sends a structured JSON error response.
//
// Parameters:
//   - w: The HTTP response writer for sending error responses
//   - errorResponse: The error response to send in JSON mode
//
// Returns:
//   - cipher.AEAD: The system cipher if available, nil otherwise
//   - error: An error if the cipher is unavailable, nil otherwise
func getCipherOrFailJSON[T any](
	w http.ResponseWriter, errorResponse T,
) (cipher.AEAD, error) {
	const fName = "getCipherOrFailJSON"

	c := persist.Backend().GetCipher()
	if c == nil {
		return nil, net.Fail(
			errorResponse, w,
			http.StatusInternalServerError,
			sdkErrors.ErrCryptoCipherNotAvailable,
			fName,
		)
	}

	return c, nil
}
