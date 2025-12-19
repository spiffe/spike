//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"sync"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/security/mem"
)

func TestRootKeyZero_InitiallyZero(t *testing.T) {
	// Reset to ensure a clean state
	resetRootKey()

	// Check initial state - should be zero
	if !RootKeyZero() {
		t.Error("Root key should initially be zero")
	}
}

func TestRootKeyZero_NonZeroKey(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Set a non-zero key
	testKey := createTestKey(t)
	setRootKeyDirect(testKey)

	// Check that it's not zero
	if RootKeyZero() {
		t.Error("Root key should not be zero after setting non-zero key")
	}

	// Clean up
	resetRootKey()
}

func TestRootKeyZero_PartiallyZeroKey(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Create key with only first byte non-zero
	testKey := &[crypto.AES256KeySize]byte{}
	testKey[0] = 1 // Only first byte non-zero
	setRootKeyDirect(testKey)

	// Check that it's not considered zero
	if RootKeyZero() {
		t.Error("Root key should not be zero if any byte is non-zero")
	}

	// Clean up
	resetRootKey()
}

func TestRootKeyZero_LastByteNonZero(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Create key with only last byte non-zero
	testKey := &[crypto.AES256KeySize]byte{}
	testKey[crypto.AES256KeySize-1] = 0xFF // Only last byte non-zero
	setRootKeyDirect(testKey)

	// Check that it's not considered zero
	if RootKeyZero() {
		t.Error("Root key should not be zero if last byte is non-zero")
	}

	// Clean up
	resetRootKey()
}

func TestSetRootKey_ValidKey(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Create a valid test key
	testKey := createTestKey(t)

	// Set the key
	SetRootKey(testKey)

	// Verify the key was set
	if RootKeyZero() {
		t.Error("Root key should not be zero after setting valid key")
	}

	// Verify the actual key value
	rootKeyMu.RLock()
	for i, b := range rootKey {
		if b != testKey[i] {
			t.Errorf("Root key byte %d mismatch: expected %d, got %d",
				i, testKey[i], b)
		}
	}
	rootKeyMu.RUnlock()

	// Clean up
	resetRootKey()
}

func TestSetRootKey_NilKey(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Enable panic mode for log.FatalErr so we can test error handling
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")

	defer func() {
		if r := recover(); r == nil {
			t.Error("SetRootKey(nil) should panic with FatalErr")
		}
		// Root key should remain zero after the panic
		if !RootKeyZero() {
			t.Error("Root key should remain zero when setting nil key")
		}
	}()

	// Attempt to set nil key - this should cause log.FatalErr to panic
	SetRootKey(nil)

	t.Error("Should have panicked before reaching this point")
}

func TestSetRootKey_ZeroKey(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Enable panic mode for log.FatalErr so we can test error handling
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")

	defer func() {
		if r := recover(); r == nil {
			t.Error("SetRootKey(zeroKey) should panic with FatalErr")
		}
		// Root key should remain zero after the panic
		if !RootKeyZero() {
			t.Error("Root key should remain zero when setting zero key")
		}
	}()

	// Create a zero key
	zeroKey := &[crypto.AES256KeySize]byte{} // All zeros

	// Attempt to set a zero key - this should cause log.FatalErr to panic
	SetRootKey(zeroKey)

	t.Error("Should have panicked before reaching this point")
}

func TestSetRootKey_OverwriteExistingKey(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Set the initial key
	firstKey := createPatternKey(0xAA)
	SetRootKey(firstKey)

	// Verify the first key was set
	if RootKeyZero() {
		t.Fatal("First key should have been set")
	}

	// Set the second key
	secondKey := createPatternKey(0x55)
	SetRootKey(secondKey)

	// Verify second key overwrote first key
	rootKeyMu.RLock()
	for i, b := range rootKey {
		if b != 0x55 {
			t.Errorf("Root key byte %d should be 0x55, got 0x%02X", i, b)
		}
	}
	rootKeyMu.RUnlock()

	// Clean up
	resetRootKey()
}

func TestSetRootKey_KeyIndependence(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Create the test key
	testKey := createPatternKey(0x42)
	originalValue := testKey[0]

	// Set the key
	SetRootKey(testKey)

	// Modify the original key after setting
	testKey[0] = 0x99

	// Verify the internal root key wasn't affected
	rootKeyMu.RLock()
	if rootKey[0] != originalValue {
		t.Errorf("Root key should not be affected by changes to source key: expected 0x%02X, got 0x%02X",
			originalValue, rootKey[0])
	}
	rootKeyMu.RUnlock()

	// Clean up
	resetRootKey()
}

func TestRootKeyNoLock_ReturnsPointer(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Set a test key
	testKey := createPatternKey(0x33)
	setRootKeyDirect(testKey)

	// Get the pointer without a lock
	ptr := RootKeyNoLock()

	// Verify the pointer is not nil
	if ptr == nil {
		t.Fatal("RootKeyNoLock should not return nil")
		return
	}

	// Verify it points to correct data
	if (*ptr)[0] != 0x33 {
		t.Errorf("RootKeyNoLock should return pointer to current key: expected 0x33, got 0x%02X", (*ptr)[0])
	}

	// Verify it's the actual root key (modifying through pointer affects global)
	originalValue := (*ptr)[1]
	(*ptr)[1] = 0xDD

	rootKeyMu.RLock()
	if rootKey[1] != 0xDD {
		t.Error("Modifying through RootKeyNoLock pointer should affect global root key")
	}
	rootKeyMu.RUnlock()

	// Restore original value
	(*ptr)[1] = originalValue

	// Clean up
	resetRootKey()
}

func TestLockUnlockRootKey_BasicOperation(t *testing.T) {
	// Test that lock/unlock operations work
	LockRootKey()
	UnlockRootKey()

	// Test multiple lock/unlock cycles
	for i := 0; i < 5; i++ {
		LockRootKey()
		UnlockRootKey()
	}
}

func TestConcurrentRootKeyAccess(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Test concurrent reads using RootKeyZero
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				RootKeyZero()
			}
		}()
	}
	wg.Wait()

	// Test concurrent writes using SetRootKey
	// Note: Use goroutineID+1 to avoid creating zero-pattern keys which would
	// trigger FatalErr (zero keys are rejected for security reasons)
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < 10; j++ { // Fewer operations for writes
				testKey := createPatternKey(byte(goroutineID + 1))
				SetRootKey(testKey)
			}
		}(i)
	}
	wg.Wait()

	// Verify the system is still in a consistent state
	isZero := RootKeyZero()
	t.Logf("After concurrent operations, root key is zero: %v", isZero)

	// Clean up
	resetRootKey()
}

func TestConcurrentLockOperations(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 10

	// Test concurrent lock/unlock operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				LockRootKey()
				// Access the root key while holding the lock
				_ = RootKeyNoLock()
				UnlockRootKey()
			}
		}()
	}
	wg.Wait()
}

func TestMixedConcurrentOperations(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	var wg sync.WaitGroup

	// Goroutine performing reads
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			RootKeyZero()
			time.Sleep(1 * time.Microsecond)
		}
	}()

	// Goroutine performing writes
	// Note: Use (i%255)+1 to avoid creating zero-pattern keys which would
	// trigger FatalErr (zero keys are rejected for security reasons)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			testKey := createPatternKey(byte((i % 255) + 1))
			SetRootKey(testKey)
			time.Sleep(5 * time.Microsecond)
		}
	}()

	// Goroutine performing manual lock operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			LockRootKey()
			ptr := RootKeyNoLock()
			_ = ptr
			UnlockRootKey()
			time.Sleep(2 * time.Microsecond)
		}
	}()

	wg.Wait()

	// Verify the system is still functional
	finalIsZero := RootKeyZero()
	t.Logf("Final root key state - is zero: %v", finalIsZero)

	// Clean up
	resetRootKey()
}

func TestRootKeyStateTransitions(t *testing.T) {
	// Reset to a clean state
	resetRootKey()
	defer resetRootKey()

	// The initial state should be zero
	if !RootKeyZero() {
		t.Error("Initial state should be zero")
	}

	// Set a valid key - should not be zero
	validKey := createTestKey(t)
	SetRootKey(validKey)
	if RootKeyZero() {
		t.Error("After setting valid key, should not be zero")
	}

	// Set another valid key - should overwrite
	newValidKey := createPatternKey(0x77)
	SetRootKey(newValidKey)
	if RootKeyZero() {
		t.Error("Setting new valid key should maintain non-zero state")
	}

	// Verify the new key was actually set
	rootKeyMu.RLock()
	if rootKey[0] != 0x77 {
		t.Errorf("New key should be set: expected 0x77, got 0x%02X", rootKey[0])
	}
	rootKeyMu.RUnlock()
}

func TestRootKeyMemoryOperations(t *testing.T) {
	// Reset to a clean state
	resetRootKey()
	defer resetRootKey()

	// Test mem.Zeroed32 integration
	testKey := &[crypto.AES256KeySize]byte{}

	// All zero keys should be detected as zeroed
	if !mem.Zeroed32(testKey) {
		t.Error("All-zero key should be detected as zeroed")
	}

	// Non-zero key should not be detected as zeroed
	testKey[0] = 1
	if mem.Zeroed32(testKey) {
		t.Error("Non-zero key should not be detected as zeroed")
	}

	// Note: SetRootKey now calls log.FatalErr for zero keys, which is tested
	// separately in TestSetRootKey_ZeroKey using panic recovery.
}

func TestRootKeyDataIntegrity(t *testing.T) {
	// Reset to a clean state
	resetRootKey()

	// Create a test key with a known pattern
	testKey := &[crypto.AES256KeySize]byte{}
	for i := range testKey {
		testKey[i] = byte(i % 256)
	}

	// Set the key
	SetRootKey(testKey)

	// Verify each byte was copied correctly
	rootKeyMu.RLock()
	for i := range rootKey {
		expected := byte(i % 256)
		if rootKey[i] != expected {
			t.Errorf("Byte %d integrity check failed: expected %d, got %d",
				i, expected, rootKey[i])
		}
	}
	rootKeyMu.RUnlock()

	// Clean up
	resetRootKey()
}

func BenchmarkRootKeyZero(b *testing.B) {
	resetRootKey()
	defer resetRootKey()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RootKeyZero()
	}
}

func BenchmarkSetRootKey(b *testing.B) {
	resetRootKey()
	defer resetRootKey()

	testKey := createPatternKey(0x42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetRootKey(testKey)
	}
}

func BenchmarkRootKeyNoLock(b *testing.B) {
	resetRootKey()
	defer resetRootKey()

	// Set a test key first
	testKey := createPatternKey(0x42)
	setRootKeyDirect(testKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RootKeyNoLock()
	}
}

func BenchmarkConcurrentRootKeyZero(b *testing.B) {
	resetRootKey()
	defer resetRootKey()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RootKeyZero()
		}
	})
}
