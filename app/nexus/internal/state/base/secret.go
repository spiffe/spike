//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"fmt"
	"github.com/spiffe/spike-sdk-go/kv"
	"time"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// UpsertSecret stores or updates a secret at the specified path with the
// provided values. It provides thread-safe access to the underlying key-value
// kv.
//
// Parameters:
//   - path: The location where the secret should be stored
//   - values: A map containing the secret key-value pairs to be stored
func UpsertSecret(path string, values map[string]string) {
	//secretStoreMu.Lock()
	//secretStore.Put(path, values)
	//secretStoreMu.Unlock()
	//
	//persist.StoreSecret(secretStore, path)

}

// DeleteSecret deletes one or more versions of a secret at the specified path.
// It acquires a mutex lock before performing the deletion to ensure thread
// safety.
//
// Parameters:
//   - path: The path to the secret to be deleted
//   - versions: A slice of version numbers to delete. If empty, deletes the
//     current version only. Version number 0 is the current version.
func DeleteSecret(path string, versions []int) error {
	ctx := context.Background()

	// Load the current secret from backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to load secret: %w", err)
	}
	if secret == nil {
		return fmt.Errorf("secret not found at path: %s", path)
	}

	// If no versions specified OR version 0 specified, delete the current version
	if len(versions) == 0 {
		versions = []int{secret.Metadata.CurrentVersion}
	} else {
		// Replace any 0s with the current version
		for i, v := range versions {
			if v == 0 {
				versions[i] = secret.Metadata.CurrentVersion
			}
		}
	}

	// Mark specified versions as deleted
	now := time.Now()
	deletingCurrent := false
	for _, version := range versions {
		if v, exists := secret.Versions[version]; exists {
			v.DeletedTime = &now
			secret.Versions[version] = v

			if version == secret.Metadata.CurrentVersion {
				deletingCurrent = true
			}
		}
	}

	// If we deleted the current version, find the highest non-deleted version
	if deletingCurrent {
		newCurrent := 0 // Start at 0 (meaning "no valid version")
		for version, v := range secret.Versions {
			if v.DeletedTime == nil && version > newCurrent && version < secret.Metadata.CurrentVersion {
				newCurrent = version
			}
		}

		secret.Metadata.CurrentVersion = newCurrent
		secret.Metadata.UpdatedTime = now
	}

	// Update OldestVersion to track the oldest non-deleted version
	oldestVersion := 0
	for version, v := range secret.Versions {
		if v.DeletedTime == nil {
			if oldestVersion == 0 || version < oldestVersion {
				oldestVersion = version
			}
		}
	}
	secret.Metadata.OldestVersion = oldestVersion

	// Store the updated secret back to the backend
	err = persist.Backend().StoreSecret(ctx, path, *secret)
	if err != nil {
		return fmt.Errorf("failed to store updated secret: %w", err)
	}

	return nil
}

// UndeleteSecret restores previously deleted versions of a secret at the
// specified path. It takes a path string identifying the secret's location and
// a slice of version numbers to restore. The function acquires a lock on the
// key-value kv to ensure thread-safe operations during the `undelete` process.
//
// The function operates synchronously and will block until the undelete
// operation is complete. If any specified version numbers don't exist or were
// not previously deleted, those versions will be silently skipped.
//
// Parameters:
//   - path: The path to the secret to be restored
//   - versions: A slice of integer version numbers to restore
//
// Example:
//
//	// Restore versions 1 and 3 of a secret
//	UndeleteSecret("/app/secrets/api-key", []int{1, 3})
func UndeleteSecret(path string, versions []int) error {
	ctx := context.Background()

	// Load the current secret from backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to load secret: %w", err)
	}
	if secret == nil {
		return fmt.Errorf("secret not found at path: %s", path)
	}

	currentVersion := secret.Metadata.CurrentVersion

	// If no versions specified, undelete the current version (or latest if current is 0)
	if len(versions) == 0 {
		// If CurrentVersion is 0 (all deleted), find the highest deleted version
		if currentVersion == 0 {
			highestDeleted := 0
			for version, v := range secret.Versions {
				if v.DeletedTime != nil && version > highestDeleted {
					highestDeleted = version
				}
			}
			if highestDeleted > 0 {
				versions = []int{highestDeleted}
			} else {
				return fmt.Errorf("no deleted versions to undelete at path: %s", path)
			}
		} else {
			versions = []int{currentVersion}
		}
	}

	// Undelete specific versions
	anyUndeleted := false
	highestUndeleted := 0
	for _, version := range versions {
		// Handle version 0 (current version)
		if version == 0 {
			if currentVersion == 0 {
				continue // Can't undelete "current" when there is no current
			}
			version = currentVersion
		}

		if v, exists := secret.Versions[version]; exists {
			if v.DeletedTime != nil {
				v.DeletedTime = nil // Mark as undeleted
				secret.Versions[version] = v
				anyUndeleted = true

				if version > highestUndeleted {
					highestUndeleted = version
				}
			}
		}
	}

	if !anyUndeleted {
		return fmt.Errorf("no versions were undeleted at path: %s", path)
	}

	// If CurrentVersion was 0 (all deleted), set it to the highest undeleted version
	if secret.Metadata.CurrentVersion == 0 && highestUndeleted > 0 {
		secret.Metadata.CurrentVersion = highestUndeleted
		secret.Metadata.UpdatedTime = time.Now()
	}

	// Store the updated secret back to the backend
	err = persist.Backend().StoreSecret(ctx, path, *secret)
	if err != nil {
		return fmt.Errorf("failed to store updated secret: %w", err)
	}

	return nil
}

// GetSecret retrieves a secret from the specified path and version.
// It provides thread-safe read access to the secret kv.
//
// Parameters:
//   - path: The location of the secret to retrieve
//   - version: The specific version of the secret to fetch
//
// Returns:
//   - map[string]string: The secret key-value pairs
//   - bool: Whether the secret was found
func GetSecret(path string, version int) (map[string]string, error) {
	ctx := context.Background()

	// Load secret from backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	// Handle version 0 (current version)
	if version == 0 {
		version = secret.Metadata.CurrentVersion
		if version == 0 {
			return nil, fmt.Errorf("no active versions for secret at path: %s", path)
		}
	}

	// Get the specific version
	v, exists := secret.Versions[version]
	if !exists {
		return nil, fmt.Errorf("version %d not found for secret at path: %s", version, path)
	}

	// Check if version is deleted
	if v.DeletedTime != nil {
		return nil, fmt.Errorf("version %d is deleted for secret at path: %s", version, path)
	}

	return v.Data, nil
}

//// ImportSecrets imports a set of secrets into the application's memory state.
//// Locks the secret store mutex during the operation to ensure thread safety.
//func ImportSecrets(secrets map[string]*kv.Value) {
//	secretStoreMu.Lock()
//	defer secretStoreMu.Unlock()
//	secretStore.ImportSecrets(secrets)
//}

// GetRawSecret retrieves a secret with metadata from the specified path and
// version. It provides thread-safe read access to the secret kv.
//
// Parameters:
//   - path: The location of the secret to retrieve
//   - version: The specific version of the secret to fetch
//
// Returns:
//   - *kv.Secret: The secret type
//   - bool: Whether the secret was found
func GetRawSecret(path string, version int) (*kv.Value, error) {
	ctx := context.Background()

	// Load secret from backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path: %s", path)
	}

	// Validate the requested version exists and is not deleted
	checkVersion := version
	if checkVersion == 0 {
		checkVersion = secret.Metadata.CurrentVersion
		if checkVersion == 0 {
			return nil, fmt.Errorf("no active versions for secret at path: %s", path)
		}
	}

	v, exists := secret.Versions[checkVersion]
	if !exists {
		return nil, fmt.Errorf("version %d not found for secret at path: %s", checkVersion, path)
	}

	if v.DeletedTime != nil {
		return nil, fmt.Errorf("version %d is deleted for secret at path: %s", checkVersion, path)
	}

	// Return the full secret, but we've validated the requested version exists
	return secret, nil
}
