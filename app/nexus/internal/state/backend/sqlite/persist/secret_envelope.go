//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
	"github.com/spiffe/spike/app/nexus/internal/state/kek"
)

// SetKEKManager sets the KEK manager for envelope encryption
// If set, the DataStore will use envelope encryption for secrets
func (s *DataStore) SetKEKManager(manager *kek.Manager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.kekManager = manager
}

// StoreSecretWithEnvelope stores a secret using envelope encryption
// This uses the KEK manager to wrap a per-secret DEK
func (s *DataStore) StoreSecretWithEnvelope(
	ctx context.Context, path string, secret kv.Value,
) error {
	const fName = "StoreSecretWithEnvelope"
	if ctx == nil {
		log.FatalLn(fName, "message", "nil context")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.kekManager == nil {
		return fmt.Errorf("%s: KEK manager not initialized", fName)
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Log().Error(fName, "message", "failed to rollback transaction", "err", rollbackErr.Error())
			}
		}
	}(tx)

	// Update metadata
	_, err = tx.ExecContext(ctx, ddl.QueryUpdateSecretMetadata,
		path, secret.Metadata.CurrentVersion, secret.Metadata.OldestVersion,
		secret.Metadata.CreatedTime, secret.Metadata.UpdatedTime, secret.Metadata.MaxVersions)
	if err != nil {
		return fmt.Errorf("failed to update secret metadata: %w", err)
	}

	// Get current KEK
	currentKekID := s.kekManager.GetCurrentKEKID()
	currentKek, err := s.kekManager.GetKEK(currentKekID)
	if err != nil {
		return fmt.Errorf("failed to get current KEK: %w", err)
	}

	// Update versions with envelope encryption
	for version, sv := range secret.Versions {
		// Marshal secret data
		data, err := json.Marshal(sv.Data)
		if err != nil {
			return fmt.Errorf("failed to marshal secret values: %w", err)
		}

		// Generate a random DEK for this secret version
		dek, err := kek.GenerateDEK()
		if err != nil {
			return fmt.Errorf("failed to generate DEK: %w", err)
		}

		// Encrypt data with DEK
		encryptedData, dataNonce, err := kek.EncryptWithDEK(data, dek)
		if err != nil {
			return fmt.Errorf("failed to encrypt with DEK: %w", err)
		}

		// Wrap DEK with KEK
		wrapResult, err := kek.WrapDEK(dek, currentKek, currentKekID, nil)
		if err != nil {
			return fmt.Errorf("failed to wrap DEK: %w", err)
		}

		// Store secret version with envelope metadata
		rewrappedAt := time.Now()
		_, err = tx.ExecContext(ctx, ddl.QueryUpsertSecret,
			path, version, dataNonce, encryptedData, sv.CreatedTime, sv.DeletedTime,
			wrapResult.KekID, wrapResult.WrappedDEK, wrapResult.Nonce,
			wrapResult.AEADAlg, rewrappedAt)
		if err != nil {
			return fmt.Errorf("failed to store secret version: %w", err)
		}

		// Increment KEK wraps count
		if err := s.kekManager.IncrementWrapsCount(currentKekID); err != nil {
			log.Log().Warn(fName,
				"message", "failed to increment KEK wraps count",
				"kek_id", currentKekID,
				"err", err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true

	log.Log().Info(fName,
		"message", "stored secret with envelope encryption",
		"path", path,
		"kek_id", currentKekID,
		"versions", len(secret.Versions))

	return nil
}

// LoadSecretWithEnvelope loads a secret and automatically handles:
// - Legacy secrets (direct encryption)
// - Envelope-encrypted secrets
// - Lazy rewrapping if KEK is outdated
func (s *DataStore) LoadSecretWithEnvelope(
	ctx context.Context, path string,
) (*kv.Value, bool, error) {
	const fName = "LoadSecretWithEnvelope"
	if ctx == nil {
		log.FatalLn(fName, "message", "nil context")
	}

	s.mu.RLock()
	needsRewrap := false
	currentKekID := ""
	if s.kekManager != nil {
		currentKekID = s.kekManager.GetCurrentKEKID()
	}
	s.mu.RUnlock()

	secret, legacyFormat, err := s.loadSecretEnvelopeInternal(ctx, path)
	if err != nil {
		return nil, false, err
	}

	if secret == nil {
		return nil, false, nil
	}

	// Check if any version needs rewrapping
	if !legacyFormat && s.kekManager != nil {
		for _, sv := range secret.Versions {
			// Check if version has KEK metadata
			if kekID, ok := sv.Data["__kek_metadata__"]; ok && kekID != currentKekID {
				needsRewrap = true
				break
			}
		}
	}

	return secret, needsRewrap, nil
}

// loadSecretEnvelopeInternal loads a secret and determines if it's legacy format
func (s *DataStore) loadSecretEnvelopeInternal(
	ctx context.Context, path string,
) (*kv.Value, bool, error) {
	const fName = "loadSecretEnvelopeInternal"

	s.mu.RLock()
	defer s.mu.RUnlock()

	var secret kv.Value

	// Load metadata
	err := s.db.QueryRowContext(ctx, ddl.QuerySecretMetadata, path).Scan(
		&secret.Metadata.CurrentVersion,
		&secret.Metadata.OldestVersion,
		&secret.Metadata.CreatedTime,
		&secret.Metadata.UpdatedTime,
		&secret.Metadata.MaxVersions)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to load secret metadata: %w", err)
	}

	// Load versions
	rows, err := s.db.QueryContext(ctx, ddl.QuerySecretVersions, path)
	if err != nil {
		return nil, false, fmt.Errorf("failed to query secret versions: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Log().Error(fName, "message", "failed to close rows", "err", closeErr.Error())
		}
	}()

	secret.Versions = make(map[int]kv.Version)
	legacyFormat := false

	for rows.Next() {
		var (
			version     int
			nonce       []byte
			encrypted   []byte
			createdTime time.Time
			deletedTime sql.NullTime
			kekID       sql.NullString
			wrappedDEK  []byte
			dekNonce    []byte
			aeadAlg     sql.NullString
			rewrappedAt sql.NullTime
		)

		if err := rows.Scan(
			&version, &nonce, &encrypted, &createdTime, &deletedTime,
			&kekID, &wrappedDEK, &dekNonce, &aeadAlg, &rewrappedAt,
		); err != nil {
			return nil, false, fmt.Errorf("failed to scan secret version: %w", err)
		}

		var decrypted []byte

		// Check if this is envelope-encrypted or legacy format
		if kekID.Valid && len(wrappedDEK) > 0 && s.kekManager != nil {
			// Envelope encryption format
			kekKey, err := s.kekManager.GetKEK(kekID.String)
			if err != nil {
				return nil, false, fmt.Errorf("failed to get KEK %s: %w", kekID.String, err)
			}

			// Unwrap DEK
			unwrapResult, err := kek.UnwrapDEK(wrappedDEK, kekKey, kekID.String, dekNonce, nil)
			if err != nil {
				return nil, false, fmt.Errorf("failed to unwrap DEK: %w", err)
			}

			// Decrypt data with DEK
			decrypted, err = kek.DecryptWithDEK(encrypted, unwrapResult.DEK, nonce)
			if err != nil {
				return nil, false, fmt.Errorf("failed to decrypt with DEK: %w", err)
			}
		} else {
			// Legacy format - direct encryption with root key cipher
			legacyFormat = true
			var err error
			decrypted, err = s.decrypt(encrypted, nonce)
			if err != nil {
				return nil, false, fmt.Errorf("failed to decrypt secret version: %w", err)
			}
		}

		var values map[string]string
		if err := json.Unmarshal(decrypted, &values); err != nil {
			return nil, false, fmt.Errorf("failed to unmarshal secret values: %w", err)
		}

		// Store KEK metadata in a special field for rewrap detection
		if kekID.Valid {
			if values == nil {
				values = make(map[string]string)
			}
			values["__kek_metadata__"] = kekID.String
		}

		sv := kv.Version{
			Data:        values,
			CreatedTime: createdTime,
		}
		if deletedTime.Valid {
			sv.DeletedTime = &deletedTime.Time
		}

		secret.Versions[version] = sv
	}

	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("failed to iterate secret versions: %w", err)
	}

	return &secret, legacyFormat, nil
}

// RewrapSecret rewraps a secret's DEK with the current KEK
// This is called by the lazy rewrap mechanism
func (s *DataStore) RewrapSecret(ctx context.Context, path string, version int) error {
	const fName = "RewrapSecret"
	if ctx == nil {
		return fmt.Errorf("%s: nil context", fName)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.kekManager == nil {
		return fmt.Errorf("%s: KEK manager not initialized", fName)
	}

	// Get current KEK
	currentKekID := s.kekManager.GetCurrentKEKID()
	currentKek, err := s.kekManager.GetKEK(currentKekID)
	if err != nil {
		return fmt.Errorf("%s: failed to get current KEK: %w", fName, err)
	}

	// Load the secret version to rewrap
	var (
		oldKekID   sql.NullString
		wrappedDEK []byte
		dekNonce   []byte
	)

	err = s.db.QueryRowContext(ctx,
		`SELECT kek_id, wrapped_dek, dek_nonce FROM secrets WHERE path = ? AND version = ?`,
		path, version).Scan(&oldKekID, &wrappedDEK, &dekNonce)
	if err != nil {
		return fmt.Errorf("%s: failed to load secret version: %w", fName, err)
	}

	if !oldKekID.Valid || oldKekID.String == currentKekID {
		// Already using current KEK
		return nil
	}

	// Get old KEK
	oldKek, err := s.kekManager.GetKEK(oldKekID.String)
	if err != nil {
		return fmt.Errorf("%s: failed to get old KEK: %w", fName, err)
	}

	// Rewrap DEK
	wrapResult, err := kek.RewrapDEK(
		wrappedDEK, oldKek, oldKekID.String, dekNonce, nil,
		currentKek, currentKekID, nil,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to rewrap DEK: %w", fName, err)
	}

	// Update the secret version with new wrapped DEK
	rewrappedAt := time.Now()
	_, err = s.db.ExecContext(ctx,
		`UPDATE secrets SET kek_id = ?, wrapped_dek = ?, dek_nonce = ?, 
		 aead_alg = ?, rewrapped_at = ? WHERE path = ? AND version = ?`,
		wrapResult.KekID, wrapResult.WrappedDEK, wrapResult.Nonce,
		wrapResult.AEADAlg, rewrappedAt, path, version)
	if err != nil {
		return fmt.Errorf("%s: failed to update secret version: %w", fName, err)
	}

	log.Log().Info(fName,
		"message", "rewrapped secret",
		"path", path,
		"version", version,
		"old_kek_id", oldKekID.String,
		"new_kek_id", currentKekID)

	return nil
}
