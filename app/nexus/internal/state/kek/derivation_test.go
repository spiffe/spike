//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"bytes"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
)

func TestDeriveKEK(t *testing.T) {
	// Generate a test RMK
	rmk := &[crypto.AES256KeySize]byte{}
	for i := range rmk {
		rmk[i] = byte(i)
	}

	// Generate a salt
	salt, err := GenerateKEKSalt()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	// Create metadata
	meta := &Metadata{
		ID:         "test-kek-v1",
		Version:    1,
		Salt:       salt,
		RMKVersion: 1,
		CreatedAt:  time.Now(),
		WrapsCount: 0,
		Status:     KekStatusActive,
	}

	// Derive KEK
	kek1, err := DeriveKEK(rmk, meta)
	if err != nil {
		t.Fatalf("Failed to derive KEK: %v", err)
	}

	// Derive again - should get same result (deterministic)
	kek2, err := DeriveKEK(rmk, meta)
	if err != nil {
		t.Fatalf("Failed to derive KEK second time: %v", err)
	}

	if !bytes.Equal(kek1[:], kek2[:]) {
		t.Error("KEK derivation is not deterministic")
	}

	// Different salt should produce different KEK
	salt2, _ := GenerateKEKSalt()
	meta2 := &Metadata{
		ID:         "test-kek-v2",
		Version:    2,
		Salt:       salt2,
		RMKVersion: 1,
		CreatedAt:  time.Now(),
		WrapsCount: 0,
		Status:     KekStatusActive,
	}

	kek3, err := DeriveKEK(rmk, meta2)
	if err != nil {
		t.Fatalf("Failed to derive KEK with different salt: %v", err)
	}

	if bytes.Equal(kek1[:], kek3[:]) {
		t.Error("Different salts produced same KEK")
	}
}

func TestDeriveKEKWithNilInputs(t *testing.T) {
	rmk := &[crypto.AES256KeySize]byte{}
	salt, _ := GenerateKEKSalt()
	meta := &Metadata{
		ID:         "test",
		Salt:       salt,
		RMKVersion: 1,
	}

	// Test nil RMK
	_, err := DeriveKEK(nil, meta)
	if err == nil {
		t.Error("Expected error with nil RMK")
	}

	// Test nil metadata
	_, err = DeriveKEK(rmk, nil)
	if err == nil {
		t.Error("Expected error with nil metadata")
	}
}

func TestGenerateKEKSalt(t *testing.T) {
	salt1, err := GenerateKEKSalt()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	if len(salt1) != KekSaltSize {
		t.Errorf("Expected salt size %d, got %d", KekSaltSize, len(salt1))
	}

	// Generate another salt - should be different
	salt2, err := GenerateKEKSalt()
	if err != nil {
		t.Fatalf("Failed to generate second salt: %v", err)
	}

	if bytes.Equal(salt1[:], salt2[:]) {
		t.Error("Two generated salts are identical (should be random)")
	}
}

func TestGenerateDEK(t *testing.T) {
	dek1, err := GenerateDEK()
	if err != nil {
		t.Fatalf("Failed to generate DEK: %v", err)
	}

	if dek1 == nil {
		t.Fatal("Generated DEK is nil")
	}

	// Generate another DEK - should be different
	dek2, err := GenerateDEK()
	if err != nil {
		t.Fatalf("Failed to generate second DEK: %v", err)
	}

	if bytes.Equal(dek1[:], dek2[:]) {
		t.Error("Two generated DEKs are identical (should be random)")
	}
}

func TestKEKDomainSeparation(t *testing.T) {
	rmk := &[crypto.AES256KeySize]byte{}
	for i := range rmk {
		rmk[i] = byte(i)
	}

	// Use same salt but different KEK IDs
	salt, _ := GenerateKEKSalt()

	meta1 := &Metadata{
		ID:         "kek-v1-2025-01",
		Version:    1,
		Salt:       salt,
		RMKVersion: 1,
	}

	meta2 := &Metadata{
		ID:         "kek-v2-2025-02",
		Version:    2,
		Salt:       salt,
		RMKVersion: 1,
	}

	kek1, err := DeriveKEK(rmk, meta1)
	if err != nil {
		t.Fatalf("Failed to derive KEK 1: %v", err)
	}

	kek2, err := DeriveKEK(rmk, meta2)
	if err != nil {
		t.Fatalf("Failed to derive KEK 2: %v", err)
	}

	// Even with same salt, different IDs should produce different KEKs
	// due to domain separation
	if bytes.Equal(kek1[:], kek2[:]) {
		t.Error("Same salt with different KEK IDs produced identical KEKs")
	}
}
