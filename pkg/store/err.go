//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import "errors"

var (
	ErrVersionNotFound   = errors.New("version not found")
	ErrSecretNotFound    = errors.New("secret not found")
	ErrSecretSoftDeleted = errors.New("secret marked as deleted")
	ErrInvalidVersion    = errors.New("invalid version")
)
