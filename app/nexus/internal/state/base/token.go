//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import "github.com/spiffe/spike/app/nexus/internal/state/persist"

func adminSigToken() string {
	adminSigningTokenMu.RLock()
	token := adminSigningToken
	adminSigningTokenMu.RUnlock()

	// If token isn't in memory, try loading from SQLite
	if token == "" {
		cachedToken := persist.ReadAdminSigningToken()
		if cachedToken != "" {
			adminSigningTokenMu.Lock()
			adminSigningToken = cachedToken
			adminSigningTokenMu.Unlock()
			return cachedToken
		}
	}

	return adminSigningToken
}

func Initialized() bool {
	// We don't use the admin signing token to sign anything,
	// but its existence means the system is initialized.
	// The only time this token is set is after the successful
	// completion of the `spike init` command.
	return adminSigToken() != ""
}

// SetAdminSigningToken updates the admin token with the provided value.
// It uses a mutex to ensure thread-safe write operations.
//
// Parameters:
//   - token: The new admin token value to be set
func SetAdminSigningToken(token string) {
	adminSigningTokenMu.Lock()
	adminSigningToken = token
	adminSigningTokenMu.Unlock()

	persist.AsyncPersistAdminToken(token)
}
