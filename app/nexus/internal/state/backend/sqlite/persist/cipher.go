//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import "crypto/cipher"

// GetCipher retrieves the AEAD cipher instance used for encrypting and
// decrypting secrets stored in the database. The cipher is initialized when
// the DataStore is created and remains constant throughout its lifetime.
//
// This method includes a nil receiver guard because it may be passed as a
// method value (e.g., `backend.GetCipher` without parentheses) where the
// receiver is bound at capture time. If the backend is nil when the method
// value is later invoked, the guard prevents a panic by returning nil
// instead of dereferencing a nil pointer. Callers should check for a nil
// return value.
//
// Returns:
//   - cipher.AEAD: The authenticated encryption with associated data cipher
//     instance, or nil if the receiver is nil
func (s *DataStore) GetCipher() cipher.AEAD {
	if s == nil {
		return nil
	}
	return s.Cipher
}
