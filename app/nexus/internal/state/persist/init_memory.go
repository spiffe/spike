//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/memory"
)

// initializeInMemoryBackend creates and returns a new in-memory backend
// instance. It configures the backend with the system cipher and maximum
// secret versions from the environment configuration.
//
// Returns a Backend implementation that stores all data in memory without
// persistence. This backend is suitable for testing or scenarios where
// persistent storage is not required.
func initializeInMemoryBackend() backend.Backend {
	return memory.NewInMemoryStore(
		crypto.CreateCipher(), env.MaxSecretVersionsVal(),
	)
}
