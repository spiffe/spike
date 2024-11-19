//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package data

import "time"

type Credentials struct {
	PasswordHash string
	Salt         string
}

type TokenMetadata struct {
	Username  string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type SessionToken struct {
	Token     string
	Signature string
	IssuedAt  time.Time
	ExpiresAt time.Time
}
