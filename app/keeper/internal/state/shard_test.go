//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestSetShardValidData(t *testing.T) {
	// Reset shard to zero state before test
	resetShard()

	// Create test data
	testData := [crypto.AES256KeySize]byte{}
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	// Set the shard
	SetShard(&testData)

	// Verify it was set correctly
	RLockShard()
	retrievedShard := ShardNoSync()
	if !bytes.Equal(retrievedShard[:], testData[:]) {
		t.Errorf("Shard data mismatch. Expected: %v, Got: %v", testData, *retrievedShard)
	}
	RUnlockShard()
}

func TestSetShardZeroData(t *testing.T) {
	// Reset shard to a known non-zero state
	resetShard()
	nonZeroData := [crypto.AES256KeySize]byte{}
	for i := range nonZeroData {
		nonZeroData[i] = byte(i + 1)
	}
	SetShard(&nonZeroData)

	// Verify initial state is non-zero
	RLockShard()
	initialShard := *ShardNoSync()
	RUnlockShard()

	// Try to set zero data
	zeroData := [crypto.AES256KeySize]byte{} // All zeros
	SetShard(&zeroData)

	// Verify shard was NOT changed (should still be non-zero)
	RLockShard()
	currentShard := *ShardNoSync()
	RUnlockShard()

	if !bytes.Equal(initialShard[:], currentShard[:]) {
		t.Error("Shard should not be updated when setting zero data")
	}

	// Verify it's still the original non-zero data
	if bytes.Equal(currentShard[:], zeroData[:]) {
		t.Error("Shard should not be zero after attempting to set zero data")
	}
}

func TestSetShardPartialZeroData(t *testing.T) {
	// Reset shard
	resetShard()

	// Create data that has some zero bytes but is not all zeros
	testData := [crypto.AES256KeySize]byte{}
	testData[0] = 0   // The first byte is zero
	testData[1] = 0   // The second byte is zero
	testData[31] = 42 // The last byte is non-zero

	// This should be accepted since it's not all zeros
	SetShard(&testData)

	// Verify it was set correctly
	RLockShard()
	retrievedShard := ShardNoSync()
	if !bytes.Equal(retrievedShard[:], testData[:]) {
		t.Error("Shard with partial zero data should be accepted")
	}
	RUnlockShard()
}

func TestShardNoSyncReturnsPointer(t *testing.T) {
	// Reset shard
	resetShard()

	// Set test data
	testData := [crypto.AES256KeySize]byte{}
	for i := range testData {
		testData[i] = byte(i)
	}
	SetShard(&testData)

	// Get a pointer without sync
	shardPtr := ShardNoSync()

	// Verify it's the same data
	if !bytes.Equal(shardPtr[:], testData[:]) {
		t.Error("ShardNoSync should return pointer to current shard data")
	}

	// Verify it's actually a pointer to the internal shard
	// (changing the returned value should change the internal state)
	originalByte := shardPtr[0]
	shardPtr[0] = originalByte + 1

	// Get another pointer and check if the change is reflected
	shardPtr2 := ShardNoSync()
	if shardPtr2[0] != originalByte+1 {
		t.Error("ShardNoSync should return pointer to the actual internal shard")
	}

	// Restore original value
	shardPtr[0] = originalByte
}

func TestLockUnlockFunctions(t *testing.T) {
	// Test that RLockShard and RUnlockShard work without panicking
	RLockShard()
	RUnlockShard()

	// Test multiple read locks (should be allowed)
	RLockShard()
	RLockShard()
	RUnlockShard()
	RUnlockShard()
}

func TestConcurrentSetShard(t *testing.T) {
	// Reset shard
	resetShard()

	const numGoroutines = 100
	const numOperations = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines that set different shard values
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				testData := [crypto.AES256KeySize]byte{}
				// Create a unique pattern for each goroutine
				for k := range testData {
					testData[k] = byte((id + j + k) % 256)
				}
				// Ensure it's not all zeros
				testData[0] = byte(id + 1)

				SetShard(&testData)

				// Small delay to increase the chance of race conditions
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify that the shard has some valid (non-zero) data
	RLockShard()
	finalShard := *ShardNoSync()
	RUnlockShard()

	allZero := true
	for _, b := range finalShard {
		if b != 0 {
			allZero = false
			break
		}
	}

	if allZero {
		t.Error("After concurrent operations, shard should not be all zeros")
	}
}

func TestConcurrentReadWrite(t *testing.T) {
	// Reset shard
	resetShard()

	const numReaders = 50
	const numWriters = 10
	const duration = 100 * time.Millisecond

	var wg sync.WaitGroup
	stopCh := make(chan bool)

	// Start readers
	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					RLockShard()
					shard := ShardNoSync()
					_ = shard[0] // Read first byte
					RUnlockShard()
				}
			}
		}(i)
	}

	// Start writers
	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					testData := [crypto.AES256KeySize]byte{}
					for j := range testData {
						testData[j] = byte((id + j) % 256)
					}
					testData[0] = byte(id + 1) // Ensure non-zero
					SetShard(&testData)
				}
			}
		}(i)
	}

	// Let them run for a bit
	time.Sleep(duration)
	close(stopCh)
	wg.Wait()

	// Test should complete without deadlocks or race conditions
	t.Log("Concurrent read/write test completed successfully")
}

func TestShardIsolation(t *testing.T) {
	// Reset shard
	resetShard()

	// Set initial data
	initialData := [crypto.AES256KeySize]byte{}
	for i := range initialData {
		initialData[i] = byte(i)
	}
	SetShard(&initialData)

	// Get a copy of the shard
	RLockShard()
	shardCopy := *ShardNoSync()
	RUnlockShard()

	// Modify the copy
	shardCopy[0] = 255

	// Verify the original shard is unchanged
	RLockShard()
	originalShard := *ShardNoSync()
	RUnlockShard()

	if originalShard[0] == 255 {
		t.Error("Modifying a copy should not affect the original shard")
	}

	if originalShard[0] != initialData[0] {
		t.Error("Original shard data should remain unchanged")
	}
}

func TestZeroDetection(t *testing.T) {
	tests := []struct {
		name   string
		data   [crypto.AES256KeySize]byte
		isZero bool
	}{
		{
			name:   "all zeros",
			data:   [crypto.AES256KeySize]byte{}, // Zero value
			isZero: true,
		},
		{
			name: "first byte non-zero",
			data: func() [crypto.AES256KeySize]byte {
				var d [crypto.AES256KeySize]byte
				d[0] = 1
				return d
			}(),
			isZero: false,
		},
		{
			name: "last byte non-zero",
			data: func() [crypto.AES256KeySize]byte {
				var d [crypto.AES256KeySize]byte
				d[crypto.AES256KeySize-1] = 1
				return d
			}(),
			isZero: false,
		},
		{
			name: "middle byte non-zero",
			data: func() [crypto.AES256KeySize]byte {
				var d [crypto.AES256KeySize]byte
				d[crypto.AES256KeySize/2] = 1
				return d
			}(),
			isZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset shard to known state
			resetShard()
			nonZeroInitial := [crypto.AES256KeySize]byte{}
			nonZeroInitial[0] = 42
			SetShard(&nonZeroInitial)

			// Get initial state
			RLockShard()
			beforeShard := *ShardNoSync()
			RUnlockShard()

			// Try to set the test data
			SetShard(&tt.data)

			// Get state after SetShard
			RLockShard()
			afterShard := *ShardNoSync()
			RUnlockShard()

			if tt.isZero {
				// Should not have changed
				if !bytes.Equal(beforeShard[:], afterShard[:]) {
					t.Error("Shard should not change when setting zero data")
				}
			} else {
				// Should have changed to the new data
				if !bytes.Equal(tt.data[:], afterShard[:]) {
					t.Error("Shard should change when setting non-zero data")
				}
			}
		})
	}
}

func TestShardSize(t *testing.T) {
	// Verify the shard has the correct size
	RLockShard()
	shardPtr := ShardNoSync()
	RUnlockShard()

	fmt.Println(*(shardPtr))

	// noinspection GoBoolExpressions
	if len(shardPtr) != crypto.AES256KeySize {
		t.Errorf("Shard size should be %d, got %d", crypto.AES256KeySize, len(shardPtr))
	}

	// noinspection GoBoolExpressions
	if crypto.AES256KeySize != 32 {
		t.Errorf("Expected AES256KeySize to be 32, got %d", crypto.AES256KeySize)
	}
}
