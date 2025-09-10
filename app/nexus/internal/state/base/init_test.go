//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"os"
	"testing"

	appEnv "github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/env"
)

func TestInitialize_MemoryBackend_ValidKey(t *testing.T) {
	withEnvironment(t, appEnv.NexusBackendStore, "memory", func() {
		// Verify the environment is set correctly
		if env.BackendStoreType() != env.Memory {
			t.Fatal("Expected Memory backend store type")
		}

		// Create a valid test key
		// testKey := createRandomTestKey(t)

		// With the new defensive approach, memory backends MUST be initialized with nil keys
		// Passing a valid key to memory backend should cause log.FatalLn
		t.Skip("Skipping test that would call log.FatalLn - memory backend must use nil key")
	})
}

func TestInitialize_MemoryBackend_NilKey(t *testing.T) {
	// Reset root key state
	resetRootKey()
	defer resetRootKey()

	withEnvironment(t, "SPIKE_NEXUS_BACKEND_STORE", "memory", func() {
		// Verify the environment is set correctly
		if env.BackendStoreType() != env.Memory {
			t.Fatal("Expected Memory backend store type")
		}

		// Initialize with memory backend and nil key - this should now work
		// The new defensive approach requires memory backends to use nil keys
		Initialize(nil)

		// Root key should remain zero for memory backend
		if !RootKeyZero() {
			t.Error("Root key should remain zero for memory backend")
		}
	})
}

func TestInitialize_MemoryBackend_ZeroKey(t *testing.T) {
	withEnvironment(t, appEnv.NexusBackendStore, "memory", func() {
		// Verify the environment is set correctly
		if env.BackendStoreType() != env.Memory {
			t.Fatal("Expected Memory backend store type")
		}

		// Create a zero key
		// zeroKey := &[crypto.AES256KeySize]byte{} // All zeros

		// With the new defensive approach, memory backends MUST be initialized with nil keys
		// Passing any non-nil key (including the zero key) to memory backend should cause log.FatalLn
		t.Skip("Skipping test that would call log.FatalLn - memory backend must use nil key")
	})
}

func TestInitialize_NonMemoryBackend_ValidKey(t *testing.T) {
	// Reset root key state
	resetRootKey()
	defer resetRootKey()

	withEnvironment(t, appEnv.NexusBackendStore, "sqlite", func() {
		// Verify the environment is set correctly
		if env.BackendStoreType() != env.Sqlite {
			t.Fatal("Expected SQLite backend store type")
		}

		// Create a valid test key
		testKey := createTestKeyWithPattern(0x42)

		// Initialize with a non-memory backend and valid key
		Initialize(testKey)

		// Root key should be set for non-memory backend
		if RootKeyZero() {
			t.Error("Root key should not be zero after initialization with valid key")
		}

		// Verify the key was set correctly
		rootKeyMu.RLock()
		if rootKey[0] != 0x42 {
			t.Errorf("Expected root key first byte to be 0x42, got 0x%02X", rootKey[0])
		}
		rootKeyMu.RUnlock()
	})
}

func TestInitialize_NonMemoryBackend_NilKey(t *testing.T) {
	withEnvironment(t, appEnv.NexusBackendStore, "sqlite", func() {
		// Verify the environment is set correctly
		if env.BackendStoreType() != env.Sqlite {
			t.Fatal("Expected SQLite backend store type")
		}

		// This test would call log.FatalLn which terminates the process
		// We skip this test since it would terminate the test runner
		t.Skip("Skipping test that would call log.FatalLn with nil key - would terminate process")
	})
}

func TestInitialize_NonMemoryBackend_ZeroKey(t *testing.T) {
	withEnvironment(t, appEnv.NexusBackendStore, "lite", func() {
		// Verify the environment is set correctly
		if env.BackendStoreType() != env.Lite {
			t.Fatal("Expected Lite backend store type")
		}

		// This test would call log.FatalLn which terminates the process
		// We skip this test since it would terminate the test runner
		t.Skip("Skipping test that would call log.FatalLn with zero key - would terminate process")
	})
}

func TestInitialize_KeyValidation(t *testing.T) {
	// Test that the key validation logic works as expected
	tests := []struct {
		name     string
		keySetup func() *[crypto.AES256KeySize]byte
		isValid  bool
	}{
		{
			name:     "nil key",
			keySetup: func() *[crypto.AES256KeySize]byte { return nil },
			isValid:  false,
		},
		{
			name: "zero key",
			keySetup: func() *[crypto.AES256KeySize]byte {
				return &[crypto.AES256KeySize]byte{} // All zeros
			},
			isValid: false,
		},
		{
			name: "valid random key",
			keySetup: func() *[crypto.AES256KeySize]byte {
				key := &[crypto.AES256KeySize]byte{}
				key[0] = 1 // At least one non-zero byte
				return key
			},
			isValid: true,
		},
		{
			name: "pattern key",
			keySetup: func() *[crypto.AES256KeySize]byte {
				return createTestKeyWithPattern(0xFF)
			},
			isValid: true,
		},
		{
			name: "key with only last byte set",
			keySetup: func() *[crypto.AES256KeySize]byte {
				key := &[crypto.AES256KeySize]byte{}
				key[crypto.AES256KeySize-1] = 0x99
				return key
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.keySetup()

			// Test validation logic directly (matches Initialize function logic)
			isNilOrZero := key == nil || mem.Zeroed32(key)

			if isNilOrZero == tt.isValid {
				t.Errorf("Key validation mismatch: expected valid=%v, got isNilOrZero=%v",
					tt.isValid, isNilOrZero)
			}
		})
	}
}

func TestInitialize_DifferentBackendTypes(t *testing.T) {
	// Test behavior with different backend store types
	backendTests := []struct {
		name        string
		backendType string
		envType     env.StoreType
		isMemory    bool
	}{
		{"memory backend", "memory", env.Memory, true},
		{"sqlite backend", "sqlite", env.Sqlite, false},
		{"lite backend", "lite", env.Lite, false},
	}

	for _, bt := range backendTests {
		t.Run(bt.name, func(t *testing.T) {
			// Reset root key state
			resetRootKey()
			defer resetRootKey()

			withEnvironment(t, appEnv.NexusBackendStore, bt.backendType, func() {
				// Verify the environment is set correctly
				if env.BackendStoreType() != bt.envType {
					t.Fatalf("Expected %s backend store type", bt.name)
				}

				if bt.isMemory {
					// Memory backend MUST use nil key with the new defensive approach
					Initialize(nil)
					// Root key should remain zero for memory backend
					if !RootKeyZero() {
						t.Error("Root key should remain zero for memory backend")
					}
				} else {
					// Non-memory backend should set the root key
					testKey := createTestKeyWithPattern(0x55)
					Initialize(testKey)
					// Root key should be set
					if RootKeyZero() {
						t.Error("Root key should not be zero after initialization")
					}

					// Verify the key was actually set
					rootKeyMu.RLock()
					if rootKey[0] != 0x55 {
						t.Errorf("Expected root key first byte to be 0x55, got 0x%02X", rootKey[0])
					}
					rootKeyMu.RUnlock()
				}
			})
		})
	}
}

func TestInitialize_CallsSetRootKey(t *testing.T) {
	// Reset root key state
	resetRootKey()
	defer resetRootKey()

	withEnvironment(t, appEnv.NexusBackendStore, "sqlite", func() {
		// Create a unique test key
		testKey := createTestKeyWithPattern(0xAB)

		// Initialize with the key
		Initialize(testKey)

		// Verify SetRootKey was called by checking the result
		rootKeyMu.RLock()
		allMatch := true
		for i, b := range rootKey {
			if b != 0xAB {
				allMatch = false
				t.Errorf("Root key byte %d should be 0xAB, got 0x%02X", i, b)
			}
		}
		rootKeyMu.RUnlock()

		if !allMatch {
			t.Error("SetRootKey was not called properly")
		}
	})
}

func TestInitialize_KeyIndependence(t *testing.T) {
	// Reset root key state
	resetRootKey()
	defer resetRootKey()

	withEnvironment(t, appEnv.NexusBackendStore, "sqlite", func() {
		// Create a test key
		testKey := createTestKeyWithPattern(0x12)
		originalFirstByte := testKey[0]

		// Initialize with the key
		Initialize(testKey)

		// Modify the original key after initialization
		testKey[0] = 0x99

		// Verify the internal root key wasn't affected
		rootKeyMu.RLock()
		if rootKey[0] != originalFirstByte {
			t.Errorf("Root key should not be affected by changes to source key: expected 0x%02X, got 0x%02X",
				originalFirstByte, rootKey[0])
		}
		rootKeyMu.RUnlock()
	})
}

func TestInitialize_MemoryVersusNonMemoryBehavior(t *testing.T) {
	// Test the difference in behavior between memory and non-memory backends
	testKey := createTestKeyWithPattern(0x33)

	// Test memory backend
	resetRootKey()
	withEnvironment(t, appEnv.NexusBackendStore, "memory", func() {
		Initialize(nil) // Memory backend MUST use nil key with the new defensive approach
		memoryResult := RootKeyZero()
		if !memoryResult {
			t.Error("Memory backend should leave root key as zero")
		}
	})

	// Test non-memory backend
	resetRootKey()
	withEnvironment(t, appEnv.NexusBackendStore, "sqlite", func() {
		Initialize(testKey)
		nonMemoryResult := RootKeyZero()
		if nonMemoryResult {
			t.Error("Non-memory backend should set root key to non-zero")
		}
	})

	// Clean up
	resetRootKey()
}

func TestInitialize_EnvironmentVariableHandling(t *testing.T) {
	// Test that the function properly reads environment variables
	originalValue := os.Getenv(appEnv.NexusBackendStore)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(appEnv.NexusBackendStore, originalValue)
		} else {
			_ = os.Unsetenv(appEnv.NexusBackendStore)
		}
	}()

	testCases := []string{"memory", "sqlite", "lite"}

	for _, backendType := range testCases {
		t.Run("backend_"+backendType, func(t *testing.T) {
			resetRootKey()
			defer resetRootKey()

			_ = os.Setenv(appEnv.NexusBackendStore, backendType)

			// Verify the environment variable is read correctly
			actualType := env.BackendStoreType()
			expectedType := env.StoreType(backendType)

			if actualType != expectedType {
				t.Errorf("Expected backend type %s, got %s", expectedType, actualType)
			}

			// Test initialization behavior
			if backendType == "memory" {
				// Memory backend MUST use nil key with the internal defensive approach
				Initialize(nil)
			} else {
				// Non-memory backends require valid keys
				testKey := createTestKeyWithPattern(0x66)
				Initialize(testKey)
			}

			// Verify expected behavior based on the backend type
			isRootKeyZero := RootKeyZero()
			shouldBeZero := backendType == "memory"

			if isRootKeyZero != shouldBeZero {
				t.Errorf("For backend %s, expected root key zero: %v, got: %v",
					backendType, shouldBeZero, isRootKeyZero)
			}
		})
	}
}

// Benchmark tests
func BenchmarkInitialize_MemoryBackend(b *testing.B) {
	// Save and restore environment variable
	original := os.Getenv(appEnv.NexusBackendStore)
	_ = os.Setenv(appEnv.NexusBackendStore, "memory")
	defer func() {
		if original != "" {
			_ = os.Setenv(appEnv.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(appEnv.NexusBackendStore)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetRootKey()
		Initialize(nil) // Memory backend MUST use nil key
	}
}

func BenchmarkInitialize_NonMemoryBackend(b *testing.B) {
	// Save and restore environment variable
	original := os.Getenv(appEnv.NexusBackendStore)
	_ = os.Setenv(appEnv.NexusBackendStore, "sqlite")
	defer func() {
		if original != "" {
			_ = os.Setenv(appEnv.NexusBackendStore, original)
		} else {
			_ = os.Unsetenv(appEnv.NexusBackendStore)
		}
	}()

	testKey := createTestKeyWithPattern(0x44)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetRootKey()
		Initialize(testKey)
	}
}
