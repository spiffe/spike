//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/kv"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// UpsertSecret stores or updates a secret at the specified pathPattern with the
// provided values. It handles version management, maintaining a history of
// secret values up to the configured maximum number of versions.
//
// For new secrets, it creates the initial version (version 1). For existing
// secrets, it increments the version number and adds the new values while
// preserving history. Old versions are automatically pruned when the total
// number of versions exceeds the configured maximum.
//
// All operations are performed directly against the backing store without
// caching, ensuring consistency across multiple instances in high-availability
// deployments.
//
// Parameters:
//   - pathPattern: The location where the secret should be stored
//   - values: A map containing the secret key-value pairs to be stored
//
// Returns:
//   - error: An error if the operation fails, nil on success
//
// Example:
//
//	err := UpsertSecret("app/database/credential", map[string]string{
//	    "username": "admin",
//	    "password": "SPIKE_Rocks",
//	})
//	if err != nil {
//	    log.Printf("Failed to store secret: %v", err)
//	}
func UpsertSecret(path string, values map[string]string) error {
	ctx := context.Background()

	// Load the current secret (if it exists) to handle versioning
	// Backend does NOT return an error if the secret is not found and returns
	// `nil` instead. Any other error means there is a problem with the
	// backing store, so it's better to return it and exit the function.
	currentSecret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to load current secret: %w", err)
	}

	now := time.Now()

	// Build the secret structure
	var secret *kv.Value
	if currentSecret == nil {
		// New secret - create from scratch
		secret = &kv.Value{
			Versions: map[int]kv.Version{
				1: {
					Data:        values,
					CreatedTime: now,
					Version:     1,
					DeletedTime: nil,
				},
			},
			Metadata: kv.Metadata{
				CreatedTime:    now,
				UpdatedTime:    now,
				CurrentVersion: 1,
				OldestVersion:  1,
				MaxVersions:    env.MaxSecretVersionsVal(), // Get from the environment
			},
		}
	} else {
		// Existing secret - increment version
		newVersion := currentSecret.Metadata.CurrentVersion + 1

		// Add the new version
		currentSecret.Versions[newVersion] = kv.Version{
			Data:        values,
			CreatedTime: now,
			Version:     newVersion,
			DeletedTime: nil,
		}

		// Update metadata
		currentSecret.Metadata.CurrentVersion = newVersion
		currentSecret.Metadata.UpdatedTime = now

		// Clean up old versions if exceeding MaxVersions
		if len(currentSecret.Versions) > currentSecret.Metadata.MaxVersions {
			// Find and remove oldest versions
			versionsToDelete := len(currentSecret.Versions) - currentSecret.Metadata.MaxVersions
			sortedVersions := make([]int, 0, len(currentSecret.Versions))
			for v := range currentSecret.Versions {
				sortedVersions = append(sortedVersions, v)
			}
			sort.Ints(sortedVersions)

			for i := 0; i < versionsToDelete; i++ {
				delete(currentSecret.Versions, sortedVersions[i])
			}

			// Update OldestVersion
			if len(sortedVersions) > versionsToDelete {
				currentSecret.Metadata.OldestVersion = sortedVersions[versionsToDelete]
			}
		}

		secret = currentSecret
	}

	// Store to the backend
	err = persist.Backend().StoreSecret(ctx, path, *secret)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}

	return nil
}

// DeleteSecret deletes one or more versions of a secret at the specified pathPattern.
// It acquires a mutex lock before performing the deletion to ensure thread
// safety.
//
// Parameters:
//   - pathPattern: The pathPattern to the secret to be deleted
//   - versions: A slice of version numbers to delete. If empty, deletes the
//     current version only. Version number 0 is the current version.
func DeleteSecret(path string, versions []int) error {
	ctx := context.Background()

	// Load the current secret from the backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to load secret: %w", err)
	}
	if secret == nil {
		return fmt.Errorf("secret not found at pathPattern: %s", path)
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
			if v.DeletedTime == nil &&
				version > newCurrent && version < secret.Metadata.CurrentVersion {
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
// specified pathPattern. It takes a pathPattern string identifying the secret's location and
// a slice of version numbers to restore. The function acquires a lock on the
// key-value kv to ensure thread-safe operations during the `undelete` process.
//
// The function operates synchronously and will block until the undelete
// operation is complete. If any specified version numbers don't exist or were
// not previously deleted, those versions will be silently skipped.
//
// Parameters:
//   - pathPattern: The pathPattern to the secret to be restored
//   - versions: A slice of integer version numbers to restore
//
// Example:
//
//	// Restore versions 1 and 3 of a secret
//	UndeleteSecret("app/secrets/api-key", []int{1, 3})
func UndeleteSecret(path string, versions []int) error {
	ctx := context.Background()

	// Load the current secret from the backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to load secret: %w", err)
	}
	if secret == nil {
		return fmt.Errorf("secret not found at pathPattern: %s", path)
	}

	currentVersion := secret.Metadata.CurrentVersion

	// If no versions specified,
	// undelete the current version (or latest if current is 0)
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
				return fmt.Errorf("no deleted versions to undelete at pathPattern: %s", path)
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
		return fmt.Errorf("no versions were undeleted at pathPattern: %s", path)
	}

	// If CurrentVersion was 0 (all deleted), set it to
	// the highest undeleted version
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

// GetSecret retrieves a secret from the specified pathPattern and version.
// It provides thread-safe read access to the secret kv.
//
// Parameters:
//   - pathPattern: The location of the secret to retrieve
//   - version: The specific version of the secret to fetch
//
// Returns:
//   - map[string]string: The secret key-value pairs
//   - bool: Whether the secret was found
func GetSecret(path string, version int) (map[string]string, error) {
	ctx := context.Background()

	// Load secret from the backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at pathPattern: %s", path)
	}

	// Handle version 0 (current version)
	if version == 0 {
		version = secret.Metadata.CurrentVersion
		if version == 0 {
			return nil, fmt.Errorf("no active versions for secret at pathPattern: %s", path)
		}
	}

	// Get the specific version
	v, exists := secret.Versions[version]
	if !exists {
		return nil, fmt.Errorf(
			"version %d not found for secret at pathPattern: %s", version, path)
	}

	// Check if the version is deleted
	if v.DeletedTime != nil {
		return nil, fmt.Errorf(
			"version %d is deleted for secret at pathPattern: %s", version, path)
	}

	return v.Data, nil
}

// GetRawSecret retrieves a secret with metadata from the specified pathPattern and
// version. It provides thread-safe read access to the backing store.
//
// Parameters:
//   - pathPattern: The location of the secret to retrieve
//   - version: The specific version of the secret to fetch
//
// Returns:
//   - *kv.Secret: The secret type
//   - bool: Whether the secret was found
func GetRawSecret(path string, version int) (*kv.Value, error) {
	ctx := context.Background()

	// Load secret from the backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at pathPattern: %s", path)
	}

	// Validate the requested version exists and is not deleted
	checkVersion := version
	if checkVersion == 0 {
		checkVersion = secret.Metadata.CurrentVersion
		if checkVersion == 0 {
			return nil, fmt.Errorf("no active versions for secret at pathPattern: %s", path)
		}
	}

	v, exists := secret.Versions[checkVersion]
	if !exists {
		return nil, fmt.Errorf("version %d not found for secret at pathPattern: %s", checkVersion, path)
	}

	if v.DeletedTime != nil {
		return nil, fmt.Errorf("version %d is deleted for secret at pathPattern: %s", checkVersion, path)
	}

	// Return the full secret, but we've validated the requested version exists
	return secret, nil
}
