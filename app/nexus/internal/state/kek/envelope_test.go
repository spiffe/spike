//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"bytes"
	"testing"

	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestWrapUnwrapDEK(t *testing.T) {
	// Generate KEK and DEK
	kek := &[crypto.AES256KeySize]byte{}
	for i := range kek {
		kek[i] = byte(i * 2)
	}

	dek, err := GenerateDEK()
	if err != nil {
		t.Fatalf("Failed to generate DEK: %v", err)
	}

	kekID := "test-kek-v1"

	// Wrap DEK
	wrapResult, err := WrapDEK(dek, kek, kekID, nil)
	if err != nil {
		t.Fatalf("Failed to wrap DEK: %v", err)
	}

	if wrapResult.KekID != kekID {
		t.Errorf("Expected KEK ID %s, got %s", kekID, wrapResult.KekID)
	}

	if len(wrapResult.WrappedDEK) == 0 {
		t.Error("Wrapped DEK is empty")
	}

	if len(wrapResult.Nonce) == 0 {
		t.Error("Nonce is empty")
	}

	// Unwrap DEK
	unwrapResult, err := UnwrapDEK(
		wrapResult.WrappedDEK,
		kek,
		kekID,
		wrapResult.Nonce,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to unwrap DEK: %v", err)
	}

	// Verify unwrapped DEK matches original
	if !bytes.Equal(unwrapResult.DEK[:], dek[:]) {
		t.Error("Unwrapped DEK does not match original")
	}

	if unwrapResult.KekID != kekID {
		t.Errorf("Expected KEK ID %s, got %s", kekID, unwrapResult.KekID)
	}
}

func TestWrapUnwrapDEKWithAAD(t *testing.T) {
	kek := &[crypto.AES256KeySize]byte{}
	for i := range kek {
		kek[i] = byte(i * 2)
	}

	dek, _ := GenerateDEK()
	kekID := "test-kek-v1"
	aad := []byte("additional-authenticated-data")

	// Wrap with AAD
	wrapResult, err := WrapDEK(dek, kek, kekID, aad)
	if err != nil {
		t.Fatalf("Failed to wrap DEK with AAD: %v", err)
	}

	// Unwrap with correct AAD
	unwrapResult, err := UnwrapDEK(
		wrapResult.WrappedDEK,
		kek,
		kekID,
		wrapResult.Nonce,
		aad,
	)
	if err != nil {
		t.Fatalf("Failed to unwrap DEK with AAD: %v", err)
	}

	if !bytes.Equal(unwrapResult.DEK[:], dek[:]) {
		t.Error("Unwrapped DEK does not match original")
	}

	// Unwrap with wrong AAD should fail
	wrongAAD := []byte("wrong-aad")
	_, err = UnwrapDEK(
		wrapResult.WrappedDEK,
		kek,
		kekID,
		wrapResult.Nonce,
		wrongAAD,
	)
	if err == nil {
		t.Error("Expected error when unwrapping with wrong AAD")
	}
}

func TestRewrapDEK(t *testing.T) {
	// Generate two KEKs (old and new)
	oldKek := &[crypto.AES256KeySize]byte{}
	for i := range oldKek {
		oldKek[i] = byte(i)
	}

	newKek := &[crypto.AES256KeySize]byte{}
	for i := range newKek {
		newKek[i] = byte(i * 3)
	}

	// Generate DEK
	dek, _ := GenerateDEK()

	oldKekID := "kek-v1"
	newKekID := "kek-v2"

	// Wrap with old KEK
	oldWrap, err := WrapDEK(dek, oldKek, oldKekID, nil)
	if err != nil {
		t.Fatalf("Failed to wrap with old KEK: %v", err)
	}

	// Rewrap to new KEK
	newWrap, err := RewrapDEK(
		oldWrap.WrappedDEK,
		oldKek,
		oldKekID,
		oldWrap.Nonce,
		nil,
		newKek,
		newKekID,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to rewrap DEK: %v", err)
	}

	if newWrap.KekID != newKekID {
		t.Errorf("Expected KEK ID %s, got %s", newKekID, newWrap.KekID)
	}

	// Unwrap with new KEK
	unwrapResult, err := UnwrapDEK(
		newWrap.WrappedDEK,
		newKek,
		newKekID,
		newWrap.Nonce,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to unwrap with new KEK: %v", err)
	}

	// Verify DEK is still the same
	if !bytes.Equal(unwrapResult.DEK[:], dek[:]) {
		t.Error("Rewrapped DEK does not match original")
	}
}

func TestEncryptDecryptWithDEK(t *testing.T) {
	dek, _ := GenerateDEK()
	plaintext := []byte("test secret data")

	// Encrypt
	ciphertext, nonce, err := EncryptWithDEK(plaintext, dek)
	if err != nil {
		t.Fatalf("Failed to encrypt with DEK: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Error("Ciphertext is empty")
	}

	if len(nonce) == 0 {
		t.Error("Nonce is empty")
	}

	// Decrypt
	decrypted, err := DecryptWithDEK(ciphertext, dek, nonce)
	if err != nil {
		t.Fatalf("Failed to decrypt with DEK: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted data does not match. Expected %s, got %s",
			string(plaintext), string(decrypted))
	}
}

func TestEncryptDecryptWithWrongDEK(t *testing.T) {
	dek1, _ := GenerateDEK()
	dek2, _ := GenerateDEK()

	plaintext := []byte("test secret data")

	// Encrypt with DEK1
	ciphertext, nonce, err := EncryptWithDEK(plaintext, dek1)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Try to decrypt with DEK2 - should fail
	_, err = DecryptWithDEK(ciphertext, dek2, nonce)
	if err == nil {
		t.Error("Expected error when decrypting with wrong DEK")
	}
}

func TestWrapDEKWithNilInputs(t *testing.T) {
	kek := &[crypto.AES256KeySize]byte{}
	dek, _ := GenerateDEK()

	// Nil DEK
	_, err := WrapDEK(nil, kek, "test", nil)
	if err == nil {
		t.Error("Expected error with nil DEK")
	}

	// Nil KEK
	_, err = WrapDEK(dek, nil, "test", nil)
	if err == nil {
		t.Error("Expected error with nil KEK")
	}
}

func TestUnwrapDEKWithNilKEK(t *testing.T) {
	wrappedDEK := []byte("fake-wrapped-dek")
	nonce := []byte("fake-nonce")

	_, err := UnwrapDEK(wrappedDEK, nil, "test", nonce, nil)
	if err == nil {
		t.Error("Expected error with nil KEK")
	}
}

func TestBuildSecretMetadata(t *testing.T) {
	wrapResult := &WrapResult{
		KekID:      "test-kek-v1",
		WrappedDEK: []byte("wrapped-dek-data"),
		Nonce:      []byte("nonce-data"),
		AAD:        []byte("aad-data"),
		AEADAlg:    AEADAlgAES256GCM,
	}

	meta := BuildSecretMetadata(wrapResult)

	if meta.KekID != wrapResult.KekID {
		t.Errorf("KEK ID mismatch: expected %s, got %s", wrapResult.KekID, meta.KekID)
	}

	if !bytes.Equal(meta.WrappedDEK, wrapResult.WrappedDEK) {
		t.Error("Wrapped DEK mismatch")
	}

	if !bytes.Equal(meta.Nonce, wrapResult.Nonce) {
		t.Error("Nonce mismatch")
	}

	if meta.AEADAlg != wrapResult.AEADAlg {
		t.Errorf("AEAD alg mismatch: expected %s, got %s", wrapResult.AEADAlg, meta.AEADAlg)
	}

	if meta.RewrappedAt == nil {
		t.Error("RewrappedAt should not be nil")
	}
}
