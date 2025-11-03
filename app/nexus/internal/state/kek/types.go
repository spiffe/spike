//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
)

const (
	// KekSaltSize is the size of the KEK salt in bytes (32 bytes for security)
	KekSaltSize = 32

	// DefaultKekRotationDays is the default KEK rotation period in days
	DefaultKekRotationDays = 90

	// DefaultKekMaxWraps is the maximum number of wraps before KEK rotation
	DefaultKekMaxWraps = 20_000_000

	// DefaultKekGraceDays is the grace period for old KEKs in days
	DefaultKekGraceDays = 180

	// DomainSeparationInfo is the info string for HKDF domain separation
	DomainSeparationInfo = "spike:kek:v1"
)

// Metadata represents the metadata for a Key Encryption Key
type Metadata struct {
	// ID is the unique identifier for this KEK (e.g., "v2025-01")
	ID string `json:"kek_id"`

	// Version is the numeric version of this KEK
	Version int `json:"version"`

	// Salt is the random salt used in HKDF derivation (stored as non-secret)
	Salt [KekSaltSize]byte `json:"salt"`

	// RMKVersion is the version of the RMK used to derive this KEK
	RMKVersion int `json:"rmk_version"`

	// CreatedAt is when this KEK was created
	CreatedAt time.Time `json:"created_at"`

	// WrapsCount tracks how many DEK wraps have been performed with this KEK
	WrapsCount int64 `json:"wraps_count"`

	// Status indicates the current state of this KEK
	Status KekStatus `json:"status"`

	// RetiredAt is when this KEK was retired (nil if active)
	RetiredAt *time.Time `json:"retired_at,omitempty"`
}

// KekStatus represents the lifecycle status of a KEK
type KekStatus string

const (
	// KekStatusActive means the KEK is currently in use for new wraps
	KekStatusActive KekStatus = "active"

	// KekStatusGrace means the KEK is in grace period (readable but not for new wraps)
	KekStatusGrace KekStatus = "grace"

	// KekStatusRetired means the KEK should no longer be used
	KekStatusRetired KekStatus = "retired"
)

// SecretMetadata represents the encryption metadata stored with each secret
type SecretMetadata struct {
	// KekID identifies which KEK was used to wrap this secret's DEK
	KekID string `json:"kek_id"`

	// WrappedDEK is the DEK encrypted by the KEK
	WrappedDEK []byte `json:"wrapped_dek"`

	// AEADAlg is the AEAD algorithm used (e.g., "AES-256-GCM")
	AEADAlg string `json:"aead_alg"`

	// Nonce for the DEK encryption
	Nonce []byte `json:"nonce"`

	// Tag for authentication
	Tag []byte `json:"tag"`

	// AAD is additional authenticated data
	AAD []byte `json:"aad,omitempty"`

	// RewrappedAt tracks when this DEK was last rewrapped
	RewrappedAt *time.Time `json:"rewrapped_at,omitempty"`
}

// RotationPolicy defines the KEK rotation policy
type RotationPolicy struct {
	// RotationDays is the number of days before automatic rotation
	RotationDays int

	// MaxWraps is the maximum number of wraps before rotation
	MaxWraps int64

	// GraceDays is the grace period for old KEKs
	GraceDays int

	// LazyRewrapEnabled enables lazy rewrapping on read
	LazyRewrapEnabled bool

	// MaxRewrapQPS limits the rewrap rate
	MaxRewrapQPS int
}

// DefaultRotationPolicy returns the default KEK rotation policy
func DefaultRotationPolicy() *RotationPolicy {
	return &RotationPolicy{
		RotationDays:      DefaultKekRotationDays,
		MaxWraps:          DefaultKekMaxWraps,
		GraceDays:         DefaultKekGraceDays,
		LazyRewrapEnabled: true,
		MaxRewrapQPS:      100,
	}
}

// Cache represents an in-memory cache of derived KEKs
type Cache struct {
	// keys maps KEK ID to the derived key material
	keys map[string]*[crypto.AES256KeySize]byte
}
