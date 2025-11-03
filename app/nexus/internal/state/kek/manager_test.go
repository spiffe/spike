//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"crypto/rand"
	"sync"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
)

// mockStorage is a mock implementation of Storage for testing
type mockStorage struct {
	mu       sync.RWMutex
	metadata map[string]*Metadata
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		metadata: make(map[string]*Metadata),
	}
}

func (m *mockStorage) StoreKEKMetadata(metadata *Metadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata[metadata.ID] = metadata
	return nil
}

func (m *mockStorage) LoadKEKMetadata(kekID string) (*Metadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	meta, ok := m.metadata[kekID]
	if !ok {
		return nil, nil
	}
	return meta, nil
}

func (m *mockStorage) ListKEKMetadata() ([]*Metadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*Metadata, 0, len(m.metadata))
	for _, meta := range m.metadata {
		result = append(result, meta)
	}
	return result, nil
}

func (m *mockStorage) UpdateKEKWrapsCount(kekID string, delta int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if meta, ok := m.metadata[kekID]; ok {
		meta.WrapsCount += delta
	}
	return nil
}

func (m *mockStorage) UpdateKEKStatus(kekID string, status KekStatus, retiredAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if meta, ok := m.metadata[kekID]; ok {
		meta.Status = status
		meta.RetiredAt = retiredAt
	}
	return nil
}

func generateTestRMK() *[crypto.AES256KeySize]byte {
	rmk := new([crypto.AES256KeySize]byte)
	if _, err := rand.Read(rmk[:]); err != nil {
		panic(err)
	}
	return rmk
}

func TestManagerCreation(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := DefaultRotationPolicy()

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	if manager.GetCurrentKEKID() == "" {
		t.Error("expected manager to have a current KEK ID")
	}

	currentMeta, err := manager.GetMetadata(manager.GetCurrentKEKID())
	if err != nil {
		t.Fatalf("failed to get current KEK metadata: %v", err)
	}

	if currentMeta.Status != KekStatusActive {
		t.Errorf("expected current KEK status to be active, got %s", currentMeta.Status)
	}

	if currentMeta.Version != 1 {
		t.Errorf("expected KEK version 1, got %d", currentMeta.Version)
	}
}

func TestKEKRotation(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := DefaultRotationPolicy()

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	firstKEKID := manager.GetCurrentKEKID()

	// Rotate KEK
	err = manager.RotateKEK()
	if err != nil {
		t.Fatalf("failed to rotate KEK: %v", err)
	}

	secondKEKID := manager.GetCurrentKEKID()

	if firstKEKID == secondKEKID {
		t.Error("expected KEK ID to change after rotation")
	}

	// Check first KEK is in grace period
	firstMeta, err := manager.GetMetadata(firstKEKID)
	if err != nil {
		t.Fatalf("failed to get first KEK metadata: %v", err)
	}

	if firstMeta.Status != KekStatusGrace {
		t.Errorf("expected first KEK to be in grace period, got %s", firstMeta.Status)
	}

	// Check second KEK is active
	secondMeta, err := manager.GetMetadata(secondKEKID)
	if err != nil {
		t.Fatalf("failed to get second KEK metadata: %v", err)
	}

	if secondMeta.Status != KekStatusActive {
		t.Errorf("expected second KEK to be active, got %s", secondMeta.Status)
	}
}

func TestKEKCaching(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := DefaultRotationPolicy()

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	kekID := manager.GetCurrentKEKID()

	// Get KEK first time (should derive)
	kek1, err := manager.GetKEK(kekID)
	if err != nil {
		t.Fatalf("failed to get KEK: %v", err)
	}

	// Get KEK second time (should use cache)
	kek2, err := manager.GetKEK(kekID)
	if err != nil {
		t.Fatalf("failed to get KEK from cache: %v", err)
	}

	// Should be the same pointer (cached)
	if kek1 != kek2 {
		t.Error("expected KEK to be cached")
	}

	// Verify KEK values are identical
	for i := 0; i < crypto.AES256KeySize; i++ {
		if kek1[i] != kek2[i] {
			t.Error("KEK values differ between calls")
			break
		}
	}
}

func TestWrapsCountIncrement(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := DefaultRotationPolicy()

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	kekID := manager.GetCurrentKEKID()

	// Get initial wraps count
	meta1, err := manager.GetMetadata(kekID)
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}

	initialCount := meta1.WrapsCount

	// Increment wraps count multiple times
	const increments = 5
	for i := 0; i < increments; i++ {
		err = manager.IncrementWrapsCount(kekID)
		if err != nil {
			t.Fatalf("failed to increment wraps count: %v", err)
		}
	}

	// Get updated wraps count
	meta2, err := manager.GetMetadata(kekID)
	if err != nil {
		t.Fatalf("failed to get updated metadata: %v", err)
	}

	expectedCount := initialCount + increments
	if meta2.WrapsCount != expectedCount {
		t.Errorf("expected wraps count to be %d, got %d", expectedCount, meta2.WrapsCount)
	}
}

func TestShouldRotateTimeBased(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := &RotationPolicy{
		RotationDays:      1, // 1 day for testing
		MaxWraps:          1000000,
		GraceDays:         180,
		LazyRewrapEnabled: true,
		MaxRewrapQPS:      100,
	}

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Should not need rotation initially
	if manager.ShouldRotate() {
		t.Error("expected new KEK not to need rotation")
	}

	// Manually age the KEK for testing
	kekID := manager.GetCurrentKEKID()
	meta, err := manager.GetMetadata(kekID)
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}

	// Set created time to 2 days ago
	meta.CreatedAt = time.Now().Add(-48 * time.Hour)

	// Now should need rotation
	if !manager.ShouldRotate() {
		t.Error("expected old KEK to need rotation")
	}
}

func TestShouldRotateUsageBased(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := &RotationPolicy{
		RotationDays:      90,
		MaxWraps:          10, // Low number for testing
		GraceDays:         180,
		LazyRewrapEnabled: true,
		MaxRewrapQPS:      100,
	}

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Should not need rotation initially
	if manager.ShouldRotate() {
		t.Error("expected new KEK not to need rotation")
	}

	// Increment wraps count beyond threshold
	kekID := manager.GetCurrentKEKID()
	for i := 0; i < 11; i++ {
		err = manager.IncrementWrapsCount(kekID)
		if err != nil {
			t.Fatalf("failed to increment wraps count: %v", err)
		}
	}

	// Now should need rotation
	if !manager.ShouldRotate() {
		t.Error("expected KEK with high usage to need rotation")
	}
}

func TestListAllKEKs(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := DefaultRotationPolicy()

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Initially should have 1 KEK
	keks, err := manager.ListAllKEKs()
	if err != nil {
		t.Fatalf("failed to list KEKs: %v", err)
	}

	if len(keks) != 1 {
		t.Errorf("expected 1 KEK, got %d", len(keks))
	}

	// Rotate twice
	err = manager.RotateKEK()
	if err != nil {
		t.Fatalf("failed to rotate KEK: %v", err)
	}

	err = manager.RotateKEK()
	if err != nil {
		t.Fatalf("failed to rotate KEK again: %v", err)
	}

	// Should now have 3 KEKs
	keks, err = manager.ListAllKEKs()
	if err != nil {
		t.Fatalf("failed to list KEKs: %v", err)
	}

	if len(keks) != 3 {
		t.Errorf("expected 3 KEKs, got %d", len(keks))
	}

	// Count KEKs by status
	statusCounts := make(map[KekStatus]int)
	for _, kek := range keks {
		statusCounts[kek.Status]++
	}

	if statusCounts[KekStatusActive] != 1 {
		t.Errorf("expected 1 active KEK, got %d", statusCounts[KekStatusActive])
	}

	if statusCounts[KekStatusGrace] != 2 {
		t.Errorf("expected 2 grace period KEKs, got %d", statusCounts[KekStatusGrace])
	}
}

func TestGracePeriodCleanup(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := &RotationPolicy{
		RotationDays:      90,
		MaxWraps:          1000000,
		GraceDays:         1, // 1 day for testing
		LazyRewrapEnabled: true,
		MaxRewrapQPS:      100,
	}

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Rotate to create a KEK in grace period
	err = manager.RotateKEK()
	if err != nil {
		t.Fatalf("failed to rotate KEK: %v", err)
	}

	keks, err := manager.ListAllKEKs()
	if err != nil {
		t.Fatalf("failed to list KEKs: %v", err)
	}

	// Find the grace period KEK
	var graceKEKID string
	for _, kek := range keks {
		if kek.Status == KekStatusGrace {
			graceKEKID = kek.ID
			break
		}
	}

	if graceKEKID == "" {
		t.Fatal("no KEK in grace period found")
	}

	// Manually age the grace period KEK
	graceMeta, err := manager.GetMetadata(graceKEKID)
	if err != nil {
		t.Fatalf("failed to get grace KEK metadata: %v", err)
	}
	graceMeta.CreatedAt = time.Now().Add(-48 * time.Hour) // 2 days ago

	// Run cleanup
	err = manager.CleanupGracePeriodKEKs()
	if err != nil {
		t.Fatalf("failed to cleanup grace period KEKs: %v", err)
	}

	// Check that the KEK is now retired
	updatedMeta, err := manager.GetMetadata(graceKEKID)
	if err != nil {
		t.Fatalf("failed to get updated metadata: %v", err)
	}

	if updatedMeta.Status != KekStatusRetired {
		t.Errorf("expected KEK to be retired, got %s", updatedMeta.Status)
	}

	if updatedMeta.RetiredAt == nil {
		t.Error("expected RetiredAt to be set")
	}
}

func TestConcurrentKEKAccess(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := DefaultRotationPolicy()

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	kekID := manager.GetCurrentKEKID()

	// Concurrent reads
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, err := manager.GetKEK(kekID)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
	}
}

func TestKEKDeterminism(t *testing.T) {
	rmk := generateTestRMK()
	storage := newMockStorage()
	policy := DefaultRotationPolicy()

	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	kekID := manager.GetCurrentKEKID()
	meta, err := manager.GetMetadata(kekID)
	if err != nil {
		t.Fatalf("failed to get metadata: %v", err)
	}

	// Derive KEK twice with same metadata
	kek1, err := DeriveKEK(rmk, meta)
	if err != nil {
		t.Fatalf("failed to derive KEK: %v", err)
	}

	kek2, err := DeriveKEK(rmk, meta)
	if err != nil {
		t.Fatalf("failed to derive KEK second time: %v", err)
	}

	// Should produce identical keys
	for i := 0; i < crypto.AES256KeySize; i++ {
		if kek1[i] != kek2[i] {
			t.Error("KEK derivation is not deterministic")
			break
		}
	}
}

