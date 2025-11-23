//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/kv"
)

// toSecretMetadataSuccessResponse converts a key-value store secret value into
// a secret metadata response.
//
// The function transforms the internal kv.Value representation into the API
// response format by:
//   - Converting all secret versions into a map of version info
//   - Extracting metadata including current/oldest versions and timestamps
//   - Preserving version-specific details like creation and deletion times
//
// This conversion is used when clients request secret metadata without
// retrieving the actual secret data, allowing them to inspect version history
// and lifecycle information.
//
// Parameters:
//   - secret: The key-value store secret value containing version history
//     and metadata
//
// Returns:
//   - reqres.SecretMetadataResponse: The formatted metadata response containing
//     version information and metadata suitable for API responses
func toSecretMetadataSuccessResponse(
	secret *kv.Value,
) reqres.SecretMetadataResponse {
	versions := make(map[int]data.SecretVersionInfo)
	for _, version := range secret.Versions {
		versions[version.Version] = data.SecretVersionInfo{
			CreatedTime: version.CreatedTime,
			Version:     version.Version,
			DeletedTime: version.DeletedTime,
		}
	}

	return reqres.SecretMetadataResponse{
		SecretMetadata: data.SecretMetadata{
			Versions: versions,
			Metadata: data.SecretMetaDataContent{
				CurrentVersion: secret.Metadata.CurrentVersion,
				OldestVersion:  secret.Metadata.OldestVersion,
				CreatedTime:    secret.Metadata.CreatedTime,
				UpdatedTime:    secret.Metadata.UpdatedTime,
				MaxVersions:    secret.Metadata.MaxVersions,
			},
		},
	}.Success()
}
