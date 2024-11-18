//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import "github.com/spiffe/spike/app/nexus/internal/state/persist"

// AdminToken returns the current admin token in a thread-safe manner.
// The returned token is protected by a read lock to ensure concurrent
// access safety.
func AdminToken() string {
	adminTokenMu.RLock()
	token := adminToken
	adminTokenMu.RUnlock()

	// If token isn't in memory, try loading from SQLite
	if token == "" {
		cachedToken := persist.ReadAdminToken()
		if cachedToken != "" {
			adminTokenMu.Lock()
			adminToken = cachedToken
			adminTokenMu.Unlock()
			return cachedToken
		}
	}

	return adminToken
}

// SetAdminToken updates the admin token with the provided value.
// It uses a mutex to ensure thread-safe write operations.
//
// Parameters:
//   - token: The new admin token value to be set
func SetAdminToken(token string) {
	adminTokenMu.Lock()
	adminToken = token
	adminTokenMu.Unlock()

	persist.AsyncPersistAdminToken(token)
}
