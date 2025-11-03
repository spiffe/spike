//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
)

// DeriveKEK derives a KEK from the RMK using HKDF-SHA256
//
// The derivation uses:
// - IKM (Input Key Material): RMK
// - Salt: per-KEK random salt (stored as non-secret metadata)
// - Info: domain separation string including KEK ID
//
// Parameters:
//   - rmk: The Root Master Key
//   - metadata: KEK metadata containing the salt and ID
//
// Returns:
//   - Derived KEK as a 32-byte key
//   - Error if derivation fails
func DeriveKEK(
	rmk *[crypto.AES256KeySize]byte,
	metadata *Metadata,
) (*[crypto.AES256KeySize]byte, error) {
	const fName = "DeriveKEK"

	if rmk == nil {
		return nil, fmt.Errorf("%s: nil RMK provided", fName)
	}

	if metadata == nil {
		return nil, fmt.Errorf("%s: nil metadata provided", fName)
	}

	// Construct domain separation info string
	// Format: "spike:kek:v1:KEK_ID"
	info := fmt.Sprintf("%s:%s", DomainSeparationInfo, metadata.ID)

	// Create HKDF reader
	hkdfReader := hkdf.New(
		sha256.New,       // Hash function
		rmk[:],           // Input Key Material (IKM)
		metadata.Salt[:], // Salt
		[]byte(info),     // Info for domain separation
	)

	// Derive KEK
	kek := new([crypto.AES256KeySize]byte)
	if _, err := io.ReadFull(hkdfReader, kek[:]); err != nil {
		log.Log().Error(fName, "message", "failed to derive KEK", "err", err.Error())
		return nil, fmt.Errorf("%s: failed to derive KEK: %w", fName, err)
	}

	log.Log().Info(fName,
		"message", "successfully derived KEK",
		"kek_id", metadata.ID,
		"kek_version", metadata.Version)

	return kek, nil
}

// GenerateKEKSalt generates a cryptographically secure random salt for KEK derivation
func GenerateKEKSalt() ([KekSaltSize]byte, error) {
	const fName = "GenerateKEKSalt"

	var salt [KekSaltSize]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		log.Log().Error(fName, "message", "failed to generate KEK salt", "err", err.Error())
		return salt, fmt.Errorf("%s: failed to generate KEK salt: %w", fName, err)
	}

	return salt, nil
}

// GenerateDEK generates a random Data Encryption Key for a secret
func GenerateDEK() (*[crypto.AES256KeySize]byte, error) {
	const fName = "GenerateDEK"

	dek := new([crypto.AES256KeySize]byte)
	if _, err := io.ReadFull(rand.Reader, dek[:]); err != nil {
		log.Log().Error(fName, "message", "failed to generate DEK", "err", err.Error())
		return nil, fmt.Errorf("%s: failed to generate DEK: %w", fName, err)
	}

	return dek, nil
}
