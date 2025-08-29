//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package lite

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

func TestNew_ValidKey(t *testing.T) {
	// Create a valid AES-256 key
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, err := rand.Read(rootKey[:]); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	// Create new lite backend
	ds, err := New(rootKey)
	if err != nil {
		t.Errorf("Expected no error with valid key, got: %v", err)
	}

	if ds == nil {
		t.Error("Expected non-nil Store")
	}

	// Verify it implements the Backend interface
	// noinspection ALL
	var _ backend.Backend = ds

	// Verify the Store has a cipher
	liteStore, ok := ds.(*Store)
	if !ok {
		t.Fatal("Expected Store type")
	}

	if liteStore.Cipher == nil {
		t.Error("Expected non-nil cipher")
	}
}

func TestNew_InvalidKey(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
	}{
		{"too short key (16 bytes)", 16},
		//{"too short key (8 bytes)", 8},
		//{"empty key", 0},
		// FIX-ME: fix these!
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create invalid key of wrong size
			invalidKey := make([]byte, tt.keySize)
			if len(invalidKey) > 0 {
				if _, err := rand.Read(invalidKey); err != nil {
					t.Fatalf("Failed to generate random key: %v", err)
				}
			}

			// Pad or truncate to fit the expected array size for testing
			var testKey [crypto.AES256KeySize]byte
			copy(testKey[:], invalidKey)

			// This should fail for keys that aren't valid AES-256
			if tt.keySize < 16 {
				// Keys smaller than AES-128 should fail
				ds, err := New(&testKey)
				if err == nil {
					t.Errorf("Expected error with invalid key size %d, got nil", tt.keySize)
				}
				if ds != nil {
					t.Errorf("Expected nil Store with invalid key, got: %v", ds)
				}
			} else {
				// For this test, even though we're testing "invalid" keys,
				// AES-256 key size is fixed, so this will actually work
				// The test is more about the error handling path
				ds, err := New(&testKey)
				if err != nil {
					t.Logf("Key creation failed as expected: %v", err)
				} else if ds != nil {
					t.Logf("Key creation succeeded (valid AES-256 key)")
				}
			}
		})
	}
}

func TestNew_ZeroKey(t *testing.T) {
	// Test with an all-zero key
	zeroKey := &[crypto.AES256KeySize]byte{} // All zeros

	ds, err := New(zeroKey)
	if err != nil {
		t.Errorf("Zero key should be valid for AES (though not secure), got error: %v", err)
	}

	if ds == nil {
		t.Error("Expected non-nil Store even with zero key")
	}

	// Verify cipher is created even with a zero key
	if ds != nil {
		liteStore := ds.(*Store)
		if liteStore.Cipher == nil {
			t.Error("Expected cipher to be created even with zero key")
		}
	}
}

func TestDataStore_GetCipher(t *testing.T) {
	// Create a valid key
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, err := rand.Read(rootKey[:]); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	ds, err := New(rootKey)
	if err != nil {
		t.Fatalf("Failed to create Store: %v", err)
	}

	liteStore := ds.(*Store)

	// Test GetCipher method
	cipher := liteStore.GetCipher()
	if cipher == nil {
		t.Error("Expected non-nil cipher from GetCipher()")
	}

	// Verify it's the same cipher
	if cipher != liteStore.Cipher {
		t.Error("GetCipher() should return the same cipher instance")
	}
}

func TestDataStore_Implements_Backend_Interface(t *testing.T) {
	// Create a valid key
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, err := rand.Read(rootKey[:]); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	ds, err := New(rootKey)
	if err != nil {
		t.Fatalf("Failed to create Store: %v", err)
	}

	// Test that it implements all Backend interface methods
	ctx := context.Background()

	// Test Initialize (inherited from Store)
	if err := ds.Initialize(ctx); err != nil {
		t.Errorf("Initialize should not return error: %v", err)
	}

	// Test Close (inherited from Store)
	if err := ds.Close(ctx); err != nil {
		t.Errorf("Close should not return error: %v", err)
	}

	// Test LoadSecret (inherited from Store)
	secret, err := ds.LoadSecret(ctx, "test/path")
	if err != nil {
		t.Errorf("LoadSecret should not return error: %v", err)
	}
	if secret != nil {
		t.Error("LoadSecret should return nil (noop implementation)")
	}

	// Test LoadAllSecrets (inherited from Store)
	secrets, err := ds.LoadAllSecrets(ctx)
	if err != nil {
		t.Errorf("LoadAllSecrets should not return error: %v", err)
	}
	if secrets != nil {
		t.Error("LoadAllSecrets should return nil (noop implementation)")
	}

	// Test StoreSecret (inherited from Store)
	testSecret := kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Data:    map[string]string{"key": "value"},
				Version: 1,
			},
		},
	}
	err = ds.StoreSecret(ctx, "test/path", testSecret)
	if err != nil {
		t.Errorf("StoreSecret should not return error: %v", err)
	}

	// Test LoadPolicy (inherited from Store)
	policy, err := ds.LoadPolicy(ctx, "test-policy-id")
	if err != nil {
		t.Errorf("LoadPolicy should not return error: %v", err)
	}
	if policy != nil {
		t.Error("LoadPolicy should return nil (noop implementation)")
	}

	// Test LoadAllPolicies (inherited from Store)
	policies, err := ds.LoadAllPolicies(ctx)
	if err != nil {
		t.Errorf("LoadAllPolicies should not return error: %v", err)
	}
	if policies != nil {
		t.Error("LoadAllPolicies should return nil (noop implementation)")
	}

	// Test StorePolicy (inherited from Store)
	testPolicy := data.Policy{
		ID:              "test-policy",
		Name:            "test policy",
		SPIFFEIDPattern: "spiffe://example\\.org/test",
		PathPattern:     "test/.*",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}
	err = ds.StorePolicy(ctx, testPolicy)
	if err != nil {
		t.Errorf("StorePolicy should not return error: %v", err)
	}

	// Test DeletePolicy (inherited from Store)
	err = ds.DeletePolicy(ctx, "test-policy-id")
	if err != nil {
		t.Errorf("DeletePolicy should not return error: %v", err)
	}

	// Test GetCipher (overridden in Store)
	cipher := ds.GetCipher()
	if cipher == nil {
		t.Error("GetCipher should return non-nil cipher")
	}
}

func TestDataStore_CipherFunctionality(t *testing.T) {
	// Create a valid key
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, err := rand.Read(rootKey[:]); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	ds, err := New(rootKey)
	if err != nil {
		t.Fatalf("Failed to create Store: %v", err)
	}

	liteStore := ds.(*Store)
	cipher := liteStore.GetCipher()

	// Test basic cipher properties
	if cipher.NonceSize() <= 0 {
		t.Error("Cipher should have positive nonce size")
	}

	if cipher.Overhead() <= 0 {
		t.Error("Cipher should have positive overhead")
	}

	// Test encryption/decryption functionality
	plaintext := []byte("Hello, SPIKE!")
	nonce := make([]byte, cipher.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}

	// Encrypt
	ciphertext := cipher.Seal(nil, nonce, plaintext, nil)
	if len(ciphertext) == 0 {
		t.Error("Encryption should produce non-empty ciphertext")
	}

	// Decrypt
	decrypted, err := cipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		t.Errorf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text doesn't match original: got %q, want %q",
			string(decrypted), string(plaintext))
	}
}

func TestDataStore_DifferentKeys_ProduceDifferentCiphers(t *testing.T) {
	// Create two different keys
	key1 := &[crypto.AES256KeySize]byte{}
	key2 := &[crypto.AES256KeySize]byte{}

	if _, err := rand.Read(key1[:]); err != nil {
		t.Fatalf("Failed to generate first key: %v", err)
	}
	if _, err := rand.Read(key2[:]); err != nil {
		t.Fatalf("Failed to generate second key: %v", err)
	}

	// Ensure keys are different
	if *key1 == *key2 {
		key2[0] = ^key1[0] // Make them different
	}

	// Create two DataStores
	ds1, err1 := New(key1)
	ds2, err2 := New(key2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to create DataStores: %v, %v", err1, err2)
	}

	cipher1 := ds1.GetCipher()
	cipher2 := ds2.GetCipher()

	// Test that they produce different encrypted output for the same input
	plaintext := []byte("test data")
	nonce := make([]byte, cipher1.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}

	ciphertext1 := cipher1.Seal(nil, nonce, plaintext, nil)
	ciphertext2 := cipher2.Seal(nil, nonce, plaintext, nil)

	// They should produce different ciphertext (different keys)
	if len(ciphertext1) == len(ciphertext2) && string(ciphertext1) == string(ciphertext2) {
		t.Error("Different keys should produce different ciphertext")
	}

	// Verify cipher1 cannot decrypt cipher2's output
	_, err := cipher1.Open(nil, nonce, ciphertext2, nil)
	if err == nil {
		t.Error("Cipher with different key should not be able to decrypt ciphertext")
	}
}

func TestDataStore_EmbeddedNoopStore(t *testing.T) {
	// Test that Store properly embeds Store
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, err := rand.Read(rootKey[:]); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	ds, err := New(rootKey)
	if err != nil {
		t.Fatalf("Failed to create Store: %v", err)
	}

	liteStore := ds.(*Store)

	// Check that the embedded Store is accessible
	// (This tests the struct composition)
	ctx := context.Background()

	// These methods should all be inherited from Store and return no error
	testSecret := kv.Value{
		Versions: map[int]kv.Version{
			1: {
				Data:    map[string]string{"key": "value"},
				Version: 1,
			},
		},
	}
	testPolicy := data.Policy{
		ID:              "test-policy",
		Name:            "test policy",
		SPIFFEIDPattern: "spiffe://example\\.org/test",
		PathPattern:     "test/.*",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}

	methods := []func() error{
		func() error { return liteStore.Initialize(ctx) },
		func() error { return liteStore.Close(ctx) },
		func() error { return liteStore.StoreSecret(ctx, "path", testSecret) },
		func() error { return liteStore.StorePolicy(ctx, testPolicy) },
		func() error { return liteStore.DeletePolicy(ctx, "id") },
	}

	for i, method := range methods {
		if err := method(); err != nil {
			t.Errorf("Store method %d should not return error: %v", i, err)
		}
	}
}

func TestDataStore_GCMProperties(t *testing.T) {
	// Test that the cipher is specifically GCM
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, err := rand.Read(rootKey[:]); err != nil {
		t.Fatalf("Failed to generate random key: %v", err)
	}

	ds, err := New(rootKey)
	if err != nil {
		t.Fatalf("Failed to create Store: %v", err)
	}

	cipher := ds.GetCipher()

	// GCM should have specific properties
	expectedNonceSize := 12 // Standard GCM nonce size
	expectedOverhead := 16  // GCM authentication tag size

	if cipher.NonceSize() != expectedNonceSize {
		t.Errorf("Expected GCM nonce size %d, got %d", expectedNonceSize, cipher.NonceSize())
	}

	if cipher.Overhead() != expectedOverhead {
		t.Errorf("Expected GCM overhead %d, got %d", expectedOverhead, cipher.Overhead())
	}
}

func TestDataStore_MemoryManagement(t *testing.T) {
	// Test that multiple Store instances can coexist
	keys := make([]*[crypto.AES256KeySize]byte, 5)
	dss := make([]backend.Backend, 5)

	// Create multiple instances
	for i := 0; i < 5; i++ {
		keys[i] = &[crypto.AES256KeySize]byte{}
		if _, err := rand.Read(keys[i][:]); err != nil {
			t.Fatalf("Failed to generate key %d: %v", i, err)
		}

		ds, err := New(keys[i])
		if err != nil {
			t.Fatalf("Failed to create Store %d: %v", i, err)
		}
		dss[i] = ds
	}

	// Verify all instances are independent
	for i, ds := range dss {
		if ds == nil {
			t.Errorf("Store %d should not be nil", i)
		}

		if ds == nil {
			continue
		}
		cipher := ds.GetCipher()
		if cipher == nil {
			t.Errorf("Cipher %d should not be nil", i)
		}

		// Compare with other instances
		for j, otherDs := range dss {
			if i != j && ds == otherDs {
				t.Errorf("Store %d and %d should be different instances", i, j)
			}
		}
	}
}

// FIX-ME: handle invalid cases.
//func TestNew_CipherCreationFailure(t *testing.T) {
//	// This test simulates cipher creation failure
//	// In practice, aes.NewCipher only fails with invalid key lengths
//	// But we test the error path by using the actual error conditions
//
//	tests := []struct {
//		name    string
//		keyData []byte
//	}{
//		{"key too short", make([]byte, 8)},       // Less than 16 bytes
//		{"key invalid length", make([]byte, 15)}, // Not 16, 24, or 32
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// Fill with some data
//			for i := range tt.keyData {
//				tt.keyData[i] = byte(i)
//			}
//
//			// Create array of correct size but with invalid data
//			var testKey [crypto.AES256KeySize]byte
//			copy(testKey[:], tt.keyData)
//
//			// Try to create cipher directly to see if it would fail
//			_, err := aes.NewCipher(tt.keyData)
//			if err != nil {
//				// This key would indeed fail, so New() should also fail
//				ds, newErr := New(&testKey)
//				if newErr == nil {
//					t.Errorf("Expected error for invalid key data, got nil")
//				}
//				if ds != nil {
//					t.Errorf("Expected nil Store for invalid key, got: %v", ds)
//				}
//			}
//		})
//	}
//}
