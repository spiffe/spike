//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"os"
	"strconv"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestIterateKeepersAndInitializeState_MemoryMode(t *testing.T) {
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
	_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")

	// Verify the environment is set correctly
	if env.BackendStoreTypeVal() != env.Memory {
		t.Fatal("Expected memory backend store type")
	}

	// Test with nil source and empty shards map
	successfulKeeperShards := make(map[string]*[crypto.AES256KeySize]byte)
	result := iterateKeepersAndInitializeState(nil, successfulKeeperShards)

	// Should return true and skip recovery in memory mode
	if !result {
		t.Error("Expected true in memory mode (should skip recovery)")
	}

	// Shards map should remain empty in memory mode
	if len(successfulKeeperShards) != 0 {
		t.Error("Shards map should remain empty in memory mode")
	}
}

func TestIterateKeepersAndInitializeState_NonMemoryMode(t *testing.T) {
	// Save original environment variables
	originalStore := os.Getenv(env.NexusBackendStore)
	originalKeeperPeers := os.Getenv(env.NexusKeeperPeers)
	originalThreshold := os.Getenv(env.NexusShamirThreshold)

	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
		if originalKeeperPeers != "" {
			_ = os.Setenv(env.NexusKeeperPeers, originalKeeperPeers)
		} else {
			_ = os.Unsetenv(env.NexusKeeperPeers)
		}
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
	}()

	// Set to non-memory mode
	_ = os.Setenv(env.NexusBackendStore, "sqlite")
	_ = os.Setenv(env.NexusKeeperPeers,
		"https://keeper1.example.com,https://keeper2.example.com") // Test keepers
	_ = os.Setenv("SPIKE_NEXUS_SHAMIR_THRESHOLD", "2")

	// Verify the environment is set correctly
	if env.BackendStoreTypeVal() == env.Memory {
		t.Fatal("Expected non-memory backend store type")
	}

	// Test with configured keepers but nil source (will fail network calls)
	successfulKeeperShards := make(map[string]*[crypto.AES256KeySize]byte)
	result := iterateKeepersAndInitializeState(nil, successfulKeeperShards)

	// Should return false when network calls fail due to the nil source
	if result {
		t.Error("Expected false when network calls fail due to nil source/unreachable keepers")
	}
}

func TestIterateKeepersAndInitializeState_ShardMapHandling(t *testing.T) {
	// Test that the function properly handles the successfulKeeperShards map

	// Create a test map with some existing data
	successfulKeeperShards := make(map[string]*[crypto.AES256KeySize]byte)

	// Add some test data
	testShard := &[crypto.AES256KeySize]byte{}
	for i := range testShard {
		testShard[i] = byte(i % 256)
	}
	successfulKeeperShards["test-keeper"] = testShard

	// Save the original environment
	originalStore := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Set to memory mode for a quick test
	_ = os.Setenv(env.NexusBackendStore, "memory")

	// The map should not be modified in memory mode
	originalLen := len(successfulKeeperShards)
	result := iterateKeepersAndInitializeState(nil, successfulKeeperShards)

	if !result {
		t.Error("Expected true in memory mode")
	}

	if len(successfulKeeperShards) != originalLen {
		t.Error("Shard map should not be modified in memory mode")
	}

	// Verify the original data is still there
	if successfulKeeperShards["test-keeper"] == nil {
		t.Error("Original shard data should still be present")
	}
}

func TestIterateKeepersAndInitializeState_ParameterValidation(t *testing.T) {
	// Save the original environment
	originalStore := os.Getenv(env.NexusBackendStore)
	originalKeeperPeers := os.Getenv(env.NexusKeeperPeers)
	defer func() {
		if originalStore != "" {
			_ = os.Setenv(env.NexusBackendStore, originalStore)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
		if originalKeeperPeers != "" {
			_ = os.Setenv(env.NexusKeeperPeers, originalKeeperPeers)
		} else {
			_ = os.Unsetenv(env.NexusKeeperPeers)
		}
	}()

	// Set to memory mode for testing
	_ = os.Setenv(env.NexusBackendStore, "memory")

	// Test with a nil source (should work in memory mode)
	result := iterateKeepersAndInitializeState(nil, make(map[string]*[crypto.AES256KeySize]byte))
	if !result {
		t.Error("Should work with nil source in memory mode")
	}

	_ = os.Setenv(env.NexusBackendStore, "sqlite")
	_ = os.Setenv(env.NexusKeeperPeers, "https://localhost:8443")

	// Test with a nil shards map (should not panic, just return false)
	// The function gracefully handles nil maps by never reaching code that would panic
	// since network calls will fail and no shards will be processed
	result2 := iterateKeepersAndInitializeState(nil, nil)
	if result2 {
		t.Error("Expected false when using nil shards map with unreachable keepers")
	}
}

func TestShamirShardStructure(t *testing.T) {
	// Test the ShamirShard structure used in the function
	testData := &[crypto.AES256KeySize]byte{}
	for i := range testData {
		testData[i] = byte(i)
	}

	shard := ShamirShard{
		ID:    123,
		Value: testData,
	}

	if shard.ID != 123 {
		t.Errorf("Expected ID 123, got %d", shard.ID)
	}

	if shard.Value == nil {
		t.Fatal("Shard value should not be nil")
	}

	// noinspection GoBoolExpressions
	if len(shard.Value) != crypto.AES256KeySize {
		t.Errorf("Expected shard value size %d, got %d", crypto.AES256KeySize, len(shard.Value))
	}

	// Verify data integrity
	for i, b := range shard.Value {
		if b != byte(i) {
			t.Errorf("Data mismatch at index %d: expected %d, got %d", i, byte(i), b)
		}
	}
}

func TestShamirShardSliceHandling(t *testing.T) {
	// Test creating and manipulating slices of ShamirShard (as done in the function)
	shards := make([]ShamirShard, 0)

	// Add some test shards
	for i := 0; i < 3; i++ {
		testData := &[crypto.AES256KeySize]byte{}
		for j := range testData {
			testData[j] = byte((i + j) % 256)
		}

		shard := ShamirShard{
			ID:    uint64(i + 1),
			Value: testData,
		}
		shards = append(shards, shard)
	}

	if len(shards) != 3 {
		t.Errorf("Expected 3 shards, got %d", len(shards))
	}

	// Test that each shard has unique ID and data
	for i, shard := range shards {
		expectedID := uint64(i + 1)
		if shard.ID != expectedID {
			t.Errorf("Shard %d: expected ID %d, got %d", i, expectedID, shard.ID)
		}

		if shard.Value == nil {
			t.Errorf("Shard %d: value should not be nil", i)
			continue
		}

		// Check the first byte of data (should be unique per shard)
		expectedFirstByte := byte(i)
		if shard.Value[0] != expectedFirstByte {
			t.Errorf("Shard %d: expected first byte %d, got %d", i, expectedFirstByte, shard.Value[0])
		}
	}
}

func TestKeeperIDConversion(t *testing.T) {
	// Test the string to integer conversion logic used in the function
	tests := []struct {
		name      string
		keeperID  string
		expectErr bool
		expected  int
	}{
		{"valid numeric ID", "123", false, 123},
		{"single digit", "5", false, 5},
		{"zero", "0", false, 0},
		{"large number", "999999", false, 999999},
		{"invalid non-numeric", "abc", true, 0},
		{"empty string", "", true, 0},
		{"mixed alphanumeric", "123abc", true, 0},
		// {"negative number", "-1", true, 0}, // strconv.Atoi would handle this, but depends on requirements
		// FIX-ME: keeper ids need to be stricter. add validation logic to the code.
		// Maybe a better sanitization before the code even gets there.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This simulates the conversion done in iterateKeepersAndInitializeState
			result, err := convertKeeperID(tt.keeperID)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error for keeper ID '%s', but got none", tt.keeperID)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for keeper ID '%s': %v", tt.keeperID, err)
				}
				if result != tt.expected {
					t.Errorf("Expected result %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestMemoryConstants(t *testing.T) {
	// Test that the crypto constants used are correct

	// noinspection GoBoolExpressions
	if crypto.AES256KeySize != 32 {
		t.Errorf("Expected AES256KeySize to be 32, got %d", crypto.AES256KeySize)
	}

	// Test creating arrays with the constant
	var testArray [crypto.AES256KeySize]byte

	// noinspection GoBoolExpressions
	if len(testArray) != 32 {
		t.Errorf("Expected array length 32, got %d", len(testArray))
	}
}

func TestShardMapOperations(t *testing.T) {
	// Test operations on the shard map type used in the function
	shardMap := make(map[string]*[crypto.AES256KeySize]byte)

	// Test adding shards
	testShard1 := &[crypto.AES256KeySize]byte{}
	testShard1[0] = 1
	shardMap["keeper1"] = testShard1

	testShard2 := &[crypto.AES256KeySize]byte{}
	testShard2[0] = 2
	shardMap["keeper2"] = testShard2

	if len(shardMap) != 2 {
		t.Errorf("Expected 2 shards in map, got %d", len(shardMap))
	}

	// Test retrieving shards
	retrievedShard1, ok := shardMap["keeper1"]
	if !ok || retrievedShard1 == nil {
		t.Fatal("Expected to retrieve shard for keeper1")
	}
	if retrievedShard1[0] != 1 {
		t.Errorf("Expected first byte to be 1, got %d", retrievedShard1[0])
	}

	// Test iterating over a map (as done in the function)
	count := 0
	for keeperID, shard := range shardMap {
		count++
		if shard == nil {
			t.Errorf("Shard for keeper %s should not be nil", keeperID)
		}
		if len(keeperID) == 0 {
			t.Error("Keeper ID should not be empty")
		}
	}

	if count != 2 {
		t.Errorf("Expected to iterate over 2 shards, got %d", count)
	}
}

// Helper function to test keeper ID conversion (mimics the logic in the main function)
func convertKeeperID(keeperID string) (int, error) {
	// This mimics: id, err := strconv.Atoi(ix)
	// from the iterateKeepersAndInitializeState function
	return strconv.Atoi(keeperID)
}

func TestEnvironmentDependenciesKeeper(t *testing.T) {
	// Test that environment functions work as expected

	// Save original values
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

	// Test BackendStoreType detection
	_ = os.Setenv(env.NexusBackendStore, "memory")
	if env.BackendStoreTypeVal() != env.Memory {
		t.Error("Expected Memory backend type")
	}

	_ = os.Setenv(env.NexusBackendStore, "sqlite")
	if env.BackendStoreTypeVal() != env.Sqlite {
		t.Error("Expected Sqlite backend type")
	}

	// Test ShamirThreshold function exists and returns a reasonable value
	_ = os.Setenv(env.NexusShamirThreshold, "3")
	threshold := env.ShamirThresholdVal()
	if threshold < 1 {
		t.Errorf("Expected threshold >= 1, got %d", threshold)
	}
}
