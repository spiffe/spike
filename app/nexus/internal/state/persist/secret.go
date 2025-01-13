//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"

	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/internal/log"
	"github.com/spiffe/spike/pkg/retry"
	"github.com/spiffe/spike/pkg/store"
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

	retrier := retry.NewExponentialRetrier()
	typedRetrier := retry.NewTypedRetrier[*store.Secret](retrier)

	ctx, cancel := context.WithTimeout(
		context.Background(), env.DatabaseOperationTimeout(),
	)
	defer cancel()

	// TODO: RetryWithBackoff retries indefinitely; we might want to limit the total duration
	// of db retry attempts based on sane default and configurable from environment variables.

	// TODO: check all database operations (secrets, policies, metadata) and
	// ensure that they are retried with exponential backooff.

	cachedSecret, err := typedRetrier.RetryWithBackoff(ctx, func() (*store.Secret, error) {
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

			// TODO: Yes memory is the source of truth; but at least
			// attempt some exponential retries before giving up.
			if err := be.StoreSecret(ctx, path, *secret); err != nil {
				// Log error but continue - memory is the source of truth
				log.Log().Warn("asyncPersistSecret",
					"msg", "Failed to cache secret",
					"path", path,
					"err", err.Error(),
				)
			}

			// TODO: we don't have any retry for policies or for recovery info.
			// they are equally important.

			// TODO: these async operations can cause race conditions
			//
			// 1. process a writes secret
			// 2. process b marks secret as deleted
			// 3. in memory we write then delete
			// 4. but to the backing store it goes as delete then write.
			// 5. memory: secret deleted; backing store: secret exists.
			//
			// to solve it; have a queue of operations (as a go channel)
			// and do not consume the next operation until the current
			// one is complete.
			//
			// have one channel for each resource:
			// - secrets
			// - policies
			// - key recovery info.
			//
			// Or as an alternative; make these async operations sync
			// and wait for them to complete before reporting success.
			// this will make the architecture way simpler without needing
			// to rely on channels.
		}()
	}
}
