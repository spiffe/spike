//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Standalone test for KEK functionality without requiring full SPIKE infrastructure
package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
	_ "github.com/mattn/go-sqlite3"

	"github.com/spiffe/spike/app/nexus/internal/state/kek"
)

type testStorage struct {
	db *sql.DB
}

func (s *testStorage) StoreKEKMetadata(metadata *kek.Metadata) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO kek_metadata 
		(kek_id, version, salt, rmk_version, created_at, wraps_count, status, retired_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, metadata.ID, metadata.Version, metadata.Salt[:], metadata.RMKVersion,
		metadata.CreatedAt.Unix(), metadata.WrapsCount, string(metadata.Status),
		nil)
	return err
}

func (s *testStorage) LoadKEKMetadata(kekID string) (*kek.Metadata, error) {
	var meta kek.Metadata
	var salt []byte
	var status string
	
	err := s.db.QueryRow(`
		SELECT kek_id, version, salt, rmk_version, created_at, wraps_count, status
		FROM kek_metadata WHERE kek_id = ?
	`, kekID).Scan(&meta.ID, &meta.Version, &salt, &meta.RMKVersion,
		&meta.CreatedAt, &meta.WrapsCount, &status)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	copy(meta.Salt[:], salt)
	meta.Status = kek.KekStatus(status)
	return &meta, nil
}

func (s *testStorage) ListKEKMetadata() ([]*kek.Metadata, error) {
	rows, err := s.db.Query(`
		SELECT kek_id, version, salt, rmk_version, created_at, wraps_count, status
		FROM kek_metadata
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var result []*kek.Metadata
	for rows.Next() {
		var meta kek.Metadata
		var salt []byte
		var status string
		
		err := rows.Scan(&meta.ID, &meta.Version, &salt, &meta.RMKVersion,
			&meta.CreatedAt, &meta.WrapsCount, &status)
		if err != nil {
			return nil, err
		}
		
		copy(meta.Salt[:], salt)
		meta.Status = kek.KekStatus(status)
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

func (s *testStorage) UpdateKEKStatus(kekID string, status kek.KekStatus, retiredAt *time.Time) error {
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

func main() {
	fmt.Println("========================================")
	fmt.Println("SPIKE KEK Standalone Test")
	fmt.Println("========================================")
	fmt.Println()
	
	passed := 0
	failed := 0
	
	// Setup
	fmt.Println("[1/8] Setting up test database...")
	db, err := setupDatabase()
	if err != nil {
		fmt.Printf("Failed to setup database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	
	storage := &testStorage{db: db}
	rmk := generateRMK()
	policy := kek.DefaultRotationPolicy()
	
	fmt.Println("✓ Database setup complete")
	fmt.Println()
	
	// Test 1: Create KEK Manager
	fmt.Println("[2/8] Creating KEK manager...")
	manager, err := kek.NewManager(rmk, 1, policy, storage)
	if err != nil {
		fmt.Printf("✗ Failed to create manager: %v\n", err)
		failed++
	} else {
		fmt.Println("✓ KEK manager created successfully")
		passed++
	}
	fmt.Println()
	
	// Test 2: Get Current KEK
	fmt.Println("[3/8] Getting current KEK...")
	currentKEKID := manager.GetCurrentKEKID()
	if currentKEKID == "" {
		fmt.Println("✗ No current KEK ID")
		failed++
	} else {
		fmt.Printf("✓ Current KEK ID: %s\n", currentKEKID)
		passed++
	}
	fmt.Println()
	
	// Test 3: Wrap and Unwrap DEK
	fmt.Println("[4/8] Testing DEK wrap/unwrap...")
	dek, err := kek.GenerateDEK()
	if err != nil {
		fmt.Printf("✗ Failed to generate DEK: %v\n", err)
		failed++
	} else {
		currentKEK, err := manager.GetKEK(currentKEKID)
		if err != nil {
			fmt.Printf("✗ Failed to get KEK: %v\n", err)
			failed++
		} else {
			wrapResult, err := kek.WrapDEK(dek, currentKEK, currentKEKID, nil)
			if err != nil {
				fmt.Printf("✗ Failed to wrap DEK: %v\n", err)
				failed++
			} else {
				unwrapResult, err := kek.UnwrapDEK(
					wrapResult.WrappedDEK,
					currentKEK,
					currentKEKID,
					wrapResult.Nonce,
					nil,
				)
				if err != nil {
					fmt.Printf("✗ Failed to unwrap DEK: %v\n", err)
					failed++
				} else {
					match := true
					for i := 0; i < crypto.AES256KeySize; i++ {
						if dek[i] != unwrapResult.DEK[i] {
							match = false
							break
						}
					}
					if match {
						fmt.Println("✓ DEK wrap/unwrap successful")
						passed++
					} else {
						fmt.Println("✗ DEK mismatch after unwrap")
						failed++
					}
				}
			}
		}
	}
	fmt.Println()
	
	// Test 4: KEK Rotation
	fmt.Println("[5/8] Testing KEK rotation...")
	err = manager.RotateKEK()
	if err != nil {
		fmt.Printf("✗ Failed to rotate KEK: %v\n", err)
		failed++
	} else {
		newKEKID := manager.GetCurrentKEKID()
		if newKEKID == currentKEKID {
			fmt.Println("✗ KEK ID did not change after rotation")
			failed++
		} else {
			fmt.Printf("✓ KEK rotated: %s -> %s\n", currentKEKID, newKEKID)
			passed++
		}
	}
	fmt.Println()
	
	// Test 5: List KEKs
	fmt.Println("[6/8] Listing all KEKs...")
	keks, err := manager.ListAllKEKs()
	if err != nil {
		fmt.Printf("✗ Failed to list KEKs: %v\n", err)
		failed++
	} else {
		fmt.Printf("✓ Found %d KEKs:\n", len(keks))
		for _, k := range keks {
			fmt.Printf("  - %s (status: %s, wraps: %d)\n", k.ID, k.Status, k.WrapsCount)
		}
		passed++
	}
	fmt.Println()
	
	// Test 6: Increment Wraps Count
	fmt.Println("[7/8] Testing wraps count increment...")
	newKEKID := manager.GetCurrentKEKID()
	meta, _ := manager.GetMetadata(newKEKID)
	initialCount := meta.WrapsCount
	
	err = manager.IncrementWrapsCount(newKEKID)
	if err != nil {
		fmt.Printf("✗ Failed to increment wraps count: %v\n", err)
		failed++
	} else {
		meta, _ = manager.GetMetadata(newKEKID)
		if meta.WrapsCount == initialCount+1 {
			fmt.Printf("✓ Wraps count incremented: %d -> %d\n", initialCount, meta.WrapsCount)
			passed++
		} else {
			fmt.Printf("✗ Wraps count not incremented correctly: %d -> %d\n", initialCount, meta.WrapsCount)
			failed++
		}
	}
	fmt.Println()
	
	// Test 7: Secret Encryption/Decryption
	fmt.Println("[8/8] Testing secret encryption with DEK...")
	secretData := []byte("This is a super secret message!")
	
	encrypted, nonce, err := kek.EncryptWithDEK(secretData, dek)
	if err != nil {
		fmt.Printf("✗ Failed to encrypt: %v\n", err)
		failed++
	} else {
		decrypted, err := kek.DecryptWithDEK(encrypted, dek, nonce)
		if err != nil {
			fmt.Printf("✗ Failed to decrypt: %v\n", err)
			failed++
		} else {
			if string(decrypted) == string(secretData) {
				fmt.Println("✓ Secret encryption/decryption successful")
				passed++
			} else {
				fmt.Println("✗ Decrypted data mismatch")
				failed++
			}
		}
	}
	fmt.Println()
	
	// Summary
	fmt.Println("========================================")
	fmt.Println("Test Summary")
	fmt.Println("========================================")
	fmt.Printf("Passed: %d\n", passed)
	if failed > 0 {
		fmt.Printf("Failed: %d\n", failed)
		os.Exit(1)
	} else {
		fmt.Println("Failed: 0")
		fmt.Println()
		fmt.Println("✓ All tests passed!")
	}
}

