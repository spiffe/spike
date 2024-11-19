//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike/app/nexus/internal/state/entity/data"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// SetAdminCredentials updates the admin credentials with the provided password
// hash and salt. This function is thread-safe and persists the credentials
// asynchronously.
//
// The function performs two operations:
//  1. Updates the in-memory credentials protected by a mutex
//  2. Triggers an asynchronous persistence operation to store the credentials
//
// Parameters:
//   - passwordHash: The hashed password string for the admin
//   - salt: The salt string used in password hashing
//
// The function uses a mutex to ensure thread-safe updates to the shared
// adminCredentials variable. After updating the in-memory credentials, it
// initiates an asynchronous operation to persist the credentials to storage.
func SetAdminCredentials(passwordHash, salt string) {
	adminCredentialsMu.Lock()
	adminCredentials = data.Credentials{
		PasswordHash: passwordHash,
		Salt:         salt,
	}
	adminCredentialsMu.Unlock()

	persist.AsyncPersistAdminCredentials(data.Credentials{
		PasswordHash: passwordHash,
		Salt:         salt,
	})
}

// AdminCredentials retrieves the current admin credentials in a thread-safe
// manner. If the in-memory credentials are empty, it attempts to load them from
// persistent storage.
//
// The function follows this logic:
//  1. Attempts to read in-memory credentials with read lock
//  2. If credentials are empty (salt or hash is ""), tries to load from
//     persistent storage
//  3. If loaded from storage, updates in-memory credentials for future use
//
// Returns:
//   - data.Credentials: Contains the password hash and salt. May be empty if no
//     credentials exist in memory or persistent storage
//
// This function is thread-safe using read/write mutex protection for accessing
// shared credential data. It implements a lazy-loading pattern, only reading
// from persistent storage when necessary.
func AdminCredentials() data.Credentials {
	adminCredentialsMu.RLock()
	creds := adminCredentials
	adminCredentialsMu.RUnlock()

	salt := creds.Salt
	hash := creds.PasswordHash

	if salt == "" || hash == "" {
		cachedCreds := persist.ReadAdminCredentials()
		if cachedCreds == nil {
			return creds
		}

		adminCredentialsMu.Lock()
		adminCredentials = *cachedCreds
		adminCredentialsMu.Unlock()

		return *cachedCreds
	}

	return creds
}
