//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"sync"
	"time"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite"
	"github.com/spiffe/spike/app/nexus/internal/state/store"
	"github.com/spiffe/spike/internal/log"
)

func AsyncPersistSecret(kv *store.KV, path string) {
	be := Backend()

	// Get the full secret for caching
	secret := kv.GetRawSecret(path)
	if secret != nil {
		go func() {
			if be == nil {
				return // No cache available
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

func ReadSecret(path string, version int) *store.Secret {
	be := Backend()
	if be == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

		if sv, ok := cachedSecret.Versions[version]; ok && sv.DeletedTime == nil {
			return cachedSecret
		}
	}

	return nil
}

func InitializeSqliteBackend(rootKey string) backend.Backend {
	// Create SQLite backend configuration
	cfg := backend.Config{
		EncryptionKey: rootKey, // Use the root key for encryption
		Location:      "cache", // Or wherever you want to store the DB
		Options: map[string]any{
			"data_dir":                  "./.data/backend", // Or use os.UserCacheDir()
			"database_file":             "spike.db",
			"journal_mode":              "WAL", // Write-Ahead Logging for better concurrency
			"busy_timeout_ms":           5000,  // 5 second busy timeout
			"max_open_conns":            10,    // Adjust based on your needs
			"max_idle_conns":            5,
			"conn_max_lifetime_seconds": 3600, // 1 hour
		},
	}

	// Initialize SQLite backend
	dbBackend, err := sqlite.New(cfg)
	if err != nil {
		// Log error but don't fail initialization
		// The system can still work with just in-memory state
		log.Log().Warn("initializeSqliteBackend",
			"msg", "Failed to create SQLite backend",
			"err", err.Error(),
		)
		return nil
	}

	ctxC, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := dbBackend.Initialize(ctxC); err != nil {
		log.Log().Warn("initializeSqliteBackend",
			"msg", "Failed to initialize SQLite backend",
			"err", err.Error(),
		)
		return nil
	}

	return dbBackend
}

func ReadAdminToken() string {
	be := Backend()
	if be == nil {
		return ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cachedToken, err := be.LoadAdminToken(ctx)
	if err != nil {
		// Log error but continue - memory is source of truth
		log.Log().Warn("readAdminToken",
			"msg", "Failed to load admin token from cache",
			"err", err.Error(),
		)
		return ""
	}

	return cachedToken
}

func AsyncPersistAdminToken(token string) {
	go func() {
		be := Backend()
		if be == nil {
			return // No cache available
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := be.StoreAdminToken(ctx, token); err != nil {
			// Log error but continue - memory is source of truth
			log.Log().Warn("asyncPersistAdminToken",
				"msg", "Failed to cache admin token",
				"err", err.Error(),
			)
		}
	}()
}

var (
	sqliteBackend *sqlite.Backend
	backendMu     sync.RWMutex
)

func InitializeBackend(rootKey string) backend.Backend {
	backendMu.Lock()
	defer backendMu.Unlock()

	// TODO: create backend based on env config.
	newBackend := InitializeSqliteBackend(rootKey)

	sqliteBackend = newBackend.(*sqlite.Backend)
	return sqliteBackend
}

func Backend() backend.Backend {
	backendMu.RLock()
	defer backendMu.RUnlock()
	// TODO: create backend based on env config.
	return sqliteBackend
}
