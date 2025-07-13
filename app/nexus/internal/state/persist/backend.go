//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

// Backend returns the currently initialized backend storage instance. The
// function is thread-safe through a read mutex lock.
//
// Returns:
//   - A backend.Backend interface pointing to the current backend instance:
//   - memoryBackend for 'memory' store type or unknown types
//   - sqliteBackend for 'sqlite' store type
//
// The return value is determined by env.BackendStoreType():
//   - env.Memory: Returns the memory backend instance
//   - env.Sqlite: Returns the SQLite backend instance
//   - default: Falls back to the memory backend instance
//
// The function is safe for concurrent access as it uses a read mutex to protect
// access to the backend instances. Unlike InitializeBackend, this function
// returns existing instances rather than creating new ones.
func Backend() backend.Backend {
	backendMu.RLock()
	defer backendMu.RUnlock()
	return be
}
