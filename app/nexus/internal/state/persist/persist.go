//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"github.com/spiffe/spike/app/nexus/internal/env"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/memory"
	"sync"
	"time"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite"
	"github.com/spiffe/spike/app/nexus/internal/state/store"
	"github.com/spiffe/spike/internal/log"
)

var (
	memoryBackend *memory.NoopStore
	sqliteBackend *sqlite.DataStore
	backendMu     sync.RWMutex
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
	opts := map[string]any{}

	opts[sqlite.KeyDataDir] = env.DatabaseDir()
	opts[sqlite.KeyDatabaseFile] = "spike.db"
	opts[sqlite.KeyJournalMode] = env.DatabaseJournalMode()
	opts[sqlite.KeyBusyTimeoutMs] = env.DatabaseBusyTimeoutMs()
	opts[sqlite.KeyMaxOpenConns] = env.DatabaseMaxOpenConns()
	opts[sqlite.KeyMaxIdleConns] = env.DatabaseMaxIdleConns()
	opts[sqlite.KeyConnMaxLifetimeSeconds] = env.DatabaseConnMaxLifetimeSec()

	// Create SQLite backend configuration
	cfg := backend.Config{
		// Use the root key as the encryption key
		EncryptionKey: rootKey,
		Options:       opts,
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

func InitializeBackend(rootKey string) backend.Backend {
	backendMu.Lock()
	defer backendMu.Unlock()

	storeType := env.BackendStoreType()

	switch storeType {
	case env.Memory:
		return &memory.NoopStore{}
	case env.Sqlite:
		return InitializeSqliteBackend(rootKey)
	default:
		return &memory.NoopStore{}
	}
}

func Backend() backend.Backend {
	backendMu.RLock()
	defer backendMu.RUnlock()

	storeType := env.BackendStoreType()

	switch storeType {
	case env.Memory:
		return memoryBackend
	case env.Sqlite:
		return sqliteBackend
	default:
		return memoryBackend
	}
}
