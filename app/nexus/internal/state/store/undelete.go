//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

// Undelete restores previously deleted versions of a secret at the specified path.
// It sets the DeletedTime to nil for each specified version that exists.
//
// Parameters:
//   - path: The location of the secret in the store
//   - versions: A slice of version numbers to undelete
//
// Returns:
//   - error: ErrSecretNotFound if the path doesn't exist, nil on success
//
// If a version number in the versions slice doesn't exist, it is silently skipped
// without returning an error. Only existing versions are modified.
func (kv *KV) Undelete(path string, versions []int) error {
	secret, exists := kv.data[path]
	if !exists {
		return ErrSecretNotFound
	}

	for _, version := range versions {
		if v, exists := secret.Versions[version]; exists {
			v.DeletedTime = nil
			secret.Versions[version] = v
		}
	}

	return nil
}
