//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

// Backend returns the currently initialized backend storage instance.
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
// This function is safe for concurrent access. It uses an atomic pointer to
// retrieve the backend reference, ensuring that callers always get a consistent
// view of the backend even if InitializeBackend is called concurrently.
//
// Note: Once a backend reference is returned, it remains valid for the
// lifetime of that backend instance. If InitializeBackend is called again,
// new calls to Backend() will return the new instance, but existing references
// remain valid until their operations complete.
func Backend() backend.Backend {
	ptr := backendPtr.Load()
	if ptr == nil {
		return nil
	}
	return *ptr
}
