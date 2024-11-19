//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"golang.org/x/crypto/pbkdf2"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

func updateStateForInit(password string, adminTokenBytes, salt []byte) {
	iterationCount := env.Pbkdf2IterationCount()
	hashLength := env.ShaHashLength()
	passwordHash := pbkdf2.Key(
		[]byte(password), salt,
		iterationCount, hashLength, sha256.New,
	)

	state.SetAdminToken("spike." + string(adminTokenBytes))
	state.SetAdminCredentials(
		hex.EncodeToString(passwordHash),
		hex.EncodeToString(salt),
	)
}
