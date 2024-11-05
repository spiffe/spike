//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import "time"

// Put stores a new version of key-value pairs at the specified path in the store.
// It implements automatic versioning with a maximum of 3 versions per path.
//
// When storing values:
// - If the path doesn't exist, it creates a new secret with initial metadata
// - Each put operation creates a new version with an incremented version number
// - Old versions are automatically pruned when exceeding MaxVersions (default: 3)
// - Timestamps are updated for both creation and modification times
//
// Parameters:
//   - path: The location where the secret will be stored
//   - values: A map of key-value pairs to store at this path
//
// The function maintains metadata including:
//   - CreatedTime: When the secret was first created
//   - UpdatedTime: When the most recent version was added
//   - CurrentVersion: The latest version number
//   - OldestVersion: The oldest available version number
//   - MaxVersions: Maximum number of versions to keep (fixed at 3)
func (kv *KV) Put(path string, values map[string]string) {
	rightNow := time.Now()

	secret, exists := kv.data[path]
	if !exists {
		secret = &Secret{
			Versions: make(map[int]Version),
			Metadata: Metadata{
				CreatedTime:    rightNow,
				UpdatedTime:    rightNow,
				MaxVersions:    3,
				CurrentVersion: 0,
				OldestVersion:  0,
			},
		}
		kv.data[path] = secret
	}

	// Increment version
	newVersion := secret.Metadata.CurrentVersion + 1

	// Add new version
	secret.Versions[newVersion] = Version{
		Data:        values,
		CreatedTime: rightNow,
		Version:     newVersion,
	}

	// Update metadata
	secret.Metadata.CurrentVersion = newVersion
	secret.Metadata.UpdatedTime = rightNow
	if secret.Metadata.OldestVersion == 0 {
		secret.Metadata.OldestVersion = 1
	}

	// Cleanup old versions if exceeding MaxVersions
	for version := range secret.Versions {
		if secret.Metadata.CurrentVersion-version >= secret.Metadata.MaxVersions {
			delete(secret.Versions, version)
			if version == secret.Metadata.OldestVersion {
				secret.Metadata.OldestVersion = version + 1
			}
		}
	}
}
