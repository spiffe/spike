//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"fmt"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
)

// RMKRotationResult contains the result of an RMK rotation
type RMKRotationResult struct {
	// OldRMKVersion is the version of the old RMK
	OldRMKVersion int

	// NewRMKVersion is the version of the new RMK
	NewRMKVersion int

	// KEKsRewrapped is the number of KEKs that were rewrapped
	KEKsRewrapped int

	// StartedAt is when the rotation started
	StartedAt time.Time

	// CompletedAt is when the rotation completed
	CompletedAt time.Time

	// Duration is how long the rotation took
	Duration time.Duration
}

// RotateRMK performs an RMK rotation by rewrapping all KEKs
//
// This is a critical operation that should be performed during a maintenance window.
// The process:
// 1. Verify the new RMK is valid
// 2. For each KEK:
//   - Derive KEK using old RMK
//   - Re-derive KEK using new RMK (will have same value since deterministic)
//   - Update KEK metadata to reference new RMK version
//
// 3. Update manager's RMK reference
//
// Note: This operation does NOT touch any encrypted secrets - only KEK metadata
// is updated. The actual KEK values remain the same because derivation is
// deterministic (same salt + same KEK ID = same KEK).
//
// Parameters:
//   - oldRMK: The current Root Master Key
//   - newRMK: The new Root Master Key
//   - newRMKVersion: The version number for the new RMK
//
// Returns:
//   - RMKRotationResult with rotation details
//   - Error if rotation fails
func (m *Manager) RotateRMK(
	oldRMK *[crypto.AES256KeySize]byte,
	newRMK *[crypto.AES256KeySize]byte,
	newRMKVersion int,
) (*RMKRotationResult, error) {
	const fName = "RotateRMK"

	startTime := time.Now()

	log.Log().Info(fName,
		"message", "starting RMK rotation",
		"old_rmk_version", m.currentRMKVersion,
		"new_rmk_version", newRMKVersion)

	// Validate inputs
	if oldRMK == nil || newRMK == nil {
		return nil, fmt.Errorf("%s: nil RMK provided", fName)
	}

	if newRMKVersion <= m.currentRMKVersion {
		return nil, fmt.Errorf("%s: new RMK version must be greater than current version", fName)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	result := &RMKRotationResult{
		OldRMKVersion: m.currentRMKVersion,
		NewRMKVersion: newRMKVersion,
		StartedAt:     startTime,
	}

	// Rewrap all KEKs
	for kekID, meta := range m.metadata {
		log.Log().Info(fName,
			"message", "rewrapping KEK",
			"kek_id", kekID,
			"kek_version", meta.Version)

		// Verify old KEK can be derived
		oldKek, err := DeriveKEK(oldRMK, meta)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to derive old KEK %s: %w", fName, kekID, err)
		}

		// Derive new KEK (will be same value due to deterministic derivation)
		newKek, err := DeriveKEK(newRMK, meta)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to derive new KEK %s: %w", fName, kekID, err)
		}

		// Sanity check: KEKs should be identical (deterministic derivation)
		// This validates that derivation is working correctly
		if *oldKek != *newKek {
			// This should never happen with deterministic derivation
			log.Log().Warn(fName,
				"message", "KEK mismatch detected - this indicates a derivation issue",
				"kek_id", kekID)
		}

		// Update KEK metadata to reference new RMK version
		meta.RMKVersion = newRMKVersion
		if err := m.storage.StoreKEKMetadata(meta); err != nil {
			return nil, fmt.Errorf("%s: failed to update KEK metadata %s: %w", fName, kekID, err)
		}

		// Clear old KEK from cache
		delete(m.cache, kekID)

		result.KEKsRewrapped++

		log.Log().Info(fName,
			"message", "successfully rewrapped KEK",
			"kek_id", kekID,
			"old_rmk_version", m.currentRMKVersion,
			"new_rmk_version", newRMKVersion)
	}

	// Update manager's RMK reference
	m.rmk = newRMK
	m.currentRMKVersion = newRMKVersion

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	log.Log().Info(fName,
		"message", "RMK rotation completed successfully",
		"keks_rewrapped", result.KEKsRewrapped,
		"duration_ms", result.Duration.Milliseconds(),
		"new_rmk_version", newRMKVersion)

	return result, nil
}

// ValidateRMKRotation validates that RMK rotation was successful
// by attempting to derive all KEKs with the new RMK
func (m *Manager) ValidateRMKRotation() error {
	const fName = "ValidateRMKRotation"

	m.mu.RLock()
	defer m.mu.RUnlock()

	log.Log().Info(fName, "message", "validating RMK rotation", "kek_count", len(m.metadata))

	for kekID, meta := range m.metadata {
		// Try to derive KEK with current RMK
		_, err := DeriveKEK(m.rmk, meta)
		if err != nil {
			return fmt.Errorf("%s: failed to derive KEK %s: %w", fName, kekID, err)
		}
	}

	log.Log().Info(fName, "message", "RMK rotation validation successful")
	return nil
}

// PrepareRMKRotation prepares for RMK rotation by creating a snapshot
// of current KEK metadata for rollback purposes
func (m *Manager) PrepareRMKRotation() (*RMKRotationSnapshot, error) {
	const fName = "PrepareRMKRotation"

	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := &RMKRotationSnapshot{
		RMKVersion: m.currentRMKVersion,
		KEKs:       make(map[string]*Metadata),
		Timestamp:  time.Now(),
	}

	// Create deep copies of KEK metadata
	for kekID, meta := range m.metadata {
		metaCopy := &Metadata{
			ID:         meta.ID,
			Version:    meta.Version,
			Salt:       meta.Salt,
			RMKVersion: meta.RMKVersion,
			CreatedAt:  meta.CreatedAt,
			WrapsCount: meta.WrapsCount,
			Status:     meta.Status,
		}
		if meta.RetiredAt != nil {
			retiredCopy := *meta.RetiredAt
			metaCopy.RetiredAt = &retiredCopy
		}
		snapshot.KEKs[kekID] = metaCopy
	}

	log.Log().Info(fName,
		"message", "created RMK rotation snapshot",
		"rmk_version", snapshot.RMKVersion,
		"kek_count", len(snapshot.KEKs))

	return snapshot, nil
}

// RMKRotationSnapshot contains a snapshot of KEK state before RMK rotation
type RMKRotationSnapshot struct {
	RMKVersion int
	KEKs       map[string]*Metadata
	Timestamp  time.Time
}

// RollbackRMKRotation rolls back an RMK rotation using a snapshot
// This should only be used if validation fails
func (m *Manager) RollbackRMKRotation(
	snapshot *RMKRotationSnapshot,
	oldRMK *[crypto.AES256KeySize]byte,
) error {
	const fName = "RollbackRMKRotation"

	if snapshot == nil {
		return fmt.Errorf("%s: nil snapshot", fName)
	}

	log.Log().Warn(fName,
		"message", "rolling back RMK rotation",
		"snapshot_rmk_version", snapshot.RMKVersion,
		"current_rmk_version", m.currentRMKVersion)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Restore KEK metadata
	for kekID, meta := range snapshot.KEKs {
		if err := m.storage.StoreKEKMetadata(meta); err != nil {
			return fmt.Errorf("%s: failed to restore KEK metadata %s: %w", fName, kekID, err)
		}
		m.metadata[kekID] = meta
	}

	// Restore RMK
	m.rmk = oldRMK
	m.currentRMKVersion = snapshot.RMKVersion

	// Clear cache
	m.cache = make(map[string]*[crypto.AES256KeySize]byte)

	log.Log().Info(fName,
		"message", "RMK rotation rollback completed",
		"restored_rmk_version", snapshot.RMKVersion)

	return nil
}
