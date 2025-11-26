//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import (
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/kv"
)

func TestToSecretMetadataSuccessResponse_EmptyVersions(t *testing.T) {
	secret := &kv.Value{
		Versions: map[int]kv.Version{},
		Metadata: kv.Metadata{
			CurrentVersion: 0,
			OldestVersion:  0,
			CreatedTime:    time.Now(),
			UpdatedTime:    time.Now(),
			MaxVersions:    10,
		},
	}

	response := toSecretMetadataSuccessResponse(secret)

	if len(response.SecretMetadata.Versions) != 0 {
		t.Errorf("toSecretMetadataSuccessResponse() versions = %d, want 0",
			len(response.SecretMetadata.Versions))
	}

	if response.SecretMetadata.Metadata.MaxVersions != 10 {
		t.Errorf("toSecretMetadataSuccessResponse() maxVersions = %d, want 10",
			response.SecretMetadata.Metadata.MaxVersions)
	}
}

func TestToSecretMetadataSuccessResponse_SingleVersion(t *testing.T) {
	createdTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	secret := &kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Version:     1,
				CreatedTime: createdTime,
				DeletedTime: nil,
			},
		},
		Metadata: kv.Metadata{
			CurrentVersion: 1,
			OldestVersion:  1,
			CreatedTime:    createdTime,
			UpdatedTime:    updatedTime,
			MaxVersions:    5,
		},
	}

	response := toSecretMetadataSuccessResponse(secret)

	if len(response.SecretMetadata.Versions) != 1 {
		t.Errorf("toSecretMetadataSuccessResponse() versions = %d, want 1",
			len(response.SecretMetadata.Versions))
	}

	v1, ok := response.SecretMetadata.Versions[1]
	if !ok {
		t.Fatal("toSecretMetadataSuccessResponse() missing version 1")
	}

	if v1.Version != 1 {
		t.Errorf("version.Version = %d, want 1", v1.Version)
	}

	if !v1.CreatedTime.Equal(createdTime) {
		t.Errorf("version.CreatedTime = %v, want %v", v1.CreatedTime, createdTime)
	}

	if v1.DeletedTime != nil {
		t.Errorf("version.DeletedTime = %v, want nil", v1.DeletedTime)
	}

	// Check metadata
	meta := response.SecretMetadata.Metadata
	if meta.CurrentVersion != 1 {
		t.Errorf("metadata.CurrentVersion = %d, want 1", meta.CurrentVersion)
	}
	if meta.OldestVersion != 1 {
		t.Errorf("metadata.OldestVersion = %d, want 1", meta.OldestVersion)
	}
	if meta.MaxVersions != 5 {
		t.Errorf("metadata.MaxVersions = %d, want 5", meta.MaxVersions)
	}
}

func TestToSecretMetadataSuccessResponse_MultipleVersions(t *testing.T) {
	createdTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	deletedTime := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)

	secret := &kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Version:     1,
				CreatedTime: createdTime,
				DeletedTime: &deletedTime,
			},
			2: {
				Version:     2,
				CreatedTime: createdTime.Add(24 * time.Hour),
				DeletedTime: nil,
			},
			3: {
				Version:     3,
				CreatedTime: createdTime.Add(48 * time.Hour),
				DeletedTime: nil,
			},
		},
		Metadata: kv.Metadata{
			CurrentVersion: 3,
			OldestVersion:  1,
			CreatedTime:    createdTime,
			UpdatedTime:    createdTime.Add(48 * time.Hour),
			MaxVersions:    10,
		},
	}

	response := toSecretMetadataSuccessResponse(secret)

	if len(response.SecretMetadata.Versions) != 3 {
		t.Errorf("toSecretMetadataSuccessResponse() versions = %d, want 3",
			len(response.SecretMetadata.Versions))
	}

	// Check version 1 (deleted)
	v1, ok := response.SecretMetadata.Versions[1]
	if !ok {
		t.Fatal("missing version 1")
	}
	if v1.DeletedTime == nil {
		t.Error("version 1 should have DeletedTime set")
	}

	// Check version 2 (not deleted)
	v2, ok := response.SecretMetadata.Versions[2]
	if !ok {
		t.Fatal("missing version 2")
	}
	if v2.DeletedTime != nil {
		t.Error("version 2 should not have DeletedTime set")
	}

	// Check version 3 (current)
	v3, ok := response.SecretMetadata.Versions[3]
	if !ok {
		t.Fatal("missing version 3")
	}
	if v3.Version != 3 {
		t.Errorf("version 3 Version = %d, want 3", v3.Version)
	}

	// Check metadata reflects current state
	meta := response.SecretMetadata.Metadata
	if meta.CurrentVersion != 3 {
		t.Errorf("metadata.CurrentVersion = %d, want 3", meta.CurrentVersion)
	}
}

func TestToSecretMetadataSuccessResponse_ResponseIsSuccess(t *testing.T) {
	secret := &kv.Value{
		Versions: map[int]kv.Version{},
		Metadata: kv.Metadata{
			MaxVersions: 10,
		},
	}

	response := toSecretMetadataSuccessResponse(secret)

	// The response should have empty error (success)
	if response.Err != "" {
		t.Errorf("toSecretMetadataSuccessResponse() Err = %q, want empty",
			response.Err)
	}
}
