//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"crypto/rand"
	"errors"
	"io"

	sdkErrors "github.com/spiffe/spike-sdk-go/api/errors"
)

// encrypt encrypts the given data using the DataStore's cipher.
// It generates a random nonce and returns the ciphertext, nonce, and any
// error that occurred during encryption.
func (s *DataStore) encrypt(data []byte) ([]byte, []byte, error) {
	nonce := make([]byte, s.Cipher.NonceSize())
	nonceErr := sdkErrors.ErrCryptoNonceGenerationFailed
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		failErr := errors.Join(nonceErr, err)
		return nil, nil, failErr
	}
	ciphertext := s.Cipher.Seal(nil, nonce, data, nil)
	return ciphertext, nonce, nil
}

// decrypt decrypts the given ciphertext using the DataStore's cipher
// and the provided nonce. It returns the plaintext and any error that
// occurred during decryption.
func (s *DataStore) decrypt(ciphertext, nonce []byte) ([]byte, error) {
	plaintext, err := s.Cipher.Open(nil, nonce, ciphertext, nil)
	decrpyErr := sdkErrors.ErrCryptoDecryptionFailed
	if err != nil {
		failErr := errors.Join(decrpyErr, err)
		return nil, failErr
	}
	return plaintext, nil
}
