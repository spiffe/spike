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

// Scheduler handles automatic KEK rotation based on policy
type Scheduler struct {
	manager  *Manager
	policy   *RotationPolicy
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// NewScheduler creates a new KEK rotation scheduler
func NewScheduler(manager *Manager, policy *RotationPolicy) *Scheduler {
	if policy == nil {
		policy = DefaultRotationPolicy()
	}

	return &Scheduler{
		manager: manager,
		policy:  policy,
		stopCh:  make(chan struct{}),
	}
}

// Start begins the rotation scheduler
func (s *Scheduler) Start(ctx context.Context) {
	const fName = "Scheduler.Start"

	log.Log().Info(fName, "message", "starting KEK rotation scheduler")

	s.wg.Add(1)
	go s.run(ctx)
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	const fName = "Scheduler.Stop"

	s.stopOnce.Do(func() {
		log.Log().Info(fName, "message", "stopping KEK rotation scheduler")
		close(s.stopCh)
		s.wg.Wait()
		log.Log().Info(fName, "message", "KEK rotation scheduler stopped")
	})
}

// run is the main scheduler loop
func (s *Scheduler) run(ctx context.Context) {
	const fName = "Scheduler.run"
	defer s.wg.Done()

	// Check for rotation every hour
	checkInterval := time.Hour
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	log.Log().Info(fName, "message", "KEK rotation scheduler started", "check_interval", checkInterval)

	// Do an initial check
	s.checkAndRotate()

	for {
		select {
		case <-ctx.Done():
			log.Log().Info(fName, "message", "context cancelled, stopping scheduler")
			return
		case <-s.stopCh:
			log.Log().Info(fName, "message", "stop signal received")
			return
		case <-ticker.C:
			s.checkAndRotate()
		}
	}
}

// checkAndRotate checks if rotation is needed and performs it
func (s *Scheduler) checkAndRotate() {
	const fName = "Scheduler.checkAndRotate"

	if s.manager.ShouldRotate() {
		log.Log().Info(fName, "message", "KEK rotation needed, initiating rotation")

		if err := s.manager.RotateKEK(); err != nil {
			log.Log().Error(fName,
				"message", "KEK rotation failed",
				"err", err.Error())
			return
		}

		currentKekID := s.manager.GetCurrentKEKID()
		log.Log().Info(fName,
			"message", "KEK rotation completed successfully",
			"new_kek_id", currentKekID)
	}
}
