//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
	"github.com/spiffe/spike/app/nexus/internal/state/kek"
)

// StoreKEKMetadata persists KEK metadata to the database
func (s *DataStore) StoreKEKMetadata(metadata *kek.Metadata) error {
	const fName = "StoreKEKMetadata"

	s.mu.Lock()
	defer s.mu.Unlock()

	var retiredAt interface{}
	if metadata.RetiredAt != nil {
		retiredAt = metadata.RetiredAt
	}

	_, err := s.db.Exec(
		ddl.QueryInsertKEKMetadata,
		metadata.ID,
		metadata.Version,
		metadata.Salt[:],
		metadata.RMKVersion,
		metadata.CreatedAt,
		metadata.WrapsCount,
		string(metadata.Status),
		retiredAt,
	)

	if err != nil {
		log.Log().Error(fName,
			"message", "failed to store KEK metadata",
			"kek_id", metadata.ID,
			"err", err.Error())
		return fmt.Errorf("%s: %w", fName, err)
	}

	log.Log().Info(fName,
		"message", "stored KEK metadata",
		"kek_id", metadata.ID,
		"version", metadata.Version)

	return nil
}

// LoadKEKMetadata retrieves KEK metadata by ID
func (s *DataStore) LoadKEKMetadata(kekID string) (*kek.Metadata, error) {
	const fName = "LoadKEKMetadata"

	s.mu.RLock()
	defer s.mu.RUnlock()

	var meta kek.Metadata
	var saltBytes []byte
	var statusStr string
	var retiredAt sql.NullTime

	err := s.db.QueryRow(ddl.QueryLoadKEKMetadata, kekID).Scan(
		&meta.ID,
		&meta.Version,
		&saltBytes,
		&meta.RMKVersion,
		&meta.CreatedAt,
		&meta.WrapsCount,
		&statusStr,
		&retiredAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: KEK metadata not found: %s", fName, kekID)
		}
		log.Log().Error(fName,
			"message", "failed to load KEK metadata",
			"kek_id", kekID,
			"err", err.Error())
		return nil, fmt.Errorf("%s: %w", fName, err)
	}

	// Copy salt bytes
	if len(saltBytes) != kek.KekSaltSize {
		return nil, fmt.Errorf("%s: invalid salt size: %d", fName, len(saltBytes))
	}
	copy(meta.Salt[:], saltBytes)

	meta.Status = kek.KekStatus(statusStr)

	if retiredAt.Valid {
		meta.RetiredAt = &retiredAt.Time
	}

	return &meta, nil
}

// ListKEKMetadata returns all KEK metadata records
func (s *DataStore) ListKEKMetadata() ([]*kek.Metadata, error) {
	const fName = "ListKEKMetadata"

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(ddl.QueryListKEKMetadata)
	if err != nil {
		log.Log().Error(fName, "message", "failed to query KEK metadata", "err", err.Error())
		return nil, fmt.Errorf("%s: %w", fName, err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Log().Error(fName, "message", "failed to close rows", "err", closeErr.Error())
		}
	}()

	var metadataList []*kek.Metadata

	for rows.Next() {
		var meta kek.Metadata
		var saltBytes []byte
		var statusStr string
		var retiredAt sql.NullTime

		if err := rows.Scan(
			&meta.ID,
			&meta.Version,
			&saltBytes,
			&meta.RMKVersion,
			&meta.CreatedAt,
			&meta.WrapsCount,
			&statusStr,
			&retiredAt,
		); err != nil {
			log.Log().Error(fName, "message", "failed to scan KEK metadata row", "err", err.Error())
			return nil, fmt.Errorf("%s: %w", fName, err)
		}

		// Copy salt bytes
		if len(saltBytes) != kek.KekSaltSize {
			log.Log().Error(fName, "message", "invalid salt size", "size", len(saltBytes))
			continue
		}
		copy(meta.Salt[:], saltBytes)

		meta.Status = kek.KekStatus(statusStr)

		if retiredAt.Valid {
			meta.RetiredAt = &retiredAt.Time
		}

		metadataList = append(metadataList, &meta)
	}

	if err := rows.Err(); err != nil {
		log.Log().Error(fName, "message", "error iterating KEK metadata rows", "err", err.Error())
		return nil, fmt.Errorf("%s: %w", fName, err)
	}

	log.Log().Info(fName, "message", "listed KEK metadata", "count", len(metadataList))

	return metadataList, nil
}

// UpdateKEKWrapsCount atomically increments the wraps count for a KEK
func (s *DataStore) UpdateKEKWrapsCount(kekID string, delta int64) error {
	const fName = "UpdateKEKWrapsCount"

	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.db.Exec(ddl.QueryUpdateKEKWrapsCount, delta, kekID)
	if err != nil {
		log.Log().Error(fName,
			"message", "failed to update KEK wraps count",
			"kek_id", kekID,
			"err", err.Error())
		return fmt.Errorf("%s: %w", fName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", fName, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: KEK not found: %s", fName, kekID)
	}

	return nil
}

// UpdateKEKStatus updates the status of a KEK
func (s *DataStore) UpdateKEKStatus(kekID string, status kek.KekStatus, retiredAt *time.Time) error {
	const fName = "UpdateKEKStatus"

	s.mu.Lock()
	defer s.mu.Unlock()

	var retiredAtVal interface{}
	if retiredAt != nil {
		retiredAtVal = retiredAt
	}

	result, err := s.db.Exec(ddl.QueryUpdateKEKStatus, string(status), retiredAtVal, kekID)
	if err != nil {
		log.Log().Error(fName,
			"message", "failed to update KEK status",
			"kek_id", kekID,
			"err", err.Error())
		return fmt.Errorf("%s: %w", fName, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected: %w", fName, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: KEK not found: %s", fName, kekID)
	}

	log.Log().Info(fName,
		"message", "updated KEK status",
		"kek_id", kekID,
		"status", status)

	return nil
}
