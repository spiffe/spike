//  \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"

	"github.com/spiffe/spike/pkg/store"
)

// StoreKeyRecoveryInfo stores encrypted key recovery information in the
// database. It marshals the provided KeyRecoveryData to JSON, encrypts it, and
// stores it along with the encryption nonce. The data is stored with a fixed
// Id as defined in the database schema.
//
// The method is thread-safe, using a mutex to prevent concurrent access.
//
// Returns an error if marshaling, encryption, or database operations fail.
func (s *DataStore) StoreKeyRecoveryInfo(
	ctx context.Context, data store.KeyRecoveryData,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Marshal the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal key recovery data: %w", err)
	}

	// Encrypt the JSON data
	encrypted, nonce, err := s.encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt key recovery data: %w", err)
	}

	// Store in database with id=1 (as per DDL)
	_, err = s.db.ExecContext(ctx, ddl.QueryUpsertKeyRecoveryInfo,
		nonce, encrypted)
	if err != nil {
		return fmt.Errorf("failed to store key recovery data: %w", err)
	}

	return nil
}

// LoadKeyRecoveryInfo retrieves and decrypts key recovery information from the
// database. If no recovery information exists, it returns (nil, nil).
//
// The method is thread-safe, using a read lock to allow concurrent reads.
//
// Returns:
//   - The decrypted KeyRecoveryData if found and successfully decrypted
//   - nil, nil if no recovery data exists
//   - An error if database operations, decryption, or JSON unmarshaling fail
func (s *DataStore) LoadKeyRecoveryInfo(
	ctx context.Context,
) (*store.KeyRecoveryData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var (
		nonce         []byte
		encryptedData []byte
	)

	err := s.db.QueryRowContext(
		ctx,
		ddl.QueryLoadKeyRecoveryInfo,
	).Scan(&nonce, &encryptedData)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load key recovery data: %w", err)
	}

	// Decrypt the data
	decrypted, err := s.decrypt(encryptedData, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key recovery data: %w", err)
	}

	// Unmarshal into KeyRecoveryData struct
	var data store.KeyRecoveryData
	if err := json.Unmarshal(decrypted, &data); err != nil {
		return nil, fmt.Errorf(
			"failed to unmarshal key recovery data: %w", err,
		)
	}

	return &data, nil
}
