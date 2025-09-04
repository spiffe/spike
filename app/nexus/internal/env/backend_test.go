//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
)

func TestStoreTypeConstants(t *testing.T) {
	// Test that constants have expected string values
	tests := []struct {
		name      string
		storeType StoreType
		expected  string
	}{
		{"Lite constant", Lite, "lite"},
		{"Sqlite constant", Sqlite, "sqlite"},
		{"Memory constant", Memory, "memory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.storeType) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.storeType))
			}
		})
	}
}

func TestBackendStoreTypeDefault(t *testing.T) {
	// Save the original environment variable
	originalValue := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(env.NexusBackendStore, originalValue)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Unset the environment variable
	_ = os.Unsetenv(env.NexusBackendStore)

	result := BackendStoreType()
	if result != Sqlite {
		t.Errorf("Expected default to be Sqlite, got %s", string(result))
	}
}

func TestBackendStoreTypeValidValues(t *testing.T) {
	// Save the original environment variable
	originalValue := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(env.NexusBackendStore, originalValue)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	tests := []struct {
		name     string
		envValue string
		expected StoreType
	}{
		{"lite value", "lite", Lite},
		{"sqlite value", "sqlite", Sqlite},
		{"memory value", "memory", Memory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv(env.NexusBackendStore, tt.envValue)
			result := BackendStoreType()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", string(tt.expected), string(result))
			}
		})
	}
}

func TestBackendStoreTypeCaseInsensitive(t *testing.T) {
	// Save the original environment variable
	originalValue := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(env.NexusBackendStore, originalValue)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	tests := []struct {
		name     string
		envValue string
		expected StoreType
	}{
		// Test uppercase
		{"LITE uppercase", "LITE", Lite},
		{"SQLITE uppercase", "SQLITE", Sqlite},
		{"MEMORY uppercase", "MEMORY", Memory},

		// Test mixed case
		{"Lite mixed case", "Lite", Lite},
		{"SQLite mixed case", "SQLite", Sqlite},
		{"Memory mixed case", "Memory", Memory},
		{"SqLiTe mixed case", "SqLiTe", Sqlite},

		// Test with extra whitespace (should not match - no trimming)
		{"lite with spaces", " lite ", Sqlite},     // Should default to Sqlite
		{"sqlite with spaces", " sqlite ", Sqlite}, // Should default to Sqlite
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv(env.NexusBackendStore, tt.envValue)
			result := BackendStoreType()
			if result != tt.expected {
				t.Errorf("For env value '%s', expected %s, got %s",
					tt.envValue, string(tt.expected), string(result))
			}
		})
	}
}

func TestBackendStoreTypeInvalidValues(t *testing.T) {
	// Save the original environment variable
	originalValue := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(env.NexusBackendStore, originalValue)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Test invalid values that should default to Sqlite
	invalidValues := []string{
		"invalid",
		"postgres",
		"redis",
		"",
		"123",
		"null",
		"undefined",
		"lite2",
		"sqlite3",
		"mem",
	}

	for _, invalidValue := range invalidValues {
		t.Run("invalid_value_"+invalidValue, func(t *testing.T) {
			_ = os.Setenv(env.NexusBackendStore, invalidValue)
			result := BackendStoreType()
			if result != Sqlite {
				t.Errorf("Invalid value '%s' should default to Sqlite, got %s",
					invalidValue, string(result))
			}
		})
	}
}

func TestBackendStoreTypeEmptyString(t *testing.T) {
	// Save the original environment variable
	originalValue := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(env.NexusBackendStore, originalValue)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Test with an empty string
	_ = os.Setenv(env.NexusBackendStore, "")
	result := BackendStoreType()
	if result != Sqlite {
		t.Errorf("Empty string should default to Sqlite, got %s", string(result))
	}
}

func TestStoreTypeStringConversion(t *testing.T) {
	// Test that StoreType can be converted to string and back
	tests := []struct {
		name      string
		storeType StoreType
	}{
		{"Lite", Lite},
		{"Sqlite", Sqlite},
		{"Memory", Memory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to string
			str := string(tt.storeType)

			// Convert back to StoreType
			backToStoreType := StoreType(str)

			if backToStoreType != tt.storeType {
				t.Errorf("String conversion failed. Original: %s, After conversion: %s",
					string(tt.storeType), string(backToStoreType))
			}
		})
	}
}

func TestBackendStoreTypeConsistency(t *testing.T) {
	// Test that multiple calls with same environment return the same result
	originalValue := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(env.NexusBackendStore, originalValue)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	testValues := []string{"lite", "sqlite", "memory", "invalid"}

	for _, testValue := range testValues {
		t.Run("consistency_"+testValue, func(t *testing.T) {
			_ = os.Setenv("SPIKE_NEXUS_BACKEND_STORE", testValue)

			// Call multiple times
			result1 := BackendStoreType()
			result2 := BackendStoreType()
			result3 := BackendStoreType()

			if result1 != result2 || result2 != result3 {
				t.Errorf("Inconsistent results for value '%s': %s, %s, %s",
					testValue, string(result1), string(result2), string(result3))
			}
		})
	}
}

func TestBackendStoreTypeDocumentedBehavior(t *testing.T) {
	// Test the specific behavior documented in the function comments
	originalValue := os.Getenv(env.NexusBackendStore)
	defer func() {
		if originalValue != "" {
			_ = os.Setenv(env.NexusBackendStore, originalValue)
		} else {
			_ = os.Unsetenv(env.NexusBackendStore)
		}
	}()

	// Test that case-insensitive matching works as documented
	_ = os.Setenv(env.NexusBackendStore, "LITE")
	result := BackendStoreType()
	if result != Lite {
		t.Errorf("Case-insensitive 'LITE' should return Lite, got %s", string(result))
	}
}

func TestStoreTypeComparison(t *testing.T) {
	// Test that StoreType values can be compared correctly

	// noinspection GoBoolExpressions
	if Lite == Sqlite {
		t.Error("Lite should not equal Sqlite")
	}

	// noinspection GoBoolExpressions
	if Sqlite == Memory {
		t.Error("Sqlite should not equal Memory")
	}

	// noinspection GoBoolExpressions
	if Memory == Lite {
		t.Error("Memory should not equal Lite")
	}
}

func TestStoreTypeInSwitch(t *testing.T) {
	// Test that StoreType works correctly in switch statements
	testCases := []struct {
		storeType StoreType
		expected  string
	}{
		{Lite, "is_lite"},
		{Sqlite, "is_sqlite"},
		{Memory, "is_memory"},
	}

	for _, tc := range testCases {
		var result string
		switch tc.storeType {
		case Lite:
			result = "is_lite"
		case Sqlite:
			result = "is_sqlite"
		case Memory:
			result = "is_memory"
		default:
			result = "unknown"
		}

		if result != tc.expected {
			t.Errorf("Switch statement failed for %s. Expected: %s, Got: %s",
				string(tc.storeType), tc.expected, result)
		}
	}
}
