//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/kv"
)

func toSecretMetadataResponse(
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
	}
}
