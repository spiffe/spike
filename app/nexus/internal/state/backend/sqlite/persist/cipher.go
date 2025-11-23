//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import "crypto/cipher"

// GetCipher retrieves the AEAD cipher instance used for encrypting and
// decrypting secrets stored in the database. The cipher is initialized when
// the DataStore is created and remains constant throughout its lifetime.
//
// Returns:
//   - cipher.AEAD: The authenticated encryption with associated data cipher
//     instance used for secret encryption and decryption operations.
func (s *DataStore) GetCipher() cipher.AEAD {
	return s.Cipher
}
