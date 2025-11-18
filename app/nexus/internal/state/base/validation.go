//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"errors"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
)

// loadAndValidateSecret loads a secret from the backend and validates that it
// exists. This helper function encapsulates the common pattern of loading and
// validating secrets used across multiple functions.
//
// Parameters:
//   - fName: The name of the calling function (for logging purposes)
//   - path: The path to the secret
//
// Returns:
//   - *kv.Value: The loaded secret if successful
//   - error: An error if loading fails or the secret does not exist
func loadAndValidateSecret(fName, path string) (*kv.Value, error) {
	ctx := context.Background()

	// Load the secret from the backing store
	secret, err := persist.Backend().LoadSecret(ctx, path)
	if err != nil {
		failErr := sdkErrors.ErrDataLoadFailed
		return nil, errors.Join(failErr, err)
	}
	if secret == nil {
		failMsg := sdkErrors.InvalidFor("secret", "pathPattern", path)
		log.Log().Warn(fName, "message", failMsg)
		return nil, errors.Join(sdkErrors.ErrInvalidForAReason, err)
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

// hasAllPermissions checks whether the "haves" permissions satisfy all the
// required "wants" permissions.
//
// The "Super" permission grants all permissions. If "Super" is present in the
// haves, this function returns true regardless of the wants.
//
// Parameters:
//   - haves: The permissions that are available
//   - wants: The permissions that are required
//
// Returns:
//   - true if all required permissions are satisfied
//   - false if any required permission is missing
func hasAllPermissions(
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
