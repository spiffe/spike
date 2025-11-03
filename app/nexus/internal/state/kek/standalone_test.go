//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package kek

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
	_ "github.com/mattn/go-sqlite3"
)

type testStorage struct {
	db *sql.DB
}

func (s *testStorage) StoreKEKMetadata(metadata *Metadata) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO kek_metadata 
		(kek_id, version, salt, rmk_version, created_at, wraps_count, status, retired_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, metadata.ID, metadata.Version, metadata.Salt[:], metadata.RMKVersion,
		metadata.CreatedAt.Unix(), metadata.WrapsCount, string(metadata.Status),
		nil)
	return err
}

func (s *testStorage) LoadKEKMetadata(kekID string) (*Metadata, error) {
	var meta Metadata
	var salt []byte
	var status string
	var createdAtUnix int64
	
	err := s.db.QueryRow(`
		SELECT kek_id, version, salt, rmk_version, created_at, wraps_count, status
		FROM kek_metadata WHERE kek_id = ?
	`, kekID).Scan(&meta.ID, &meta.Version, &salt, &meta.RMKVersion,
		&createdAtUnix, &meta.WrapsCount, &status)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	copy(meta.Salt[:], salt)
	meta.Status = KekStatus(status)
	meta.CreatedAt = time.Unix(createdAtUnix, 0)
	return &meta, nil
}

func (s *testStorage) ListKEKMetadata() ([]*Metadata, error) {
	rows, err := s.db.Query(`
		SELECT kek_id, version, salt, rmk_version, created_at, wraps_count, status
		FROM kek_metadata
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var result []*Metadata
	for rows.Next() {
		var meta Metadata
		var salt []byte
		var status string
		var createdAtUnix int64
		
		err := rows.Scan(&meta.ID, &meta.Version, &salt, &meta.RMKVersion,
			&createdAtUnix, &meta.WrapsCount, &status)
		if err != nil {
			return nil, err
		}
		
		copy(meta.Salt[:], salt)
		meta.Status = KekStatus(status)
		meta.CreatedAt = time.Unix(createdAtUnix, 0)
		result = append(result, &meta)
	}
	
	return result, nil
}

func (s *testStorage) UpdateKEKWrapsCount(kekID string, delta int64) error {
	_, err := s.db.Exec(`
		UPDATE kek_metadata SET wraps_count = wraps_count + ? WHERE kek_id = ?
	`, delta, kekID)
	return err
}

func (s *testStorage) UpdateKEKStatus(kekID string, status KekStatus, retiredAt *time.Time) error {
	_, err := s.db.Exec(`
		UPDATE kek_metadata SET status = ?, retired_at = ? WHERE kek_id = ?
	`, string(status), retiredAt, kekID)
	return err
}

func setupDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	
	_, err = db.Exec(`
		CREATE TABLE kek_metadata (
			kek_id TEXT PRIMARY KEY,
			version INTEGER NOT NULL,
			salt BLOB NOT NULL,
			rmk_version INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			wraps_count INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			retired_at INTEGER
		)
	`)
	if err != nil {
		return nil, err
	}
	
	return db, nil
}

func generateRMK() *[crypto.AES256KeySize]byte {
	rmk := new([crypto.AES256KeySize]byte)
	if _, err := rand.Read(rmk[:]); err != nil {
		panic(err)
	}
	return rmk
}

func TestStandaloneKEK(t *testing.T) {
	fmt.Println("========================================")
	fmt.Println("SPIKE KEK Standalone Test")
	fmt.Println("========================================")
	fmt.Println()
	
	// Setup
	fmt.Println("[1/8] Setting up test database...")
	db, err := setupDatabase()
	if err != nil {
		t.Fatalf("Failed to setup database: %v", err)
	}
	defer db.Close()
	
	storage := &testStorage{db: db}
	rmk := generateRMK()
	policy := DefaultRotationPolicy()
	
	fmt.Println("✓ Database setup complete")
	fmt.Println()
	
	// Test 1: Create KEK Manager
	fmt.Println("[2/8] Creating KEK manager...")
	manager, err := NewManager(rmk, 1, policy, storage)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}
	fmt.Println("✓ KEK manager created successfully")
	fmt.Println()
	
	// Test 2: Get Current KEK
	fmt.Println("[3/8] Getting current KEK...")
	currentKEKID := manager.GetCurrentKEKID()
	if currentKEKID == "" {
		t.Fatal("No current KEK ID")
	}
	fmt.Printf("✓ Current KEK ID: %s\n", currentKEKID)
	fmt.Println()
	
	// Test 3: Wrap and Unwrap DEK
	fmt.Println("[4/8] Testing DEK wrap/unwrap...")
	dek, err := GenerateDEK()
	if err != nil {
		t.Fatalf("Failed to generate DEK: %v", err)
	}
	
	currentKEK, err := manager.GetKEK(currentKEKID)
	if err != nil {
		t.Fatalf("Failed to get KEK: %v", err)
	}
	
	wrapResult, err := WrapDEK(dek, currentKEK, currentKEKID, nil)
	if err != nil {
		t.Fatalf("Failed to wrap DEK: %v", err)
	}
	
	unwrapResult, err := UnwrapDEK(
		wrapResult.WrappedDEK,
		currentKEK,
		currentKEKID,
		wrapResult.Nonce,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to unwrap DEK: %v", err)
	}
	
	for i := 0; i < crypto.AES256KeySize; i++ {
		if dek[i] != unwrapResult.DEK[i] {
			t.Fatal("DEK mismatch after unwrap")
		}
	}
	fmt.Println("✓ DEK wrap/unwrap successful")
	fmt.Println()
	
	// Test 4: KEK Rotation
	fmt.Println("[5/8] Testing KEK rotation...")
	err = manager.RotateKEK()
	if err != nil {
		t.Fatalf("Failed to rotate KEK: %v", err)
	}
	
	newKEKID := manager.GetCurrentKEKID()
	if newKEKID == currentKEKID {
		t.Fatal("KEK ID did not change after rotation")
	}
	fmt.Printf("✓ KEK rotated: %s -> %s\n", currentKEKID, newKEKID)
	fmt.Println()
	
	// Test 5: List KEKs
	fmt.Println("[6/8] Listing all KEKs...")
	keks, err := manager.ListAllKEKs()
	if err != nil {
		t.Fatalf("Failed to list KEKs: %v", err)
	}
	fmt.Printf("✓ Found %d KEKs:\n", len(keks))
	for _, k := range keks {
		fmt.Printf("  - %s (status: %s, wraps: %d)\n", k.ID, k.Status, k.WrapsCount)
	}
	fmt.Println()
	
	// Test 6: Increment Wraps Count
	fmt.Println("[7/8] Testing wraps count increment...")
	
	// Get initial count directly from storage
	initialMeta, err := storage.LoadKEKMetadata(newKEKID)
	if err != nil {
		t.Fatalf("Failed to load metadata from storage: %v", err)
	}
	initialCount := initialMeta.WrapsCount
	
	err = manager.IncrementWrapsCount(newKEKID)
	if err != nil {
		t.Fatalf("Failed to increment wraps count: %v", err)
	}
	
	// Get updated count directly from storage
	updatedMeta, err := storage.LoadKEKMetadata(newKEKID)
	if err != nil {
		t.Fatalf("Failed to load updated metadata from storage: %v", err)
	}
	if updatedMeta.WrapsCount != initialCount+1 {
		t.Fatalf("Wraps count not incremented correctly: %d -> %d", initialCount, updatedMeta.WrapsCount)
	}
	fmt.Printf("✓ Wraps count incremented: %d -> %d\n", initialCount, updatedMeta.WrapsCount)
	fmt.Println()
	
	// Test 7: Secret Encryption/Decryption
	fmt.Println("[8/8] Testing secret encryption with DEK...")
	secretData := []byte("This is a super secret message!")
	
	encrypted, nonce, err := EncryptWithDEK(secretData, dek)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}
	
	decrypted, err := DecryptWithDEK(encrypted, dek, nonce)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}
	
	if string(decrypted) != string(secretData) {
		t.Fatal("Decrypted data mismatch")
	}
	fmt.Println("✓ Secret encryption/decryption successful")
	fmt.Println()
	
	// Summary
	fmt.Println("========================================")
	fmt.Println("✓ All tests passed!")
	fmt.Println("========================================")
}

