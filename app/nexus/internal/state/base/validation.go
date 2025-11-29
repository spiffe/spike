//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"

	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// loadAndValidateSecret loads a secret from the backend and validates that it
// exists. This helper function encapsulates the common pattern of loading and
// validating secrets used across multiple functions.
//
// Parameters:
//   - path: The namespace path of the secret to load.
//
// Returns:
//   - *kv.Value: The loaded and decrypted secret value.
//   - *sdkErrors.SDKError: An error if loading fails or the secret does not
//     exist. Returns nil on success.
func loadAndValidateSecret(path string) (*kv.Value, *sdkErrors.SDKError) {
	ctx := context.Background()

	// Load the secret from the backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// contains checks whether a specific permission exists in the given slice of
// permissions.
//
// Parameters:
//   - permissions: The slice of permissions to search
//   - permission: The permission to search for
//
// Returns:
//   - true if the permission is found in the slice
//   - false otherwise
func contains(permissions []data.PolicyPermission,
	permission data.PolicyPermission) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// verifyPermissions checks whether the "haves" permissions satisfy all the
// required "wants" permissions.
//
// The "Super" permission acts as a wildcard that grants all permissions.
// If "Super" is present in haves, this function returns true regardless of
// the wants.
//
// Parameters:
//   - haves: The permissions that are available
//   - wants: The permissions that are required
//
// Returns:
//   - true if all required permissions are satisfied (or "super" is present)
//   - false if any required permission is missing
func verifyPermissions(
	haves []data.PolicyPermission,
	wants []data.PolicyPermission,
) bool {
	// The "Super" permission grants all permissions.
	if contains(haves, data.PermissionSuper) {
		return true
	}

	for _, want := range wants {
		if !contains(haves, want) {
			return false
		}
	}
	return true
}
