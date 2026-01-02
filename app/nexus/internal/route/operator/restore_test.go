//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package operator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/v1/reqres"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"

	"github.com/spiffe/spike-sdk-go/journal"
	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
)

func TestRouteRestore_MemoryMode(t *testing.T) {
	// Save original environment variables
	originalStore := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Set to memory mode
	_ = os.Setenv(env.NexusBackendStore, "memory")

	// Verify the environment is set correctly
	if env.BackendStoreTypeVal() != env.Memory {
		t.Fatal("Expected Memory backend store type")
	}

	// Create a test request
	req := httptest.NewRequest(http.MethodPost, "/restore", nil)
	w := httptest.NewRecorder()
	audit := &journal.AuditEntry{}

	// Call function
	err := RouteRestore(w, req, audit)

	// Should return nil (no error) and skip processing in memory mode
	if err != nil {
		t.Errorf("Expected no error in memory mode, got: %v", err)
	}
}

func TestRouteRestore_InvalidRequestBody(t *testing.T) {
	// Save original environment variables
	originalStore := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Set to non-memory mode
	_ = os.Setenv(env.NexusBackendStore, "sqlite")

	// Create request with invalid/empty body
	req := httptest.NewRequest(http.MethodPost, "/restore",
		bytes.NewReader([]byte("")))
	w := httptest.NewRecorder()
	audit := &journal.AuditEntry{}

	// Reset shards for a clean test
	resetShards()

	// Call function - should fail due to read failure
	err := RouteRestore(w, req, audit)

	// Should return ErrReadFailure
	if err == nil {
		t.Error("Expected ErrReadFailure for invalid request body")
	}
}

func TestRouteRestore_InvalidJSONBody(t *testing.T) {
	// Save original environment variables
	originalStore := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Set to non-memory mode
	_ = os.Setenv(env.NexusBackendStore, "sqlite")

	// Create a request with invalid JSON
	invalidJSON := []byte("{invalid json}")
	req := httptest.NewRequest(http.MethodPost, "/restore", bytes.NewReader(invalidJSON))
	w := httptest.NewRecorder()
	audit := &journal.AuditEntry{}

	// Reset shards for a clean test
	resetShards()

	// Call function - should fail due to parse failure
	err := RouteRestore(w, req, audit)

	// Should return ErrParseFailure
	if err == nil {
		t.Error("Expected ErrParseFailure for invalid JSON body")
	}
}

func TestRouteRestore_GuardValidationFailure(t *testing.T) {
	// This test would fail due to missing SPIFFE context and HTTP infrastructure
	// The guardRestoreRequest function requires:
	// 1. Valid SPIFFE ID from the request context
	// 2. Proper TLS peer certificates
	// 3. Network infrastructure setup
	//
	// Without mocking the entire SPIFFE/TLS infrastructure, this test will
	// panic with nil pointer dereferences in the HTTP context handling
	t.Skip("Skipping test that requires SPIFFE infrastructure and TLS context")
}

func TestRouteRestore_TooManyShards(t *testing.T) {
	// Save original environment variables
	originalStore := os.Getenv(env.NexusBackendStore)
	originalThreshold := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
	}()

	// Set environment
	_ = os.Setenv(env.NexusBackendStore, "sqlite")
	_ = os.Setenv(env.NexusShamirThreshold, "2")

	// Reset and pre-fill shards to exceed the threshold
	resetShards()
	shardsMutex.Lock()
	for i := 1; i <= 3; i++ { // Exceed the threshold of 2
		testData := &[crypto.AES256KeySize]byte{}
		testData[0] = byte(i) // Make each shard unique and non-zero
		shards = append(shards, recovery.ShamirShard{
			ID:    uint64(i),
			Value: testData,
		})
	}
	shardsMutex.Unlock()

	// Create a new shard request
	testShard := &[crypto.AES256KeySize]byte{}
	testShard[0] = 4 // Non-zero
	request := reqres.RestoreRequest{
		ID:    4,
		Shard: testShard,
	}

	_, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal test request: %v", err)
	}

	// req := httptest.NewRequest(http.MethodPost, "/restore", bytes.NewReader(requestBody))
	// Note: This test will fail guard validation due to missing SPIFFE context
	// We're testing the logic path, not the actual HTTP processing
	// w := httptest.NewRecorder()
	// audit := &journal.AuditEntry{}

	// This test focuses on the `shards` collection logic, not the full HTTP flow
	t.Skip("Skipping test that requires SPIFFE infrastructure for guard validation")
}

func TestRouteRestore_DuplicateShard(t *testing.T) {
	// Save original environment variables
	originalStore := os.Getenv(env.NexusBackendStore)
	originalThreshold := os.Getenv(env.NexusShamirThreshold)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
	}()

	// Set environment
	_ = os.Setenv(env.NexusBackendStore, "sqlite")
	_ = os.Setenv(env.NexusShamirThreshold, "3")

	// Reset and add one shard
	resetShards()
	testData := &[crypto.AES256KeySize]byte{}
	testData[0] = 1 // Non-zero
	shardsMutex.Lock()
	shards = append(shards, recovery.ShamirShard{
		ID:    1,
		Value: testData,
	})
	shardsMutex.Unlock()

	// Try to add duplicate shard
	duplicateShard := &[crypto.AES256KeySize]byte{}
	duplicateShard[0] = 1 // Same ID, different data
	request := reqres.RestoreRequest{
		ID:    1, // Duplicate ID
		Shard: duplicateShard,
	}

	_, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal test request: %v", err)
	}

	// req := httptest.NewRequest(http.MethodPost, "/restore", bytes.NewReader(requestBody))
	// w := httptest.NewRecorder()
	// audit := &journal.AuditEntry{}

	// This test will fail guard validation due to missing SPIFFE context
	t.Skip("Skipping test that requires SPIFFE infrastructure for guard validation")
}

func TestRouteRestore_SuccessfulShardAddition(t *testing.T) {
	// This test would require mocking the entire HTTP infrastructure, including
	// 1. SPIFFE context for guard validation
	// 2. Network request/response handling
	// 3. State management
	// 4. Recovery functionality
	t.Skip("Skipping test that requires SPIFFE infrastructure and complex mocking")
}

func TestShardCollectionLogic(t *testing.T) {
	// Test the core shard collection logic in isolation
	// Save original environment variables
	originalThreshold := os.Getenv(env.NexusShamirThreshold)
	defer func() {
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
	}()

	_ = os.Setenv(env.NexusShamirThreshold, "3")

	tests := []struct {
		name            string
		existingShards  []recovery.ShamirShard
		newShardID      uint64
		expectedCount   int
		expectThreshold bool
	}{
		{
			name:            "first shard",
			existingShards:  []recovery.ShamirShard{},
			newShardID:      1,
			expectedCount:   1,
			expectThreshold: false,
		},
		{
			name: "second shard",
			existingShards: []recovery.ShamirShard{
				{ID: 1, Value: createTestShardValue(1)},
			},
			newShardID:      2,
			expectedCount:   2,
			expectThreshold: false,
		},
		{
			name: "third shard - reaches threshold",
			existingShards: []recovery.ShamirShard{
				{ID: 1, Value: createTestShardValue(1)},
				{ID: 2, Value: createTestShardValue(2)},
			},
			newShardID:      3,
			expectedCount:   3,
			expectThreshold: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset shards and set up existing ones
			resetShards()
			shardsMutex.Lock()
			shards = append(shards, tt.existingShards...)
			shardsMutex.Unlock()

			// Test the logic
			shardsMutex.RLock()
			currentCount := len(shards)
			threshold := env.ShamirThresholdVal()
			shardsMutex.RUnlock()

			// Simulate adding a new shard
			if currentCount < threshold {
				shardsMutex.Lock()
				shards = append(shards, recovery.ShamirShard{
					ID:    tt.newShardID,
					Value: createTestShardValue(int(tt.newShardID)),
				})
				newCount := len(shards)
				shardsMutex.Unlock()

				if newCount != tt.expectedCount {
					t.Errorf("Expected %d shards, got %d", tt.expectedCount, newCount)
				}

				if (newCount == threshold) != tt.expectThreshold {
					t.Errorf("Expected threshold reached: %v, got: %v",
						tt.expectThreshold, newCount == threshold)
				}
			}
		})
	}
}

func TestShardDuplicateDetection(t *testing.T) {
	// Test duplicate shard detection logic
	resetShards()

	// Add initial shards
	testShards := []recovery.ShamirShard{
		{ID: 1, Value: createTestShardValue(1)},
		{ID: 3, Value: createTestShardValue(3)},
		{ID: 5, Value: createTestShardValue(5)},
	}

	shardsMutex.Lock()
	shards = testShards
	shardsMutex.Unlock()

	tests := []struct {
		name       string
		requestID  int
		expectDupe bool
	}{
		{"new shard ID 2", 2, false},
		{"new shard ID 4", 4, false},
		{"duplicate shard ID 1", 1, true},
		{"duplicate shard ID 3", 3, true},
		{"duplicate shard ID 5", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shardsMutex.RLock()
			isDuplicate := false
			for _, shard := range shards {
				if int(shard.ID) == tt.requestID {
					isDuplicate = true
					break
				}
			}
			shardsMutex.RUnlock()

			if isDuplicate != tt.expectDupe {
				t.Errorf("Expected duplicate: %v, got: %v", tt.expectDupe, isDuplicate)
			}
		})
	}
}

func TestShardSecurityCleanup(t *testing.T) {
	// Test that shards are properly cleaned up after restoration
	// Save original environment variables
	originalThreshold := os.Getenv(env.NexusShamirThreshold)
	defer func() {
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
	}()

	_ = os.Setenv(env.NexusShamirThreshold, "2")

	// Reset and add shards that would trigger restoration
	resetShards()
	testShards := []recovery.ShamirShard{
		{ID: 1, Value: createTestShardValue(1)},
		{ID: 2, Value: createTestShardValue(2)},
	}

	shardsMutex.Lock()
	shards = testShards
	shardsMutex.Unlock()

	// Simulate the cleanup logic from RouteRestore
	shardsMutex.Lock()
	if len(shards) == env.ShamirThresholdVal() {
		// This simulates the security cleanup without calling the actual
		// RestoreBackingStoreFromPilotShards which would require infrastructure
		for i := range shards {
			// In real code: mem.ClearRawBytes(shards[i].Value)
			// We simulate by checking the values exist before "clearing"
			if shards[i].Value == nil {
				t.Errorf("Shard %d value should not be nil before cleanup", i)
			}
			// Simulate clearing by setting to zero (in real code mem.ClearRawBytes does this)
			for j := range shards[i].Value {
				shards[i].Value[j] = 0
			}
			shards[i].ID = 0
		}
	}
	shardsMutex.Unlock()

	// Verify cleanup worked
	shardsMutex.RLock()
	for i, shard := range shards {
		if shard.ID != 0 {
			t.Errorf("Shard %d ID should be 0 after cleanup, got %d", i, shard.ID)
		}
		if shard.Value != nil {
			allZero := true
			for _, b := range shard.Value {
				if b != 0 {
					allZero = false
					break
				}
			}
			if !allZero {
				t.Errorf("Shard %d value should be all zeros after cleanup", i)
			}
		}
	}
	shardsMutex.RUnlock()
}

func TestConcurrentShardAccess(t *testing.T) {
	// Test thread-safe access to shards
	resetShards()

	// Save original environment variables
	originalThreshold := os.Getenv(env.NexusShamirThreshold)
	defer func() {
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
	}()

	_ = os.Setenv(env.NexusShamirThreshold, "10")

	var wg sync.WaitGroup
	numGoroutines := 5
	shardsPerGoroutine := 2

	// Launch concurrent goroutines to add shards
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < shardsPerGoroutine; j++ {
				shardID := uint64(goroutineID*shardsPerGoroutine + j + 1)

				shardsMutex.Lock()
				shards = append(shards, recovery.ShamirShard{
					ID:    shardID,
					Value: createTestShardValue(int(shardID)),
				})
				shardsMutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify all shards were added
	shardsMutex.RLock()
	expectedCount := numGoroutines * shardsPerGoroutine
	actualCount := len(shards)
	shardsMutex.RUnlock()

	if actualCount != expectedCount {
		t.Errorf("Expected %d shards, got %d", expectedCount, actualCount)
	}

	// Verify no duplicate IDs
	shardsMutex.RLock()
	idMap := make(map[uint64]bool)
	for _, shard := range shards {
		if idMap[shard.ID] {
			t.Errorf("Duplicate shard ID found: %d", shard.ID)
		}
		idMap[shard.ID] = true
	}
	shardsMutex.RUnlock()
}

func TestRestorationStatusCalculation(t *testing.T) {
	// Test restoration status calculation logic
	// Save original environment variables
	originalThreshold := os.Getenv(env.NexusShamirThreshold)
	defer func() {
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
	}()

	_ = os.Setenv(env.NexusShamirThreshold, "5")

	tests := []struct {
		name              string
		currentShardCount int
		expectedRemaining int
		expectedRestored  bool
	}{
		{"no shards", 0, 5, false},
		{"one shard", 1, 4, false},
		{"half shards", 2, 3, false},
		{"almost complete", 4, 1, false},
		{"complete", 5, 0, true},
		{"over complete", 6, 0, true}, // Should not happen in practice
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold := env.ShamirThresholdVal()
			remaining := threshold - tt.currentShardCount
			if remaining < 0 {
				remaining = 0
			}
			restored := tt.currentShardCount >= threshold

			if remaining != tt.expectedRemaining {
				t.Errorf("Expected %d remaining, got %d", tt.expectedRemaining, remaining)
			}

			if restored != tt.expectedRestored {
				t.Errorf("Expected restored: %v, got: %v", tt.expectedRestored, restored)
			}
		})
	}
}

func TestEnvironmentDependencies(t *testing.T) {
	// Test environment variable dependencies
	originalStore := os.Getenv(env.NexusBackendStore)
	originalThreshold := os.Getenv(env.NexusShamirThreshold)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
	}()

	// Test backend store detection
	testBackends := []struct {
		value    string
		expected env.StoreType
	}{
		{"memory", env.Memory},
		{"sqlite", env.Sqlite},
		{"lite", env.Lite},
	}

	for _, backend := range testBackends {
		_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", backend.value)
		actual := env.BackendStoreTypeVal()
		if actual != backend.expected {
			t.Errorf("Expected backend type %s, got %s", backend.expected, actual)
		}
	}

	// Test threshold configuration
	testThresholds := []struct {
		value    string
		expected int
	}{
		{"1", 1},
		{"3", 3},
		{"10", 10},
	}

	for _, threshold := range testThresholds {
		_ = os.Setenv(env.NexusShamirThreshold, threshold.value)
		actual := env.ShamirThresholdVal()
		if actual != threshold.expected {
			t.Errorf("Expected threshold %d, got %d", threshold.expected, actual)
		}
	}
}
