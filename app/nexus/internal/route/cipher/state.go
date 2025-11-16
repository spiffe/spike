//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"fmt"
	"net/http"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"github.com/spiffe/spike/internal/net"
)

// getCipherOrFail retrieves the system cipher from the backend and handles
// errors appropriately based on the request mode.
//
// If the cipher is unavailable, the function sends an appropriate error
// response to the client based on whether streaming mode is active:
//   - Streaming mode: Sends plain HTTP error
//   - JSON mode: Sends structured JSON error response
//
// Parameters:
//   - w: The HTTP response writer for sending error responses
//   - streamModeActive: Whether the request is in streaming mode
//   - errorResponse: The error response to send in JSON mode
//
// Returns:
//   - cipher.AEAD: The system cipher if available, nil otherwise
//   - error: An error if the cipher is unavailable, nil otherwise
func getCipherOrFail[T any](
	w http.ResponseWriter, streamModeActive bool, errorResponse T,
) (cipher.AEAD, error) {
	c := persist.Backend().GetCipher()
	if c == nil {
		if streamModeActive {
			http.Error(
				w, "cipher not available",
				http.StatusInternalServerError,
			)
			return nil, fmt.Errorf("cipher not available")
		}

		return nil, net.Fail(
			errorResponse,
			w,
			http.StatusInternalServerError,
			fmt.Errorf("cipher not available"),
			"",
		)
	}
	return c, nil
}
