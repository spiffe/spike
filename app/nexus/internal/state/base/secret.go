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
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// UpsertSecret stores or updates a secret at the specified path with the
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
//   - path: The namespace path where the secret should be stored
//   - values: A map containing the secret key-value pairs to be stored
//
// Returns:
//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil on
//     success
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
func UpsertSecret(path string, values map[string]string) *sdkErrors.SDKError {
	ctx := context.Background()

	// Load the current secret (if it exists) to handle versioning.
	// ErrEntityNotFound means the secret doesn't exist yet, which is fine for
	// upsert semantics. Any other error indicates a backend problem.
	currentSecret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		if !err.Is(sdkErrors.ErrEntityNotFound) {
			failErr := sdkErrors.ErrEntityLoadFailed.Wrap(err)
			failErr.Msg = "failed to load secret with path " + path
			return failErr
		}
		// Secret doesn't exist: currentSecret remains nil, and we'll create it
		currentSecret = nil
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
		var newVersion int
		if currentSecret.Metadata.CurrentVersion == 0 {
			// All versions are deleted - find the highest existing version
			// and increment to avoid collision with deleted versions
			maxVersion := 0
			for v := range currentSecret.Versions {
				if v > maxVersion {
					maxVersion = v
				}
			}
			newVersion = maxVersion + 1
		} else {
			newVersion = currentSecret.Metadata.CurrentVersion + 1
		}

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
			cmm := currentSecret.Metadata.MaxVersions
			versionsToDelete := len(currentSecret.Versions) - cmm
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
		return err
	}

	return nil
}

// DeleteSecret soft-deletes one or more versions of a secret at the specified
// path by marking them with DeletedTime. Deleted versions remain in storage
// and can be restored using UndeleteSecret.
//
// If the current version is deleted, CurrentVersion is updated to the highest
// remaining non-deleted version, or set to 0 if all versions are deleted.
// OldestVersion is also updated to reflect the oldest non-deleted version.
//
// Parameters:
//   - path: The namespace path of the secret to delete
//   - versions: A slice of version numbers to delete. If empty, deletes the
//     current version only. Version number 0 represents the current version.
//
// Returns:
//   - *sdkErrors.SDKError: An error if the operation fails. Returns nil on
//     success.
func DeleteSecret(path string, versions []int) *sdkErrors.SDKError {
	secret, err := loadAndValidateSecret(path)
	if err != nil {
		return err
	}

	ctx := context.Background()

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
			if v.DeletedTime == nil && version > newCurrent {
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
		return err
	}

	return nil
}

// UndeleteSecret restores previously deleted versions of a secret at the
// specified path by clearing their DeletedTime. If no versions are specified,
// it undeletes the current version (or the highest deleted version if all are
// deleted).
//
// If a version higher than CurrentVersion is undeleted, CurrentVersion is
// updated to that version. OldestVersion is also updated to reflect the oldest
// non-deleted version after the undelete operation.
//
// Versions that don't exist or are already undeleted are silently skipped.
//
// Parameters:
//   - path: The namespace path of the secret to restore
//   - versions: A slice of version numbers to restore. If empty, restores the
//     current version (or latest deleted if CurrentVersion is 0). Version
//     number 0 represents the current version.
//
// Returns:
//   - *sdkErrors.SDKError: An error if no versions were undeleted or if the
//     operation fails. Returns nil on success.
//
// Example:
//
//	// Restore versions 1 and 3 of a secret
//	err := UndeleteSecret("app/secrets/api-key", []int{1, 3})
func UndeleteSecret(path string, versions []int) *sdkErrors.SDKError {
	secret, err := loadAndValidateSecret(path)
	if err != nil {
		return err
	}

	ctx := context.Background()

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
				failErr := *sdkErrors.ErrEntityNotFound.Clone()
				failErr.Msg = fmt.Sprintf(
					"could not find any secret to undelete at path %s for versions %v",
					path, versions,
				)
				return &failErr
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
		failErr := *sdkErrors.ErrEntityNotFound.Clone()
		failErr.Msg = fmt.Sprintf(
			"could not find any secret to undelete at path %s for versions %v",
			path, versions,
		)
		return &failErr
	}

	// Update CurrentVersion if we undeleted a higher version than current
	if highestUndeleted > secret.Metadata.CurrentVersion {
		secret.Metadata.CurrentVersion = highestUndeleted
		secret.Metadata.UpdatedTime = time.Now()
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
		return err
	}

	return nil
}

// GetSecret retrieves the data for a specific version of a secret at the
// specified path. Deleted versions return an error.
//
// Parameters:
//   - path: The namespace path of the secret to retrieve
//   - version: The specific version to fetch. Version 0 represents the current
//     version. Returns an error if CurrentVersion is 0 (all deleted).
//
// Returns:
//   - map[string]string: The secret key-value pairs for the requested version
//   - *sdkErrors.SDKError: An error if the secret/version is not found, is
//     deleted, or is empty. Returns nil on success.
func GetSecret(
	path string, version int,
) (map[string]string, *sdkErrors.SDKError) {
	secret, err := loadAndValidateSecret(path)
	if err != nil {
		return nil, err
	}

	// Handle version 0 (current version)
	if version == 0 {
		version = secret.Metadata.CurrentVersion
		if version == 0 {
			failErr := *sdkErrors.ErrEntityNotFound.Clone()
			failErr.Msg = fmt.Sprintf("secret with path %s is empty", path)
			return nil, &failErr
		}
	}

	// Get the specific version
	v, exists := secret.Versions[version]
	if !exists {
		failErr := *sdkErrors.ErrEntityNotFound.Clone()
		failErr.Msg = fmt.Sprintf(
			"secret with path %s not found for version %v",
			path, version,
		)
		return nil, &failErr
	}

	// Check if the version is deleted
	if v.DeletedTime != nil {
		failErr := *sdkErrors.ErrEntityNotFound.Clone()
		failErr.Msg = fmt.Sprintf(
			"secret with path %s is marked deleted for version %v",
			path, version,
		)
		return nil, &failErr
	}

	return v.Data, nil
}

// GetRawSecret retrieves the complete secret structure including all versions
// and metadata from the specified path. The requested version must exist and
// be non-deleted, but the entire secret structure is returned.
//
// Parameters:
//   - path: The namespace path of the secret to retrieve
//   - version: The version to validate. Version 0 represents the current
//     version. Returns an error if CurrentVersion is 0 (all deleted).
//
// Returns:
//   - *kv.Value: The complete secret structure with all versions and metadata
//   - *sdkErrors.SDKError: An error if the secret is not found, the requested
//     version doesn't exist or is deleted, or the secret is empty. Returns
//     nil on success.
func GetRawSecret(path string, version int) (*kv.Value, *sdkErrors.SDKError) {
	secret, err := loadAndValidateSecret(path)
	if err != nil {
		return nil, err
	}

	// Validate the requested version exists and is not deleted
	checkVersion := version
	if wantsCurrentVersion := checkVersion == 0; wantsCurrentVersion {
		// Explicitly switch to the current version if the version is 0
		checkVersion = secret.Metadata.CurrentVersion
		if emptySecret := checkVersion == 0; emptySecret {
			failErr := *sdkErrors.ErrEntityNotFound.Clone()
			failErr.Msg = fmt.Sprintf("secret with path %s is empty", path)
			return nil, &failErr
		}
	}

	v, exists := secret.Versions[checkVersion]
	if !exists {
		failErr := *sdkErrors.ErrEntityNotFound.Clone()
		failErr.Msg = fmt.Sprintf(
			"secret with path %s not found for version %v",
			path, checkVersion,
		)
		return nil, &failErr
	}

	if v.DeletedTime != nil {
		failErr := *sdkErrors.ErrEntityNotFound.Clone()
		failErr.Msg = fmt.Sprintf(
			"secret with path %s is marked deleted for version %v",
			path, checkVersion,
		)
		return nil, &failErr
	}

	// Return the full secret, since we've validated the requested
	// version exists and is not deleted
	return secret, nil
}
