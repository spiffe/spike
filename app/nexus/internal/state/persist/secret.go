//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"

	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike-sdk-go/retry"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
)

// ReadSecret attempts to retrieve a secret from the backend cache at the
// specified path and version. If no specific version is provided (version = 0),
// it uses the secret's current version.
//
// Parameters:
//   - path: The path where the secret is stored in the cache
//   - version: The specific version of the secret to retrieve.
//     If 0, uses the current version
//
// Returns:
//   - A pointer to the kv.Secret if found and not deleted, nil otherwise
//
// The function returns nil in several cases:
//   - When the backend is not available
//   - When there's an error loading from the cache
//   - When the requested version doesn't exist
//   - When the secret version has been deleted (DeletedTime is set)
func ReadSecret(path string, version int) *kv.Value {
	be := Backend()
	if be == nil {
		return nil
	}

	retrier := retry.NewExponentialRetrier()
	typedRetrier := retry.NewTypedRetrier[*kv.Value](retrier)

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	cachedSecret, err := typedRetrier.RetryWithBackoff(
		ctx, func() (*kv.Value, error) {
			return be.LoadSecret(ctx, path)
		})

	if err != nil {
		log.Log().Warn("readSecret",
			"msg", "Failed to load secret from cache after retries",
			"path", path,
			"err", err.Error(),
		)
		return nil
	}

	if cachedSecret != nil {
		if version == 0 {
			version = cachedSecret.Metadata.CurrentVersion
		}

		if sv, ok := cachedSecret.Versions[version]; ok &&
			sv.DeletedTime == nil {
			return cachedSecret
		}
	}

	return nil
}

// ReadAllSecrets retrieves all secrets from the backend key-value store.
// It uses an exponential backoff retry mechanism to handle transient errors.
// The function sets a timeout based on the environment's
// DatabaseOperationTimeout.
//
// If the backend is nil or if loading secrets fails after all retry attempts,
// the function returns nil. Any errors during retrieval are logged as warnings.
//
// Returns:
//   - map[string]*kv.Value: A map of all secrets with their keys and values.
//     Returns nil if the backend is unavailable or if loading fails.
func ReadAllSecrets() map[string]*kv.Value {
	be := Backend()
	if be == nil {
		return nil
	}

	retrier := retry.NewExponentialRetrier()
	typedRetrier := retry.NewTypedRetrier[map[string]*kv.Value](retrier)

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	cachedSecrets, err := typedRetrier.RetryWithBackoff(
		ctx, func() (map[string]*kv.Value, error) {
			return be.LoadAllSecrets(ctx)
		})

	if err != nil {
		log.Log().Warn("readAllSecrets",
			"msg", "Failed to load secrets from cache after retries",
			"err", err.Error(),
		)
		return nil
	}

	return cachedSecrets
}

// StoreSecret stores a secret from the key-value store kv to the
// backend cache. It retrieves the secret from the provided path and attempts to
// cache it in a background goroutine. If the backend is not available or if the
// kv operation fails, it will only log a warning since the key-value store kv
// remains the source of truth.
//
// Parameters:
//   - kv: A pointer to the key-value store kv containing the secrets
//   - path: The path where the secret is stored in the key-value store kv
//
// The function does not return any errors since it handles them internally
// through logging. Cache failures are non-fatal as the key-value store kv is
// considered the authoritative data source.
func StoreSecret(kv *kv.KV, path string) {
	const fName = "storeSecret"
	be := Backend()

	if be == nil {
		return
	}

	// Get the full secret for caching
	secret, err := kv.GetRawSecret(path)
	if err != nil {
		log.Log().Info(fName, "msg", err.Error())
	}

	if secret == nil {
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		env.DatabaseOperationTimeout(),
	)
	defer cancel()

	if err := be.StoreSecret(ctx, path, *secret); err != nil {
		// Log error but continue - memory is the source of truth
		log.Log().Warn(fName,
			"msg", "Failed to cache secret",
			"path", path,
			"err", err.Error(),
		)
	}
}
