//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"

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
