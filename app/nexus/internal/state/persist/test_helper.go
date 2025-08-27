//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"github.com/spiffe/spike-sdk-go/crypto"
	"os"
	"testing"
)

// Helper function to create a test root key with a specific pattern
func createTestKey(_ *testing.T) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = byte(i % 256) // Predictable pattern for testing
	}
	return key
}

// Helper function to create a zero key
func createZeroKey() *[crypto.AES256KeySize]byte {
	return &[crypto.AES256KeySize]byte{} // All zeros
}

// TestingInterface allows both *testing.T and *testing.B to be used
type TestingInterface interface {
	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// Helper function to set the environment variable and restore it after test
func withEnvironment(_ TestingInterface, key, value string, testFunc func()) {
	original := os.Getenv(key)
	defer func() {
		if original != "" {
			_ = os.Setenv(key, original)
		} else {
			_ = os.Unsetenv(key)
		}
	}()

	if value != "" {
		_ = os.Setenv(key, value)
	} else {
		_ = os.Unsetenv(key)
	}

	testFunc()
}
