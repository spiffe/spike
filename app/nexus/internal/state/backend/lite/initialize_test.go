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
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

func TestNew_ValidKey(t *testing.T) {
	// Create a valid AES-256 key
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, randErr := rand.Read(rootKey[:]); randErr != nil {
		t.Fatalf("Failed to generate random key: %v", randErr)
	}

	// Create new lite backend
	ds, newErr := New(rootKey)
	if newErr != nil {
		t.Errorf("Expected no error with valid key, got: %v", newErr)
	}

	if ds == nil {
		t.Fatal("Expected non-nil Store")
		return
	}

	// Verify the Store has a cipher
	liteStore, ok := ds.(*Store)
	if !ok {
		t.Fatal("Expected Store type")
		return
	}

	if liteStore.Cipher == nil {
		t.Error("Expected non-nil cipher")
	}
}

// TestNew_NilKey verifies that New rejects a nil root key instead of
// dereferencing it. The key length is fixed by the parameter type, so a nil
// pointer is the only structurally invalid input New can receive.
func TestNew_NilKey(t *testing.T) {
	ds, newErr := New(nil)
	if newErr == nil {
		t.Fatal("Expected an error for a nil root key, got nil")
		return
	}
	if newErr.Code != sdkErrors.ErrRootKeyMissing.Code {
		t.Errorf("Expected error code %q, got %q",
			sdkErrors.ErrRootKeyMissing.Code, newErr.Code)
	}
	if ds != nil {
		t.Errorf("Expected a nil Store for a nil root key, got: %v", ds)
	}
}

func TestNew_ZeroKey(t *testing.T) {
	// Test with an all-zero key
	zeroKey := &[crypto.AES256KeySize]byte{} // All zeros

	ds, newErr := New(zeroKey)
	if newErr != nil {
		t.Errorf(
			"Zero key should be valid for AES (though not secure), got error: %v",
			newErr,
		)
	}

	if ds == nil {
		t.Error("Expected non-nil Store even with zero key")
	}

	// Verify cipher is created even with a zero key
	if ds != nil {
		liteStore, ok := ds.(*Store)
		if !ok {
			t.Fatal("Expected Store type")
			return
		}
		if liteStore.Cipher == nil {
			t.Error("Expected cipher to be created even with zero key")
		}
	}
}

func TestDataStore_GetCipher(t *testing.T) {
	// Create a valid key
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, randErr := rand.Read(rootKey[:]); randErr != nil {
		t.Fatalf("Failed to generate random key: %v", randErr)
	}

	ds, newErr := New(rootKey)
	if newErr != nil {
		t.Fatalf("Failed to create Store: %v", newErr)
	}

	liteStore, ok := ds.(*Store)
	if !ok {
		t.Fatal("Expected Store type")
		return
	}

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
	if _, randErr := rand.Read(rootKey[:]); randErr != nil {
		t.Fatalf("Failed to generate random key: %v", randErr)
	}

	ds, newErr := New(rootKey)
	if newErr != nil {
		t.Fatalf("Failed to create Store: %v", newErr)
	}

	// Test that it implements all Backend interface methods
	ctx := context.Background()

	// Test Initialize (inherited from Store)
	if initErr := ds.Initialize(ctx); initErr != nil {
		t.Errorf("Initialize should not return error: %v", initErr)
	}

	// Test Close (inherited from Store)
	if closeErr := ds.Close(ctx); closeErr != nil {
		t.Errorf("Close should not return error: %v", closeErr)
	}

	// Test LoadSecret (inherited from Store)
	secret, loadSecretErr := ds.LoadSecret(ctx, "test/path")
	if loadSecretErr != nil {
		t.Errorf("LoadSecret should not return error: %v", loadSecretErr)
	}
	if secret != nil {
		t.Error("LoadSecret should return nil (noop implementation)")
	}

	// Test LoadAllSecrets (inherited from Store)
	secrets, loadAllSecretsErr := ds.LoadAllSecrets(ctx)
	if loadAllSecretsErr != nil {
		t.Errorf("LoadAllSecrets should not return error: %v", loadAllSecretsErr)
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
	storeSecretErr := ds.StoreSecret(ctx, "test/path", testSecret)
	if storeSecretErr != nil {
		t.Errorf("StoreSecret should not return error: %v", storeSecretErr)
	}

	// Test LoadPolicy (inherited from Store)
	policy, loadPolicyErr := ds.LoadPolicy(ctx, "test-policy-id")
	if loadPolicyErr != nil {
		t.Errorf("LoadPolicy should not return error: %v", loadPolicyErr)
	}
	if policy != nil {
		t.Error("LoadPolicy should return nil (noop implementation)")
	}

	// Test LoadAllPolicies (inherited from Store)
	policies, loadAllPoliciesErr := ds.LoadAllPolicies(ctx)
	if loadAllPoliciesErr != nil {
		t.Errorf("LoadAllPolicies should not return error: %v", loadAllPoliciesErr)
	}
	if policies != nil {
		t.Error("LoadAllPolicies should return nil (noop implementation)")
	}

	// Test StorePolicy (inherited from Store)
	testPolicy := data.Policy{
		ID:              "test-policy",
		Name:            "test policy",
		SPIFFEIDPattern: "^spiffe://example\\.org/test$",
		PathPattern:     "^test/.*$",
		Permissions:     []data.PolicyPermission{data.PermissionRead},
	}
	storePolicyErr := ds.StorePolicy(ctx, testPolicy)
	if storePolicyErr != nil {
		t.Errorf("StorePolicy should not return error: %v", storePolicyErr)
	}

	// Test DeletePolicy (inherited from Store)
	deletePolicyErr := ds.DeletePolicy(ctx, "test-policy-id")
	if deletePolicyErr != nil {
		t.Errorf("DeletePolicy should not return error: %v", deletePolicyErr)
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
	if _, randErr := rand.Read(rootKey[:]); randErr != nil {
		t.Fatalf("Failed to generate random key: %v", randErr)
	}

	ds, newErr := New(rootKey)
	if newErr != nil {
		t.Fatalf("Failed to create Store: %v", newErr)
	}

	liteStore, ok := ds.(*Store)
	if !ok {
		t.Fatal("Expected Store type")
		return
	}
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
	if _, randErr := rand.Read(nonce); randErr != nil {
		t.Fatalf("Failed to generate nonce: %v", randErr)
	}

	// Encrypt
	ciphertext := cipher.Seal(nil, nonce, plaintext, nil)
	if len(ciphertext) == 0 {
		t.Error("Encryption should produce non-empty ciphertext")
	}

	// Decrypt
	decrypted, decryptErr := cipher.Open(nil, nonce, ciphertext, nil)
	if decryptErr != nil {
		t.Errorf("Decryption failed: %v", decryptErr)
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

	if _, randErr := rand.Read(key1[:]); randErr != nil {
		t.Fatalf("Failed to generate first key: %v", randErr)
	}
	if _, randErr := rand.Read(key2[:]); randErr != nil {
		t.Fatalf("Failed to generate second key: %v", randErr)
	}

	// Ensure keys are different
	if *key1 == *key2 {
		key2[0] = ^key1[0] // Make them different
	}

	// Create two DataStores
	ds1, newErr1 := New(key1)
	ds2, newErr2 := New(key2)

	if newErr1 != nil || newErr2 != nil {
		t.Fatalf("Failed to create DataStores: %v, %v", newErr1, newErr2)
	}

	cipher1 := ds1.GetCipher()
	cipher2 := ds2.GetCipher()

	// Test that they produce different encrypted output for the same input
	plaintext := []byte("test data")
	nonce := make([]byte, cipher1.NonceSize())
	if _, randErr := rand.Read(nonce); randErr != nil {
		t.Fatalf("Failed to generate nonce: %v", randErr)
	}

	ciphertext1 := cipher1.Seal(nil, nonce, plaintext, nil)
	ciphertext2 := cipher2.Seal(nil, nonce, plaintext, nil)

	// They should produce different ciphertext (different keys)
	if len(ciphertext1) == len(ciphertext2) &&
		string(ciphertext1) == string(ciphertext2) {
		t.Error("Different keys should produce different ciphertext")
	}

	// Verify cipher1 cannot decrypt cipher2's output
	_, openErr := cipher1.Open(nil, nonce, ciphertext2, nil)
	if openErr == nil {
		t.Error(
			"Cipher with different key should not be able to decrypt ciphertext",
		)
	}
}

func TestDataStore_EmbeddedNoopStore(t *testing.T) {
	// Test that Store properly embeds Store
	rootKey := &[crypto.AES256KeySize]byte{}
	if _, randErr := rand.Read(rootKey[:]); randErr != nil {
		t.Fatalf("Failed to generate random key: %v", randErr)
	}

	ds, newErr := New(rootKey)
	if newErr != nil {
		t.Fatalf("Failed to create Store: %v", newErr)
	}

	liteStore, ok := ds.(*Store)
	if !ok {
		t.Fatal("Expected Store type")
		return
	}

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

	methods := []func() *sdkErrors.SDKError{
		func() *sdkErrors.SDKError { return liteStore.Initialize(ctx) },
		func() *sdkErrors.SDKError { return liteStore.Close(ctx) },
		func() *sdkErrors.SDKError {
			return liteStore.StoreSecret(ctx, "path", testSecret)
		},
		func() *sdkErrors.SDKError {
			return liteStore.StorePolicy(ctx, testPolicy)
		},
		func() *sdkErrors.SDKError { return liteStore.DeletePolicy(ctx, "id") },
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
	if _, randErr := rand.Read(rootKey[:]); randErr != nil {
		t.Fatalf("Failed to generate random key: %v", randErr)
	}

	ds, newErr := New(rootKey)
	if newErr != nil {
		t.Fatalf("Failed to create Store: %v", newErr)
	}

	cipher := ds.GetCipher()

	// GCM should have specific properties
	expectedNonceSize := 12 // Standard GCM nonce size
	expectedOverhead := 16  // GCM authentication tag size

	if cipher.NonceSize() != expectedNonceSize {
		t.Errorf(
			"Expected GCM nonce size %d, got %d",
			expectedNonceSize,
			cipher.NonceSize(),
		)
	}

	if cipher.Overhead() != expectedOverhead {
		t.Errorf(
			"Expected GCM overhead %d, got %d", expectedOverhead, cipher.Overhead(),
		)
	}
}

func TestDataStore_MemoryManagement(t *testing.T) {
	// Test that multiple Store instances can coexist
	keys := make([]*[crypto.AES256KeySize]byte, 5)
	dss := make([]backend.Backend, 5)

	// Create multiple instances
	for i := 0; i < 5; i++ {
		keys[i] = &[crypto.AES256KeySize]byte{}
		if _, randErr := rand.Read(keys[i][:]); randErr != nil {
			t.Fatalf("Failed to generate key %d: %v", i, randErr)
		}

		ds, newErr := New(keys[i])
		if newErr != nil {
			t.Fatalf("Failed to create Store %d: %v", i, newErr)
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
