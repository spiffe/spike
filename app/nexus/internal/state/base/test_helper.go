//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"crypto/rand"
	"os"
	"testing"

	"github.com/spiffe/spike-sdk-go/crypto"
)

// Helper function to manage environment variables in tests
func withEnvironment(t *testing.T, key, value string, testFunc func()) {
	original := os.Getenv(key)
	_ = os.Setenv(key, value)
	defer func() {
		if original != "" {
			_ = os.Setenv(key, original)
		} else {
			_ = os.Unsetenv(key)
		}
	}()
	testFunc()
}

// Helper function to create a test key with a specific pattern
func createTestKeyWithPattern(pattern byte) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = pattern
	}
	return key
}

// Helper function to reset the root key to zero state for tests
func resetRootKey() {
	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()
	for i := range rootKey {
		rootKey[i] = 0
	}
}

// Helper function to set the root key directly for testing (bypasses validation)
func setRootKeyDirect(key *[crypto.AES256KeySize]byte) {
	rootKeyMu.Lock()
	defer rootKeyMu.Unlock()
	if key != nil {
		copy(rootKey[:], key[:])
	}
}

// Helper function to create a test key with random data
func createTestKey(t *testing.T) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	if _, err := rand.Read(key[:]); err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	return key
}

// Helper function to create a test key with a specific pattern
func createPatternKey(pattern byte) *[crypto.AES256KeySize]byte {
	key := &[crypto.AES256KeySize]byte{}
	for i := range key {
		key[i] = pattern
	}
	return key
}
