//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"github.com/spiffe/spike/app/nexus/internal/state/entity/data"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// SetAdminRecoveryMetadata updates the admin recovery metadata with the provided
// token hash and salt. This function is thread-safe and persists the metadata
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
// adminRecoveryMetadata variable. After updating the in-memory credentials, it
// initiates an asynchronous operation to persist the credentials to storage.
func SetAdminRecoveryMetadata(recoveryTokenHash, encryptedRootKey, salt string) {

	adminRecoveryMetadataMu.Lock()
	adminRecoveryMetadata = data.RecoveryMetadata{
		RecoveryTokenHash: recoveryTokenHash,
		EncryptedRootKey:  encryptedRootKey,
		Salt:              salt,
	}
	adminRecoveryMetadataMu.Unlock()

	persist.AsyncPersistAdminRecoveryMetadata(data.RecoveryMetadata{
		RecoveryTokenHash: recoveryTokenHash,
		EncryptedRootKey:  encryptedRootKey,
		Salt:              salt,
	})
}

// AdminRecoveryMetadata retrieves the current admin recovery metadata in a thread-safe
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
//   - data.RecoveryMetadata: Contains the password hash and salt. May be empty if no
//     credentials exist in memory or persistent storage
//
// This function is thread-safe using read/write mutex protection for accessing
// shared credential data. It implements a lazy-loading pattern, only reading
// from persistent storage when necessary.
func AdminRecoveryMetadata() data.RecoveryMetadata {
	adminRecoveryMetadataMu.RLock()
	metadata := adminRecoveryMetadata
	adminRecoveryMetadataMu.RUnlock()

	salt := metadata.Salt
	hash := metadata.RecoveryTokenHash

	if salt == "" || hash == "" {
		cachedMetadata := persist.ReadAdminRecoveryMetadata()
		if cachedMetadata == nil {
			return metadata
		}

		adminRecoveryMetadataMu.Lock()
		adminRecoveryMetadata = *cachedMetadata
		adminRecoveryMetadataMu.Unlock()

		return *cachedMetadata
	}

	return metadata
}
