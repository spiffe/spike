//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"context"
	"sync"
	"time"

	"github.com/spiffe/spike-sdk-go/log"
)

// Sweeper handles background rewrapping of secrets with outdated KEKs
type Sweeper struct {
	manager  *Manager
	storage  SweeperStorage
	policy   *RotationPolicy
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// SweeperStorage is the interface for accessing secrets that need rewrapping
type SweeperStorage interface {
	// ListSecretsWithKEK returns paths of secrets using a specific KEK
	ListSecretsWithKEK(ctx context.Context, kekID string) ([]SecretPath, error)

	// RewrapSecret rewraps a specific secret version
	RewrapSecret(ctx context.Context, path string, version int) error
}

// SecretPath represents a secret path and version
type SecretPath struct {
	Path    string
	Version int
}

// NewSweeper creates a new background sweeper
func NewSweeper(
	manager *Manager,
	storage SweeperStorage,
	policy *RotationPolicy,
) *Sweeper {
	if policy == nil {
		policy = DefaultRotationPolicy()
	}

	return &Sweeper{
		manager: manager,
		storage: storage,
		policy:  policy,
		stopCh:  make(chan struct{}),
	}
}

// Start begins the background sweeper
func (s *Sweeper) Start(ctx context.Context) {
	const fName = "Sweeper.Start"

	if !s.policy.LazyRewrapEnabled {
		log.Log().Info(fName, "message", "lazy rewrap disabled, sweeper not starting")
		return
	}

	log.Log().Info(fName, "message", "starting KEK sweeper")

	s.wg.Add(1)
	go s.run(ctx)
}

// Stop gracefully stops the sweeper
func (s *Sweeper) Stop() {
	const fName = "Sweeper.Stop"

	s.stopOnce.Do(func() {
		log.Log().Info(fName, "message", "stopping KEK sweeper")
		close(s.stopCh)
		s.wg.Wait()
		log.Log().Info(fName, "message", "KEK sweeper stopped")
	})
}

// run is the main sweeper loop
func (s *Sweeper) run(ctx context.Context) {
	const fName = "Sweeper.run"
	defer s.wg.Done()

	// Sweep interval: run every hour
	sweepInterval := time.Hour
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()

	log.Log().Info(fName, "message", "KEK sweeper started", "interval", sweepInterval)

	for {
		select {
		case <-ctx.Done():
			log.Log().Info(fName, "message", "context cancelled, stopping sweeper")
			return
		case <-s.stopCh:
			log.Log().Info(fName, "message", "stop signal received")
			return
		case <-ticker.C:
			if err := s.sweep(ctx); err != nil {
				log.Log().Error(fName, "message", "sweep failed", "err", err.Error())
			}
		}
	}
}

// sweep performs a single sweep operation
func (s *Sweeper) sweep(ctx context.Context) error {
	const fName = "Sweeper.sweep"

	currentKekID := s.manager.GetCurrentKEKID()
	if currentKekID == "" {
		log.Log().Warn(fName, "message", "no current KEK, skipping sweep")
		return nil
	}

	log.Log().Info(fName, "message", "starting sweep", "current_kek", currentKekID)

	// Find all KEKs in grace period
	s.manager.mu.RLock()
	graceKeks := make([]string, 0)
	for kekID, meta := range s.manager.metadata {
		if meta.Status == KekStatusGrace {
			graceKeks = append(graceKeks, kekID)
		}
	}
	s.manager.mu.RUnlock()

	if len(graceKeks) == 0 {
		log.Log().Info(fName, "message", "no KEKs in grace period, nothing to rewrap")
		return nil
	}

	log.Log().Info(fName, "message", "found KEKs in grace period", "count", len(graceKeks))

	// Rewrap secrets for each grace period KEK
	totalRewrapped := 0
	for _, oldKekID := range graceKeks {
		rewrapped, err := s.rewrapSecretsForKEK(ctx, oldKekID)
		if err != nil {
			log.Log().Error(fName,
				"message", "failed to rewrap secrets for KEK",
				"kek_id", oldKekID,
				"err", err.Error())
			continue
		}
		totalRewrapped += rewrapped
	}

	log.Log().Info(fName,
		"message", "sweep completed",
		"total_rewrapped", totalRewrapped,
		"grace_keks", len(graceKeks))

	// Clean up KEKs that have exceeded grace period
	if err := s.manager.CleanupGracePeriodKEKs(); err != nil {
		log.Log().Error(fName, "message", "failed to cleanup grace period KEKs", "err", err.Error())
	}

	return nil
}

// rewrapSecretsForKEK rewraps all secrets using a specific KEK
func (s *Sweeper) rewrapSecretsForKEK(ctx context.Context, oldKekID string) (int, error) {
	const fName = "Sweeper.rewrapSecretsForKEK"

	// Get list of secrets using this KEK
	secrets, err := s.storage.ListSecretsWithKEK(ctx, oldKekID)
	if err != nil {
		return 0, err
	}

	if len(secrets) == 0 {
		log.Log().Info(fName, "message", "no secrets found for KEK", "kek_id", oldKekID)
		return 0, nil
	}

	log.Log().Info(fName,
		"message", "rewrapping secrets",
		"kek_id", oldKekID,
		"count", len(secrets))

	// Rate limiter
	qpsLimit := s.policy.MaxRewrapQPS
	if qpsLimit <= 0 {
		qpsLimit = 100
	}
	throttle := time.NewTicker(time.Second / time.Duration(qpsLimit))
	defer throttle.Stop()

	rewrapped := 0
	for _, secret := range secrets {
		// Rate limit
		select {
		case <-ctx.Done():
			log.Log().Info(fName, "message", "context cancelled during rewrap")
			return rewrapped, ctx.Err()
		case <-s.stopCh:
			log.Log().Info(fName, "message", "stop signal received during rewrap")
			return rewrapped, nil
		case <-throttle.C:
			// Continue
		}

		if err := s.storage.RewrapSecret(ctx, secret.Path, secret.Version); err != nil {
			log.Log().Error(fName,
				"message", "failed to rewrap secret",
				"path", secret.Path,
				"version", secret.Version,
				"err", err.Error())
			continue
		}

		rewrapped++

		if rewrapped%100 == 0 {
			log.Log().Info(fName,
				"message", "rewrap progress",
				"kek_id", oldKekID,
				"rewrapped", rewrapped,
				"total", len(secrets))
		}
	}

	log.Log().Info(fName,
		"message", "completed rewrapping for KEK",
		"kek_id", oldKekID,
		"rewrapped", rewrapped,
		"total", len(secrets))

	return rewrapped, nil
}
