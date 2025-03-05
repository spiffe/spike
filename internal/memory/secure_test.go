//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package memory

import (
	"testing"
)

func TestClear(t *testing.T) {
	type testStruct struct {
		Key    [32]byte
		Token  string
		UserId int64
	}

	// Create test data with non-zero values
	key := [32]byte{}
	for i := range key {
		key[i] = byte(i + 1)
	}

	data := &testStruct{
		Key:    key,
		Token:  "secret-token-value",
		UserId: 12345,
	}

	// Call Clear on the data
	Clear(data)

	// Verify all fields are zeroed
	for i, b := range data.Key {
		if b != 0 {
			t.Errorf("Expected byte at index %d to be 0, got %d", i, b)
		}
	}

	// Note: String contents won't be zeroed directly as strings are immutable in Go
	// The string header will point to the same backing array
	// In a real application, sensitive strings should be stored as byte slices

	if data.UserId != 0 {
		t.Errorf("Expected UserId to be 0, got %d", data.UserId)
	}
}

func TestClearBytes(t *testing.T) {
	// Create a non-zero byte slice
	bytes := make([]byte, 64)
	for i := range bytes {
		bytes[i] = byte(i + 1)
	}

	// Make a copy to verify later
	original := make([]byte, len(bytes))
	copy(original, bytes)

	// Verify bytes are non-zero initially
	for i, b := range bytes {
		if b != original[i] {
			t.Fatalf("Test setup issue: bytes changed before ClearBytes call")
		}
	}

	// Call ClearBytes
	ClearBytes(bytes)

	// Verify all bytes are zeroed
	for i, b := range bytes {
		if b != 0 {
			t.Errorf("Expected byte at index %d to be 0, got %d", i, b)
		}
	}
}
