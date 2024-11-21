//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package data

import "time"

// RecoveryMetadata represents the stored "break-the-glass" recovery
// metadata for an admin user.
type RecoveryMetadata struct {
	RecoveryTokenHash string
	EncryptedRootKey  string
	Salt              string
}

// TokenMetadata contains the metadata associated with a token.
type TokenMetadata struct {
	Username  string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// SessionToken represents a short-lived session token.
type SessionToken struct {
	Token     string
	Signature string
	IssuedAt  time.Time
	ExpiresAt time.Time
}
