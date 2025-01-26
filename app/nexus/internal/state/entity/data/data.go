//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package data provides structures and utilities for handling various types of
// metadata and session tokens involved in the authentication process.
// This includes recovery metadata, token metadata, and session tokens
// to ensure security and manage access control within the system.
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
