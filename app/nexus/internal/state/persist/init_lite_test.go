//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"testing"

	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestInitializeLiteBackend_Success(t *testing.T) {
	// Create a test key
	key := createTestKey(t)

	// Test initialization
	backend := initializeLiteBackend(key)

	// Verify backend was created successfully
	if backend == nil {
		t.Fatal("initializeLiteBackend returned nil")
	}
}

func TestInitializeLiteBackend_NilKey_ShouldFatal(t *testing.T) {
	t.Skip("Skipping fatal condition test - initializeLiteBackend calls log.FatalLn which exits the process")

	// This test cannot be run because log.FatalLn calls os.Exit() which would
	// terminate the entire test process. The behavior is verified by manual testing
	// or integration tests that can handle process termination.
	//
	// Expected behavior: initializeLiteBackend panics/exits when called with a nil key
}

func TestInitializeLiteBackend_ZeroKey(t *testing.T) {
	// Create a zero key
	zeroKey := createZeroKey()

	// Test with the zero key - this should work at the initializeLiteBackend level
	// (the validation happens in InitializeBackend, not here)
	backend := initializeLiteBackend(zeroKey)

	// Should succeed - zero key validation happens at a higher level
	if backend == nil {
		t.Error("Expected initializeLiteBackend to succeed with zero key")
	}
}

func TestInitializeLiteBackend_ValidKey_DifferentPatterns(t *testing.T) {
	testCases := []struct {
		name    string
		keyFunc func() *[crypto.AES256KeySize]byte
	}{
		{
			name: "SequentialPattern",
			keyFunc: func() *[crypto.AES256KeySize]byte {
				key := &[crypto.AES256KeySize]byte{}
				for i := range key {
					key[i] = byte(i % 256)
				}
				return key
			},
		},
		{
			name: "AllOnes",
			keyFunc: func() *[crypto.AES256KeySize]byte {
				key := &[crypto.AES256KeySize]byte{}
				for i := range key {
					key[i] = 0xFF
				}
				return key
			},
		},
		{
			name: "AlternatingPattern",
			keyFunc: func() *[crypto.AES256KeySize]byte {
				key := &[crypto.AES256KeySize]byte{}
				for i := range key {
					if i%2 == 0 {
						key[i] = 0xAA
					} else {
						key[i] = 0x55
					}
				}
				return key
			},
		},
		{
			name: "FirstByteOnly",
			keyFunc: func() *[crypto.AES256KeySize]byte {
				key := &[crypto.AES256KeySize]byte{}
				key[0] = 0x01
				return key
			},
		},
		{
			name: "LastByteOnly",
			keyFunc: func() *[crypto.AES256KeySize]byte {
				key := &[crypto.AES256KeySize]byte{}
				key[crypto.AES256KeySize-1] = 0xFF
				return key
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := tc.keyFunc()
			backend := initializeLiteBackend(key)

			if backend == nil {
				t.Errorf("initializeLiteBackend failed with %s key pattern", tc.name)
			}
		})
	}
}

func TestInitializeLiteBackend_MultipleInitializations(t *testing.T) {
	// Create the test key
	key := createTestKey(t)

	// Initialize multiple times
	backends := make([]interface{}, 3)
	for i := 0; i < 3; i++ {
		backend := initializeLiteBackend(key)
		if backend == nil {
			t.Fatalf("Initialization %d failed", i+1)
		}
		backends[i] = backend
	}

	// All should be valid but different instances
	for i := 0; i < len(backends); i++ {
		for j := i + 1; j < len(backends); j++ {
			if backends[i] == backends[j] {
				t.Errorf("Expected different backend instances, got same instance at positions %d and %d", i, j)
			}
		}
	}
}

func TestInitializeLiteBackend_ConcurrentAccess(t *testing.T) {
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	results := make(chan interface{}, numGoroutines)

	// Start multiple goroutines trying to initialize simultaneously
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			// Each goroutine uses its own key to avoid any shared state issues
			key := &[crypto.AES256KeySize]byte{}
			for j := range key {
				key[j] = byte((id + j) % 256)
			}

			backend := initializeLiteBackend(key)
			results <- backend

			if backend == nil {
				t.Errorf("Backend is nil in goroutine %d", id)
				return
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all backends were created successfully
	for i := 0; i < numGoroutines; i++ {
		backend := <-results
		if backend == nil {
			t.Errorf("Backend %d is nil", i)
		}
	}
}

// Benchmark tests
func BenchmarkInitializeLiteBackend(b *testing.B) {
	// Create a test key
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backend := initializeLiteBackend(key)
		if backend == nil {
			b.Fatal("initializeLiteBackend returned nil")
		}
	}
}

func BenchmarkInitializeLiteBackend_DifferentKeys(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate different key for each iteration
		key := &[crypto.AES256KeySize]byte{}
		for j := range key {
			key[j] = byte((i + j) % 256)
		}

		backend := initializeLiteBackend(key)
		if backend == nil {
			b.Fatal("initializeLiteBackend returned nil")
		}
	}
}
