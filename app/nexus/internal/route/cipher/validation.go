//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package cipher

import (
	"crypto/cipher"
	"fmt"
	"net/http"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike/internal/net"
)

// validateStreamingDecryptData validates the version and nonce size for
// streaming mode decryption requests.
//
// If validation fails, this function sends an appropriate HTTP error response
// and returns an error.
//
// Parameters:
//   - version: The protocol version byte to validate
//   - nonce: The nonce bytes to validate
//   - c: The cipher to determine expected nonce size
//   - w: The HTTP response writer for error responses
//
// Returns:
//   - nil if all validations pass
//   - error if the version is unsupported or nonce size is invalid
func validateStreamingDecryptData(
	version byte, nonce []byte, c cipher.AEAD, w http.ResponseWriter,
) error {
	if version != spikeCipherVersion {
		http.Error(w, "unsupported version", http.StatusBadRequest)
		return fmt.Errorf("unsupported version: %v", version)
	}

	if len(nonce) != c.NonceSize() {
		http.Error(w, "invalid nonce size", http.StatusBadRequest)
		return fmt.Errorf(
			"invalid nonce size: expected %d, got %d",
			c.NonceSize(), len(nonce),
		)
	}

	return nil
}

// validateJSONDecryptData validates the version and nonce size for JSON mode
// decryption requests.
//
// If validation fails, this function sends an appropriate JSON error response
// and returns an error.
//
// Parameters:
//   - version: The protocol version byte to validate
//   - nonce: The nonce bytes to validate
//   - c: The cipher to determine expected nonce size
//   - w: The HTTP response writer for error responses
//
// Returns:
//   - nil if all validations pass
//   - error if version is unsupported or nonce size is invalid
func validateJSONDecryptData(
	version byte, nonce []byte, c cipher.AEAD, w http.ResponseWriter,
) error {
	if version != spikeCipherVersion {
		return net.Fail(
			reqres.CipherDecryptResponse{Err: data.ErrBadInput},
			w,
			http.StatusBadRequest,
			fmt.Errorf("unsupported version: %v", version),
			"",
		)
	}

	if len(nonce) != c.NonceSize() {
		return net.Fail(
			reqres.CipherDecryptResponse{Err: data.ErrBadInput},
			w,
			http.StatusBadRequest,
			fmt.Errorf(
				"invalid nonce size: expected %d, got %d",
				c.NonceSize(), len(nonce),
			),
			"",
		)
	}

	return nil
}
