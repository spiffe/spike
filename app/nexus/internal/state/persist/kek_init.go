//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"sync"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/kek"
	"github.com/spiffe/spike/internal/config"
)

var (
	kekManager     *kek.Manager
	kekScheduler   *kek.Scheduler
	kekSweeper     *kek.Sweeper
	kekManagerMu   sync.RWMutex
	kekInitialized bool
)

// InitializeKEKManager initializes the KEK manager for envelope encryption
//
// This should be called after InitializeBackend but only for SQLite backend.
// It sets up:
// - KEK manager with rotation policy
// - Automatic rotation scheduler
// - Background sweeper for lazy rewrapping
//
// Parameters:
//   - rootKey: The Root Master Key
//   - rmkVersion: The version of the RMK (default: 1)
//
// Returns:
//   - error if initialization fails
func InitializeKEKManager(
	rootKey *[crypto.AES256KeySize]byte,
	rmkVersion int,
) error {
	const fName = "InitializeKEKManager"

	// Only initialize for SQLite backend
	storeType := env.BackendStoreTypeVal()
	if storeType != env.Sqlite {
		log.Log().Info(fName,
			"message", "KEK manager not needed for store type",
			"store_type", storeType)
		return nil
	}

	// Check if KEK rotation is enabled
	if !config.KEKRotationEnabled() {
		log.Log().Info(fName, "message", "KEK rotation disabled by configuration")
		return nil
	}

	kekManagerMu.Lock()
	defer kekManagerMu.Unlock()

	if kekInitialized {
		log.Log().Warn(fName, "message", "KEK manager already initialized")
		return nil
	}

	log.Log().Info(fName, "message", "initializing KEK manager")

	// Get the backend (must be SQLite)
	backendMu.RLock()
	currentBackend := be
	backendMu.RUnlock()

	if currentBackend == nil {
		log.Log().Warn(fName, "message", "backend not initialized, skipping KEK manager")
		return nil
	}

	// Unwrap to get the underlying SQLite backend
	var sqliteBackend interface{}
	if wrapped, ok := currentBackend.(*EnvelopeAwareBackend); ok {
		sqliteBackend = wrapped.backend
	} else {
		sqliteBackend = currentBackend
	}

	// Create KEK storage adapter
	var kekStorage kek.Storage
	switch backend := sqliteBackend.(type) {
	case interface{ StoreKEKMetadata(*kek.Metadata) error }:
		kekStorage = backend.(kek.Storage)
	default:
		log.Log().Warn(fName,
			"message", "backend does not support KEK storage, skipping KEK manager")
		return nil
	}

	// Create rotation policy from environment
	policy := &kek.RotationPolicy{
		RotationDays:      config.KEKRotationDays(),
		MaxWraps:          config.KEKMaxWraps(),
		GraceDays:         config.KEKGraceDays(),
		LazyRewrapEnabled: config.KEKLazyRewrapEnabled(),
		MaxRewrapQPS:      config.KEKMaxRewrapQPS(),
	}

	// Create KEK manager
	var err error
	kekManager, err = kek.NewManager(rootKey, rmkVersion, policy, kekStorage)
	if err != nil {
		log.Log().Error(fName,
			"message", "failed to create KEK manager",
			"err", err.Error())
		return err
	}

	// Set KEK manager on the underlying SQLite backend
	switch backend := sqliteBackend.(type) {
	case interface{ SetKEKManager(*kek.Manager) }:
		backend.SetKEKManager(kekManager)
		log.Log().Info(fName, "message", "KEK manager set on backend")
	}

	// Create and start scheduler
	kekScheduler = kek.NewScheduler(kekManager, policy)
	go kekScheduler.Start(context.Background())

	// Create and start sweeper
	if policy.LazyRewrapEnabled {
		var sweeperStorage kek.SweeperStorage
		switch backend := sqliteBackend.(type) {
		case interface {
			ListSecretsWithKEK(context.Context, string) ([]kek.SecretPath, error)
		}:
			sweeperStorage = backend.(kek.SweeperStorage)
		}

		if sweeperStorage != nil {
			kekSweeper = kek.NewSweeper(kekManager, sweeperStorage, policy)
			go kekSweeper.Start(context.Background())
		}
	}

	kekInitialized = true

	log.Log().Info(fName,
		"message", "KEK manager initialized successfully",
		"rotation_days", policy.RotationDays,
		"max_wraps", policy.MaxWraps,
		"grace_days", policy.GraceDays,
		"lazy_rewrap", policy.LazyRewrapEnabled)

	return nil
}

// ShutdownKEKManager gracefully shuts down the KEK manager and its components
func ShutdownKEKManager() {
	const fName = "ShutdownKEKManager"

	kekManagerMu.Lock()
	defer kekManagerMu.Unlock()

	if !kekInitialized {
		return
	}

	log.Log().Info(fName, "message", "shutting down KEK manager")

	if kekScheduler != nil {
		kekScheduler.Stop()
	}

	if kekSweeper != nil {
		kekSweeper.Stop()
	}

	kekInitialized = false

	log.Log().Info(fName, "message", "KEK manager shut down")
}

// GetKEKManager returns the global KEK manager instance
func GetKEKManager() *kek.Manager {
	kekManagerMu.RLock()
	defer kekManagerMu.RUnlock()
	return kekManager
}

// IsKEKManagerInitialized returns whether the KEK manager is initialized
func IsKEKManagerInitialized() bool {
	kekManagerMu.RLock()
	defer kekManagerMu.RUnlock()
	return kekInitialized
}
