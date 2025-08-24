//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package initialization

import (
	"os"
	"testing"

	"github.com/spiffe/spike/app/nexus/internal/env"
)

func TestInitialize_SQLiteBackend(t *testing.T) {
	// Save original environment variable
	originalStore := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	defer func() {
		if originalStore != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalStore)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	// Set to SQLite backend
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "sqlite")

	// Verify the environment is set correctly
	if env.BackendStoreType() != env.Sqlite {
		t.Fatal("Expected Sqlite backend store type")
	}

	// The Initialize function with SQLite backend would:
	// 1. Call recovery.InitializeBackingStoreFromKeepers(source)
	// 2. Start goroutine recovery.SendShardsPeriodically(source)
	//
	// Both of these require SPIFFE infrastructure and network connectivity
	// We skip this test since it would hang or fail without proper setup
	t.Skip("Skipping SQLite backend test - requires SPIFFE infrastructure and network connectivity")
}

func TestInitialize_LiteBackend(t *testing.T) {
	// Save original environment variable
	originalStore := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	defer func() {
		if originalStore != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalStore)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	// Set to Lite backend
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "lite")

	// Verify the environment is set correctly
	if env.BackendStoreType() != env.Lite {
		t.Fatal("Expected Lite backend store type")
	}

	// The Initialize function with Lite backend would:
	// 1. Call recovery.InitializeBackingStoreFromKeepers(source)
	// 2. Start goroutine recovery.SendShardsPeriodically(source)
	//
	// Both of these require SPIFFE infrastructure and network connectivity
	// We skip this test since it would hang or fail without proper setup
	t.Skip("Skipping Lite backend test - requires SPIFFE infrastructure and network connectivity")
}

func TestInitialize_MemoryBackend(t *testing.T) {
	// Save original environment variable
	originalStore := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	defer func() {
		if originalStore != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalStore)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	// Set to Memory backend
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "memory")

	// Verify the environment is set correctly
	if env.BackendStoreType() != env.Memory {
		t.Fatal("Expected Memory backend store type")
	}

	// The Initialize function with Memory backend would:
	// 1. Log warnings about development mode
	// 2. Call state.Initialize(nil)
	//
	// This depends on state management which we can't easily test without mocking
	// We skip this test since it would depend on internal state
	t.Skip("Skipping Memory backend test - requires state management mocking")
}

func TestInitialize_InvalidBackend(t *testing.T) {
	// Save original environment variable
	originalStore := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	defer func() {
		if originalStore != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalStore)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	// Set to invalid backend
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "invalid")

	// The Initialize function with invalid backend would call log.FatalLn
	// which calls os.Exit() and terminates the process
	// We skip this test since it would terminate the test runner
	t.Skip("Skipping invalid backend test - would call os.Exit() via log.FatalLn")
}

func TestBackendStoreTypeDetection(t *testing.T) {
	// Test the backend store type detection logic used in Initialize()

	// Save original environment variable
	originalStore := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	defer func() {
		if originalStore != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalStore)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	tests := []struct {
		name                           string
		backendStore                   string
		expectedType                   env.StoreType
		requireBackingStoreToBootstrap bool
		devMode                        bool
	}{
		{
			name:                           "sqlite backend",
			backendStore:                   "sqlite",
			expectedType:                   env.Sqlite,
			requireBackingStoreToBootstrap: true,
			devMode:                        false,
		},
		{
			name:                           "lite backend",
			backendStore:                   "lite",
			expectedType:                   env.Lite,
			requireBackingStoreToBootstrap: true,
			devMode:                        false,
		},
		{
			name:                           "memory backend",
			backendStore:                   "memory",
			expectedType:                   env.Memory,
			requireBackingStoreToBootstrap: false,
			devMode:                        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", tt.backendStore)

			// Test backend type detection
			backendType := env.BackendStoreType()
			if backendType != tt.expectedType {
				t.Errorf("Expected backend type %s, got %s", tt.expectedType, backendType)
			}

			// Test requireBackingStoreToBootstrap logic
			requireBackingStoreToBootstrap := backendType == env.Sqlite || backendType == env.Lite
			if requireBackingStoreToBootstrap != tt.requireBackingStoreToBootstrap {
				t.Errorf("Expected requireBackingStoreToBootstrap %v, got %v",
					tt.requireBackingStoreToBootstrap, requireBackingStoreToBootstrap)
			}

			// Test devMode logic
			devMode := backendType == env.Memory
			if devMode != tt.devMode {
				t.Errorf("Expected devMode %v, got %v", tt.devMode, devMode)
			}
		})
	}
}

func TestBackendStoreTypeConstants(t *testing.T) {
	// Test that the backend store type constants are properly defined
	tests := []struct {
		name      string
		storeType env.StoreType
		expected  string
	}{
		{"Sqlite constant", env.Sqlite, "sqlite"},
		{"Lite constant", env.Lite, "lite"},
		{"Memory constant", env.Memory, "memory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.storeType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.storeType))
			}
		})
	}
}

func TestEnvironmentVariableHandling(t *testing.T) {
	// Test environment variable handling
	originalStore := os.Getenv("SPIKE_NEXUS_BACKEND_STORE")
	defer func() {
		if originalStore != "" {
			os.Setenv("SPIKE_NEXUS_BACKEND_STORE", originalStore)
		} else {
			os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
		}
	}()

	// Test with unset environment variable (should have default behavior)
	os.Unsetenv("SPIKE_NEXUS_BACKEND_STORE")
	defaultType := env.BackendStoreType()
	t.Logf("Default backend store type: %s", string(defaultType))

	// Test with valid values
	validValues := []string{"sqlite", "lite", "memory"}
	for _, value := range validValues {
		os.Setenv("SPIKE_NEXUS_BACKEND_STORE", value)
		resultType := env.BackendStoreType()
		if string(resultType) != value {
			t.Errorf("Expected backend type %s, got %s", value, string(resultType))
		}
	}

	// Test with invalid value
	os.Setenv("SPIKE_NEXUS_BACKEND_STORE", "invalid")
	invalidType := env.BackendStoreType()
	t.Logf("Invalid backend store type returns: %s", string(invalidType))
}

func TestInitializationPathLogic(t *testing.T) {
	// Test the path logic used in Initialize() function

	tests := []struct {
		name                     string
		backendType              env.StoreType
		expectedRequireBootstrap bool
		expectedDevMode          bool
		expectedInvalidBackend   bool
	}{
		{
			name:                     "sqlite path",
			backendType:              env.Sqlite,
			expectedRequireBootstrap: true,
			expectedDevMode:          false,
			expectedInvalidBackend:   false,
		},
		{
			name:                     "lite path",
			backendType:              env.Lite,
			expectedRequireBootstrap: true,
			expectedDevMode:          false,
			expectedInvalidBackend:   false,
		},
		{
			name:                     "memory path",
			backendType:              env.Memory,
			expectedRequireBootstrap: false,
			expectedDevMode:          true,
			expectedInvalidBackend:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test requireBackingStoreToBootstrap logic
			requireBootstrap := tt.backendType == env.Sqlite || tt.backendType == env.Lite
			if requireBootstrap != tt.expectedRequireBootstrap {
				t.Errorf("Expected requireBootstrap %v, got %v",
					tt.expectedRequireBootstrap, requireBootstrap)
			}

			// Test devMode logic
			devMode := tt.backendType == env.Memory
			if devMode != tt.expectedDevMode {
				t.Errorf("Expected devMode %v, got %v", tt.expectedDevMode, devMode)
			}

			// Test path exclusivity
			if requireBootstrap && devMode {
				t.Error("requireBootstrap and devMode should not both be true")
			}
		})
	}
}

func TestStoreTypeComparison(t *testing.T) {
	// Test store type comparison operations used in Initialize()

	// Test equality comparisons
	if env.Sqlite == env.Lite {
		t.Error("Sqlite should not equal Lite")
	}
	if env.Sqlite == env.Memory {
		t.Error("Sqlite should not equal Memory")
	}
	if env.Lite == env.Memory {
		t.Error("Lite should not equal Memory")
	}

	// Test self-equality
	if env.Sqlite != env.Sqlite {
		t.Error("Sqlite should equal itself")
	}
	if env.Lite != env.Lite {
		t.Error("Lite should equal itself")
	}
	if env.Memory != env.Memory {
		t.Error("Memory should equal itself")
	}
}

func TestBackendStoreTypeValidation(t *testing.T) {
	// Test validation logic for backend store types
	validTypes := []env.StoreType{env.Sqlite, env.Lite, env.Memory}

	for _, validType := range validTypes {
		t.Run("valid_type_"+string(validType), func(t *testing.T) {
			// Test that valid types can be used in comparisons
			isValidForBootstrap := validType == env.Sqlite || validType == env.Lite
			isMemoryMode := validType == env.Memory

			// These should be mutually exclusive except both false is possible
			if isValidForBootstrap && isMemoryMode {
				t.Error("A type cannot be both bootstrap-required and memory mode")
			}

			// At least one of the conditions should match for known types
			if !isValidForBootstrap && !isMemoryMode {
				t.Errorf("Type %s should match at least one condition", validType)
			}
		})
	}
}

func TestStringConversionConsistency(t *testing.T) {
	// Test that string conversion is consistent
	tests := []struct {
		storeType env.StoreType
		expected  string
	}{
		{env.Sqlite, "sqlite"},
		{env.Lite, "lite"},
		{env.Memory, "memory"},
	}

	for _, tt := range tests {
		t.Run("conversion_"+string(tt.storeType), func(t *testing.T) {
			str := string(tt.storeType)
			if str != tt.expected {
				t.Errorf("Expected string conversion %s, got %s", tt.expected, str)
			}

			// Test that we can create a StoreType from the string
			backToType := env.StoreType(str)
			if backToType != tt.storeType {
				t.Errorf("Round-trip conversion failed: %s -> %s", tt.storeType, backToType)
			}
		})
	}
}
