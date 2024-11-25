//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/store"
)

// TODO: we'll need persistence for policies too.

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
//   - A pointer to the store.Secret if found and not deleted, nil otherwise
//
// The function returns nil in several cases:
//   - When the backend is not available
//   - When there's an error loading from the cache
//   - When the requested version doesn't exist
//   - When the secret version has been deleted (DeletedTime is set)
func ReadSecret(path string, version int) *store.Secret {
	be := Backend()
	if be == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	cachedSecret, err := be.LoadSecret(ctx, path)
	if err != nil {
		log.Log().Warn("readSecret",
			"msg", "Failed to load secret from cache",
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

// AsyncPersistSecret asynchronously stores a secret from the KV store to the
// backend cache. It retrieves the secret from the provided path and attempts to
// cache it in a background goroutine. If the backend is not available or if the
// store operation fails, it will only log a warning since the KV store remains
// the source of truth.
//
// Parameters:
//   - kv: A pointer to the KV store containing the secrets
//   - path: The path where the secret is stored in the KV store
//
// The function does not return any errors since it handles them internally
// through logging. Cache failures are non-fatal as the KV store is considered
// the authoritative data source.
func AsyncPersistSecret(kv *store.KV, path string) {
	be := Backend()

	// Get the full secret for caching
	secret, err := kv.GetRawSecret(path)

	if err != nil {
		log.Log().Info("asyncPersistSecret", "msg", err.Error())
	}

	if secret != nil {
		go func() {
			if be == nil {
				return // No cache available
			}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				env.DatabaseOperationTimeout(),
			)
			defer cancel()

			if err := be.StoreSecret(ctx, path, *secret); err != nil {
				// Log error but continue - memory is source of truth
				log.Log().Warn("asyncPersistSecret",
					"msg", "Failed to cache secret",
					"path", path,
					"err", err.Error(),
				)
			}
		}()
	}
}
