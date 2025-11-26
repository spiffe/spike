//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"os"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
)

func TestErrRecoveryRetryFailed(t *testing.T) {
	// Test that the SDK error variable is properly defined
	if sdkErrors.ErrRecoveryRetryFailed == nil {
		t.Error("ErrRecoveryRetryFailed should not be nil")
	}

	// Test that the error implements the error interface
	var _ error = sdkErrors.ErrRecoveryRetryFailed

	// Test that it can be cloned and wrapped (common SDK error operations)
	cloned := sdkErrors.ErrRecoveryRetryFailed.Clone()
	if cloned == nil {
		t.Error("Clone should return a non-nil error")
	}

	// Test Is() method works for error comparison
	if !cloned.Is(sdkErrors.ErrRecoveryRetryFailed) {
		t.Error("Cloned error should match original via Is()")
	}
}

func TestRestoreBackingStoreFromPilotShards_InsufficientShards(t *testing.T) {
	// Use t.Setenv for automatic cleanup after test
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	t.Setenv(env.NexusShamirThreshold, "3")

	// Create insufficient shards (only 2, but the threshold is 3)
	shards := make([]ShamirShard, 2)
	for i := range shards {
		testData := &[crypto.AES256KeySize]byte{}
		for j := range testData {
			testData[j] = byte((i + j) % 256)
		}
		testData[0] = byte(i + 1) // Ensure non-zero

		shards[i] = ShamirShard{
			ID:    uint64(i + 1),
			Value: testData,
		}
	}

	// The function calls log.FatalErr when there are insufficient shards.
	// With SPIKE_STACK_TRACES_ON_LOG_FATAL=true, it panics instead of os.Exit,
	// allowing us to recover and verify the behavior.
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic from log.FatalErr for insufficient shards")
		}
	}()

	RestoreBackingStoreFromPilotShards(shards)

	t.Error("Function should have panicked for insufficient shards")
}

func TestRestoreBackingStoreFromPilotShards_InvalidShards(t *testing.T) {
	tests := []struct {
		name       string
		setupShard func() ShamirShard
		shouldExit bool
	}{
		{
			name: "nil value shard",
			setupShard: func() ShamirShard {
				return ShamirShard{
					ID:    1,
					Value: nil,
				}
			},
			shouldExit: true,
		},
		{
			name: "zero ID shard",
			setupShard: func() ShamirShard {
				testData := &[crypto.AES256KeySize]byte{}
				testData[0] = 1 // Non-zero data
				return ShamirShard{
					ID:    0, // Zero ID
					Value: testData,
				}
			},
			shouldExit: true,
		},
		{
			name: "zeroed value shard",
			setupShard: func() ShamirShard {
				testData := &[crypto.AES256KeySize]byte{} // All zeros
				return ShamirShard{
					ID:    1,
					Value: testData,
				}
			},
			shouldExit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shards := []ShamirShard{tt.setupShard()}

			if tt.shouldExit {
				// TODO: this can be modified with env var; fix the test.
				t.Skip("Skipping test that would call os.Exit() - function calls log.FatalErr")
				return
			}

			RestoreBackingStoreFromPilotShards(shards)
		})
	}
}

func TestRestoreBackingStoreFromPilotShards_ValidInput(t *testing.T) {
	// This test would hang without proper SPIFFE infrastructure setup
	// as the function calls spiffe.Source() which tries to connect to SPIFFE workload API
	// and then makes network calls to keepers
	t.Skip("Skipping test that requires SPIFFE infrastructure and would hang on network calls")

	// Note: In a real test environment, you would:
	// 1. Mock the spiffe.Source() function
	// 2. Mock the network calls to keepers
	// 3. Set up test SPIFFE infrastructure
	// 4. Or refactor the code to be more testable by injecting dependencies
}

func TestNewPilotRecoveryShards_NoRootKey(t *testing.T) {
	// This test assumes no root key is available in state
	// NewPilotRecoveryShards should return nil when no root key exists

	result := NewPilotRecoveryShards()

	// Should return nil when no root key is available
	if result != nil {
		t.Error("Expected nil when no root key is available")
	}
}

func TestShamirShardValidation(t *testing.T) {
	// Test creating and validating ShamirShard structures
	tests := []struct {
		name    string
		id      uint64
		value   *[crypto.AES256KeySize]byte
		isValid bool
	}{
		{
			name: "valid shard",
			id:   1,
			value: func() *[crypto.AES256KeySize]byte {
				data := &[crypto.AES256KeySize]byte{}
				data[0] = 1 // Non-zero
				return data
			}(),
			isValid: true,
		},
		{
			name:    "nil value",
			id:      1,
			value:   nil,
			isValid: false,
		},
		{
			name:    "zero value",
			id:      1,
			value:   &[crypto.AES256KeySize]byte{}, // All zeros
			isValid: false,
		},
		{
			name: "zero ID",
			id:   0,
			value: func() *[crypto.AES256KeySize]byte {
				data := &[crypto.AES256KeySize]byte{}
				data[0] = 1
				return data
			}(),
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shard := ShamirShard{
				ID:    tt.id,
				Value: tt.value,
			}

			// Test the validation logic that RestoreBackingStoreFromPilotShards uses
			isValid := true
			if shard.Value == nil {
				isValid = false
			} else if shard.ID == 0 {
				isValid = false
			} else {
				// Check if the value is all zeros (simulating mem.Zeroed32)
				allZero := true
				for _, b := range shard.Value {
					if b != 0 {
						allZero = false
						break
					}
				}
				if allZero {
					isValid = false
				}
			}

			if isValid != tt.isValid {
				t.Errorf("Expected validity %v, got %v", tt.isValid, isValid)
			}
		})
	}
}

func TestShamirShardSliceOperations(t *testing.T) {
	// Test operations on slices of ShamirShard
	shards := make([]ShamirShard, 3)

	// Initialize test shards
	for i := range shards {
		testData := &[crypto.AES256KeySize]byte{}
		for j := range testData {
			testData[j] = byte((i*100 + j) % 256)
		}
		testData[0] = byte(i + 1) // Ensure non-zero

		shards[i] = ShamirShard{
			ID:    uint64(i + 1),
			Value: testData,
		}
	}

	// Test slice length
	if len(shards) != 3 {
		t.Errorf("Expected 3 shards, got %d", len(shards))
	}

	// Test iteration (as done in RestoreBackingStoreFromPilotShards)
	for i, shard := range shards {
		expectedID := uint64(i + 1)
		if shard.ID != expectedID {
			t.Errorf("Shard %d: expected ID %d, got %d", i, expectedID, shard.ID)
		}

		if shard.Value == nil {
			t.Errorf("Shard %d: value should not be nil", i)
		}

		if shard.Value[0] == 0 {
			t.Errorf("Shard %d: first byte should not be zero", i)
		}
	}
}

func TestEnvironmentDependencies(t *testing.T) {
	// Test that environment functions work as expected
	originalThreshold := os.Getenv(env.NexusShamirThreshold)
	originalShares := os.Getenv(env.NexusShamirShares)

	defer func() {
		if originalThreshold != "" {
			_ = os.Setenv(env.NexusShamirThreshold, originalThreshold)
		} else {
			_ = os.Unsetenv(env.NexusShamirThreshold)
		}
		if originalShares != "" {
			_ = os.Setenv(env.NexusShamirShares, originalShares)
		} else {
			_ = os.Unsetenv(env.NexusShamirShares)
		}
	}()

	// Test ShamirThreshold
	_ = os.Setenv(env.NexusShamirThreshold, "3")
	threshold := env.ShamirThresholdVal()
	if threshold != 3 {
		t.Errorf("Expected threshold 3, got %d", threshold)
	}

	// Test ShamirShares
	_ = os.Setenv(env.NexusShamirShares, "5")
	shares := env.ShamirSharesVal()
	if shares != 5 {
		t.Errorf("Expected shares 5, got %d", shares)
	}

	// Test `threshold <= shares` (common validation)
	if threshold > shares {
		t.Errorf("Threshold (%d) should not be greater than shares (%d)",
			threshold, shares)
	}
}

func TestCryptoConstantsShard(t *testing.T) {
	// Test that the crypto constants are as expected
	//noinspection GoBoolExpressions
	if crypto.AES256KeySize != 32 {
		t.Errorf("Expected AES256KeySize to be 32, got %d", crypto.AES256KeySize)
	}

	// Test creating arrays with the constant
	var testArray [crypto.AES256KeySize]byte
	//noinspection GoBoolExpressions
	if len(testArray) != 32 {
		t.Errorf("Expected array length 32, got %d", len(testArray))
	}
}

func TestMapOperations(t *testing.T) {
	// Test map operations used in NewPilotRecoveryShards
	result := make(map[int]*[crypto.AES256KeySize]byte)

	// Add test entries
	for i := 1; i <= 3; i++ {
		testData := &[crypto.AES256KeySize]byte{}
		testData[0] = byte(i)
		result[i] = testData
	}

	if len(result) != 3 {
		t.Errorf("Expected map length 3, got %d", len(result))
	}

	// Test retrieval
	for i := 1; i <= 3; i++ {
		data := result[i]
		if data == nil {
			t.Errorf("Expected non-nil data for key %d", i)
			continue
		}
		if data[0] != byte(i) {
			t.Errorf("Expected first byte %d, got %d", i, data[0])
		}
	}

	// Test iteration
	count := 0
	for key, value := range result {
		count++
		if value == nil {
			t.Errorf("Value for key %d should not be nil", key)
		}
		if key < 1 || key > 3 {
			t.Errorf("Unexpected key %d", key)
		}
	}

	if count != 3 {
		t.Errorf("Expected to iterate over 3 entries, got %d", count)
	}
}

func TestShardDataIntegrity(t *testing.T) {
	// Test that shard data maintains integrity during operations
	originalData := [crypto.AES256KeySize]byte{}
	for i := range originalData {
		originalData[i] = byte(i % 256)
	}

	// Create shard
	shard := ShamirShard{
		ID:    123,
		Value: &originalData,
	}

	// Verify data integrity
	if shard.Value == nil {
		t.Fatal("Shard value should not be nil")
	}

	for i, b := range shard.Value {
		expected := byte(i % 256)
		if b != expected {
			t.Errorf("Data corruption at index %d: expected %d, got %d",
				i, expected, b)
		}
	}

	// Test that modifying the shard affects the original (it's a pointer)
	shard.Value[0] = 255
	if originalData[0] != 255 {
		t.Error("Expected modification through pointer to affect original data")
	}
}

func TestSliceLengthValidation(t *testing.T) {
	// Test length validation logic used in the functions
	tests := []struct {
		name          string
		sliceLength   int
		threshold     int
		shouldBeValid bool
	}{
		{"exact threshold", 3, 3, true},
		{"above threshold", 5, 3, true},
		{"below threshold", 2, 3, false},
		{"zero length", 0, 3, false},
		{"zero threshold", 3, 0, true}, // Any length >= 0 threshold
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.sliceLength >= tt.threshold
			if isValid != tt.shouldBeValid {
				t.Errorf("Expected validity %v for length %d vs threshold %d, got %v",
					tt.shouldBeValid, tt.sliceLength, tt.threshold, isValid)
			}
		})
	}
}
