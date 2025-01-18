//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"encoding/hex"
	"errors"
)

const aes256KeySize = 32

// Aes256Seed generates a cryptographically secure random 256-bit key suitable
// for use with AES-256 encryption. The key is returned as a hexadecimal-encoded
// string.
//
// Returns:
//   - string: A 64-character hexadecimal string representing the 256-bit key.
//   - error: Returns nil on successful key generation, or an error if the random
//     number generation fails.
//
// The function uses a cryptographically secure random number generator to ensure
// the generated key is suitable for cryptographic use. The resulting hex string
// can be decoded back to bytes using hex.DecodeString when needed for encryption.
func Aes256Seed() (string, error) {
	// Generate a 256-bit key
	key := make([]byte, aes256KeySize)

	// TODO: maybe move this to spike-sdk-go as it is a useful function.

	_, err := reader(key)
	if err != nil {
		return "", errors.Join(
			err,
			errors.New("Aes256Seed: failed to generate random key"),
		)
	}

	return hex.EncodeToString(key), nil
}
