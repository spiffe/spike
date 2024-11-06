//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

import (
	"time"
)

// Delete marks secret versions as deleted for a given path. If no versions are
// specified, it marks only the current version as deleted. If specific versions
// are provided, it marks each existing version in the list as deleted. The
// deletion is performed by setting the DeletedTime to the current time. If the
// path doesn't exist, the function returns without making any changes.
func (kv *KV) Delete(path string, versions []int) {
	secret, exists := kv.data[path]
	if !exists {
		return
	}

	now := time.Now()

	// If no versions specified, mark the latest version as deleted
	if len(versions) == 0 {
		if v, exists := secret.Versions[secret.Metadata.CurrentVersion]; exists {
			v.DeletedTime = &now // Mark as deleted.
			secret.Versions[secret.Metadata.CurrentVersion] = v
		}

		return
	}

	// Delete specific versions
	for _, version := range versions {
		if version == 0 {
			v, exists := secret.Versions[secret.Metadata.CurrentVersion]
			if !exists {
				continue
			}

			v.DeletedTime = &now // Mark as deleted.
			secret.Versions[secret.Metadata.CurrentVersion] = v
			continue
		}

		if v, exists := secret.Versions[version]; exists {
			v.DeletedTime = &now // Mark as deleted.
			secret.Versions[version] = v
		}
	}
}
