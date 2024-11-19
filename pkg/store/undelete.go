//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

// Undelete restores previously deleted versions of a secret at the specified
// path. It sets the DeletedTime to nil for each specified version that exists.
//
// Parameters:
//   - path: The location of the secret in the store
//   - versions: A slice of version numbers to undelete
//
// Returns:
//   - error: ErrSecretNotFound if the path doesn't exist, nil on success
//
// If a version number in the versions slice doesn't exist, it is silently
// skipped without returning an error. Only existing versions are modified.
func (kv *KV) Undelete(path string, versions []int) error {
	secret, exists := kv.data[path]
	if !exists {
		return ErrSecretNotFound
	}

	cv := secret.Metadata.CurrentVersion

	// If no versions specified, mark the latest version as undeleted
	if len(versions) == 0 {
		if v, exists := secret.Versions[cv]; exists {
			v.DeletedTime = nil // Mark as undeleted.
			secret.Versions[cv] = v
		}

		return nil
	}

	// Delete specific versions
	for _, version := range versions {
		if version == 0 {
			v, exists := secret.Versions[cv]
			if !exists {
				continue
			}

			v.DeletedTime = nil // Mark as undeleted.
			secret.Versions[cv] = v
			continue
		}

		if v, exists := secret.Versions[version]; exists {
			v.DeletedTime = nil // Mark as undeleted.
			secret.Versions[version] = v
		}
	}

	return nil
}
