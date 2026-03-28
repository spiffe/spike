//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"fmt"

	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

const (
	nonceFieldSecretMetadataCurrentVersion string = "secret_metadata.current_version"
	nonceFieldSecretMetadataOldestVersion  string = "secret_metadata.oldest_version"
	nonceFieldSecretMetadataCreatedTime    string = "secret_metadata.created_time"
	nonceFieldSecretMetadataUpdatedTime    string = "secret_metadata.updated_time"
	nonceFieldSecretMetadataMaxVersions    string = "secret_metadata.max_versions"
)

// fieldNonceSalts are fixed per-field salts (must match AES-GCM nonce size).
// A per-field nonce is derived as baseNonce XOR salt.
var fieldNonceSalts = map[string][]byte{
	nonceFieldSecretMetadataCurrentVersion: []byte("current_ver_"), // 12 bytes
	nonceFieldSecretMetadataOldestVersion:  []byte("oldest_ver__"), // 12 bytes
	nonceFieldSecretMetadataCreatedTime:    []byte("created_tim_"), // 12 bytes
	nonceFieldSecretMetadataUpdatedTime:    []byte("updated_tim_"), // 12 bytes
	nonceFieldSecretMetadataMaxVersions:    []byte("max_versions"), // 12 bytes
}

// deriveFieldNonce derives a per-field AES-GCM nonce from a base nonce by
// XOR'ing it with a fixed, field-specific salt. This enables using a single
// per-row base nonce while still ensuring each encrypted field uses a distinct
// nonce.
//
// Parameters:
//   - baseNonce: The base nonce to derive from. Its length must match the
//     cipher's required nonce size and the salt length for the given field.
//   - field: The field identifier used to select the derivation salt.
//
// Returns:
//   - []byte: The derived nonce for the given field.
//   - *sdkErrors.SDKError: An error if the field is unknown
//     (ErrEntityInvalid) or if the nonce size does not match
//     (ErrCryptoNonceSizeMismatch). Returns nil on success.
func deriveFieldNonce(baseNonce []byte, field string) ([]byte, *sdkErrors.SDKError) {
	salt, ok := fieldNonceSalts[field]
	if !ok {
		failErr := *sdkErrors.ErrEntityInvalid.Clone()
		failErr.Msg = fmt.Sprintf("unknown nonce derivation field: %q", field)
		return nil, &failErr
	}

	if len(baseNonce) != len(salt) {
		failErr := *sdkErrors.ErrCryptoNonceSizeMismatch.Clone()
		failErr.Msg = fmt.Sprintf(
			"invalid nonce size for field %q: got %d, want %d",
			field, len(baseNonce), len(salt),
		)
		return nil, &failErr
	}

	derived := make([]byte, len(baseNonce))
	for i := range baseNonce {
		derived[i] = baseNonce[i] ^ salt[i]
	}
	return derived, nil
}

// encryptWithDerivedNonce encrypts data using a nonce derived from the
// provided base nonce and field identifier. The derived nonce is produced by
// applying the field-specific XOR salt in deriveFieldNonce.
//
// Parameters:
//   - s: The DataStore containing the AES-GCM cipher for encryption.
//   - baseNonce: The base nonce used to derive the per-field nonce.
//   - field: The field identifier used to select the derivation salt.
//   - data: The plaintext data to encrypt.
//
// Returns:
//   - []byte: The encrypted ciphertext.
//   - *sdkErrors.SDKError: An error if nonce derivation fails or if encryption
//     fails. Returns nil on success.
func encryptWithDerivedNonce(
	s *DataStore, baseNonce []byte, field string, data []byte,
) ([]byte, *sdkErrors.SDKError) {
	derivedNonce, deriveErr := deriveFieldNonce(baseNonce, field)
	if deriveErr != nil {
		return nil, deriveErr
	}

	return encryptWithNonce(s, derivedNonce, data)
}

// decryptWithDerivedNonce decrypts ciphertext using a nonce derived from the
// provided base nonce and field identifier. The derived nonce is produced by
// applying the field-specific XOR salt in deriveFieldNonce.
//
// Parameters:
//   - s: The DataStore containing the AES-GCM cipher for decryption.
//   - baseNonce: The base nonce used to derive the per-field nonce.
//   - field: The field identifier used to select the derivation salt.
//   - ciphertext: The encrypted data to decrypt.
//
// Returns:
//   - []byte: The decrypted plaintext.
//   - *sdkErrors.SDKError: An error if nonce derivation fails or if decryption
//     fails. Returns nil on success.
func decryptWithDerivedNonce(
	s *DataStore, baseNonce []byte, field string, ciphertext []byte,
) ([]byte, *sdkErrors.SDKError) {
	derivedNonce, deriveErr := deriveFieldNonce(baseNonce, field)
	if deriveErr != nil {
		return nil, deriveErr
	}

	return s.decrypt(ciphertext, derivedNonce)
}
