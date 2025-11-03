//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"fmt"
	"sync"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/log"
	"github.com/spiffe/spike-sdk-go/security/mem"
)

// Manager handles KEK lifecycle, caching, and rotation
type Manager struct {
	// mu protects concurrent access to the manager
	mu sync.RWMutex

	// metadata stores all KEK metadata indexed by KEK ID
	metadata map[string]*Metadata

	// cache stores derived KEKs in memory
	cache map[string]*[crypto.AES256KeySize]byte

	// currentKekID is the ID of the currently active KEK
	currentKekID string

	// policy defines rotation rules
	policy *RotationPolicy

	// rmk is a reference to the Root Master Key
	rmk *[crypto.AES256KeySize]byte

	// currentRMKVersion tracks the RMK version
	currentRMKVersion int

	// storage is the backend storage interface for KEK metadata
	storage Storage
}

// Storage is the interface for persisting KEK metadata
type Storage interface {
	// StoreKEKMetadata persists KEK metadata
	StoreKEKMetadata(metadata *Metadata) error

	// LoadKEKMetadata retrieves KEK metadata by ID
	LoadKEKMetadata(kekID string) (*Metadata, error)

	// ListKEKMetadata returns all KEK metadata
	ListKEKMetadata() ([]*Metadata, error)

	// UpdateKEKWrapsCount atomically increments the wraps count
	UpdateKEKWrapsCount(kekID string, delta int64) error

	// UpdateKEKStatus updates the KEK status
	UpdateKEKStatus(kekID string, status KekStatus, retiredAt *time.Time) error
}

// NewManager creates a new KEK manager
func NewManager(
	rmk *[crypto.AES256KeySize]byte,
	rmkVersion int,
	policy *RotationPolicy,
	storage Storage,
) (*Manager, error) {
	const fName = "NewManager"

	if rmk == nil || mem.Zeroed32(rmk) {
		return nil, fmt.Errorf("%s: invalid RMK", fName)
	}

	if policy == nil {
		policy = DefaultRotationPolicy()
	}

	m := &Manager{
		metadata:          make(map[string]*Metadata),
		cache:             make(map[string]*[crypto.AES256KeySize]byte),
		policy:            policy,
		rmk:               rmk,
		currentRMKVersion: rmkVersion,
		storage:           storage,
	}

	// Load existing KEK metadata from storage
	if err := m.loadMetadata(); err != nil {
		return nil, fmt.Errorf("%s: failed to load KEK metadata: %w", fName, err)
	}

	// If no KEKs exist, create the first one
	if len(m.metadata) == 0 {
		log.Log().Info(fName, "message", "no KEKs found, creating initial KEK")
		if err := m.createNewKEK(); err != nil {
			return nil, fmt.Errorf("%s: failed to create initial KEK: %w", fName, err)
		}
	} else {
		// Find the current active KEK
		for _, meta := range m.metadata {
			if meta.Status == KekStatusActive {
				m.currentKekID = meta.ID
				break
			}
		}
	}

	log.Log().Info(fName,
		"message", "KEK manager initialized",
		"current_kek", m.currentKekID,
		"total_keks", len(m.metadata))

	return m, nil
}

// loadMetadata loads all KEK metadata from storage
func (m *Manager) loadMetadata() error {
	const fName = "loadMetadata"

	metadataList, err := m.storage.ListKEKMetadata()
	if err != nil {
		return fmt.Errorf("%s: %w", fName, err)
	}

	for _, meta := range metadataList {
		m.metadata[meta.ID] = meta
	}

	return nil
}

// GetCurrentKEKID returns the ID of the current active KEK
func (m *Manager) GetCurrentKEKID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentKekID
}

// GetKEK retrieves a KEK by ID, deriving it if not cached
func (m *Manager) GetKEK(kekID string) (*[crypto.AES256KeySize]byte, error) {
	const fName = "GetKEK"

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check cache first
	if kek, ok := m.cache[kekID]; ok {
		return kek, nil
	}

	// Get metadata
	meta, ok := m.metadata[kekID]
	if !ok {
		return nil, fmt.Errorf("%s: KEK not found: %s", fName, kekID)
	}

	// Derive KEK
	kek, err := DeriveKEK(m.rmk, meta)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fName, err)
	}

	// Cache it
	m.cache[kekID] = kek

	return kek, nil
}

// IncrementWrapsCount atomically increments the wraps count for the current KEK
func (m *Manager) IncrementWrapsCount(kekID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.metadata[kekID]
	if !ok {
		return fmt.Errorf("KEK not found: %s", kekID)
	}

	// Update in storage (storage will update the metadata object)
	return m.storage.UpdateKEKWrapsCount(kekID, 1)
}

// ShouldRotate checks if the current KEK should be rotated
func (m *Manager) ShouldRotate() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentKekID == "" {
		return true
	}

	meta, ok := m.metadata[m.currentKekID]
	if !ok {
		return true
	}

	// Check time-based rotation
	daysSinceCreation := time.Since(meta.CreatedAt).Hours() / 24
	if daysSinceCreation >= float64(m.policy.RotationDays) {
		log.Log().Info("ShouldRotate",
			"message", "KEK rotation needed (time-based)",
			"kek_id", m.currentKekID,
			"days_old", int(daysSinceCreation))
		return true
	}

	// Check usage-based rotation
	if meta.WrapsCount >= m.policy.MaxWraps {
		log.Log().Info("ShouldRotate",
			"message", "KEK rotation needed (usage-based)",
			"kek_id", m.currentKekID,
			"wraps_count", meta.WrapsCount)
		return true
	}

	return false
}

// RotateKEK creates a new KEK and marks the old one as in grace period
func (m *Manager) RotateKEK() error {
	const fName = "RotateKEK"

	m.mu.Lock()
	defer m.mu.Unlock()

	// Move current KEK to grace period
	if m.currentKekID != "" {
		oldMeta := m.metadata[m.currentKekID]
		oldMeta.Status = KekStatusGrace
		if err := m.storage.UpdateKEKStatus(m.currentKekID, KekStatusGrace, nil); err != nil {
			return fmt.Errorf("%s: failed to update old KEK status: %w", fName, err)
		}
		log.Log().Info(fName,
			"message", "moved KEK to grace period",
			"kek_id", m.currentKekID)
	}

	// Create new KEK
	if err := m.createNewKEKUnlocked(); err != nil {
		return fmt.Errorf("%s: %w", fName, err)
	}

	log.Log().Info(fName,
		"message", "KEK rotation completed",
		"new_kek_id", m.currentKekID)

	return nil
}

// createNewKEK creates a new KEK (with lock)
func (m *Manager) createNewKEK() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createNewKEKUnlocked()
}

// createNewKEKUnlocked creates a new KEK (caller must hold lock)
func (m *Manager) createNewKEKUnlocked() error {
	const fName = "createNewKEK"

	// Generate salt
	salt, err := GenerateKEKSalt()
	if err != nil {
		return fmt.Errorf("%s: %w", fName, err)
	}

	// Determine version and ID
	version := len(m.metadata) + 1
	kekID := fmt.Sprintf("v%d-%s", version, time.Now().Format("2006-01"))

	// Create metadata
	meta := &Metadata{
		ID:         kekID,
		Version:    version,
		Salt:       salt,
		RMKVersion: m.currentRMKVersion,
		CreatedAt:  time.Now(),
		WrapsCount: 0,
		Status:     KekStatusActive,
	}

	// Store metadata
	if err := m.storage.StoreKEKMetadata(meta); err != nil {
		return fmt.Errorf("%s: failed to store KEK metadata: %w", fName, err)
	}

	// Update in-memory state
	m.metadata[kekID] = meta
	m.currentKekID = kekID

	log.Log().Info(fName,
		"message", "created new KEK",
		"kek_id", kekID,
		"version", version)

	return nil
}

// CleanupGracePeriodKEKs retires KEKs that have exceeded the grace period
func (m *Manager) CleanupGracePeriodKEKs() error {
	const fName = "CleanupGracePeriodKEKs"

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	graceDuration := time.Duration(m.policy.GraceDays) * 24 * time.Hour

	for _, meta := range m.metadata {
		if meta.Status != KekStatusGrace {
			continue
		}

		if now.Sub(meta.CreatedAt) >= graceDuration {
			retiredAt := now
			meta.Status = KekStatusRetired
			meta.RetiredAt = &retiredAt

			if err := m.storage.UpdateKEKStatus(meta.ID, KekStatusRetired, &retiredAt); err != nil {
				log.Log().Error(fName,
					"message", "failed to retire KEK",
					"kek_id", meta.ID,
					"err", err.Error())
				continue
			}

			// Remove from cache
			delete(m.cache, meta.ID)

			log.Log().Info(fName,
				"message", "retired KEK",
				"kek_id", meta.ID)
		}
	}

	return nil
}

// GetMetadata returns KEK metadata by ID
func (m *Manager) GetMetadata(kekID string) (*Metadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	meta, ok := m.metadata[kekID]
	if !ok {
		return nil, fmt.Errorf("KEK metadata not found: %s", kekID)
	}

	return meta, nil
}

// ListAllKEKs returns all KEK metadata
func (m *Manager) ListAllKEKs() ([]*Metadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Metadata, 0, len(m.metadata))
	for _, meta := range m.metadata {
		result = append(result, meta)
	}

	return result, nil
}
