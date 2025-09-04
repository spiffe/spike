//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"

	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/ddl"
)

// DeletePolicy removes a policy from the database by its ID.
//
// Uses serializable transaction isolation to ensure consistency.
// Automatically rolls back on error.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: Unique identifier of the policy to delete
//
// Returns error if:
//   - Transaction operations fail
//   - Policy deletion fails
func (s *DataStore) DeletePolicy(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
		}
	}(tx)

	_, err = tx.ExecContext(ctx, ddl.QueryDeletePolicy, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}

// StorePolicy saves or updates a policy in the database.
//
// Uses serializable transaction isolation to ensure consistency.
// Automatically rolls back on error.
//
// Parameters:
//   - ctx: Context for the database operation
//   - policy: Policy data to store, containing ID, name, patterns, and creation
//     time
//
// Returns error if:
//   - Transaction operations fail
//   - Policy storage fails
func (s *DataStore) StorePolicy(ctx context.Context, policy data.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func(tx *sql.Tx) {
		if !committed {
			err := tx.Rollback()
			if err != nil {
				fmt.Printf("failed to rollback transaction: %v\n", err)
			}
		}
	}(tx)

	// Serialize permissions to comma-separated string
	permissionsStr := ""
	if len(policy.Permissions) > 0 {
		permissions := make([]string, len(policy.Permissions))
		for i, perm := range policy.Permissions {
			permissions[i] = string(perm)
		}
		permissionsStr = strings.Join(permissions, ",")
	}

	// Encryption
	encryptedSPIFFE, nonceSPIFFE, err := s.encrypt([]byte(policy.SPIFFEIDPattern))
	if err != nil {
		return fmt.Errorf("failed to encrypt spiffeid pattern: %w", err)
	}

	encryptedPath, noncePath, err := s.encryptField(policy.PathPattern)
	if err != nil {
		return fmt.Errorf("failed to encrypt path pattern: %w", err)
	}

	_, err = tx.ExecContext(ctx, ddl.QueryUpsertPolicy,
		policy.ID,
		policy.Name,
		nonceSPIFFE,     // nonce_spiffe
		encryptedSPIFFE, // encrypted_spiffe_id_pattern
		noncePath,       // nonce_path
		encryptedPath,   // encrypted_path_pattern
		permissionsStr,
		policy.CreatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert policy: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	committed = true
	return nil
}

func (s *DataStore) encryptField(value string) ([]byte, []byte, error) {
	return s.encrypt([]byte(value))
}

// LoadPolicy retrieves a policy from the database and compiles its patterns.
//
// Parameters:
//   - ctx: Context for the database operation
//   - id: Unique identifier of the policy to load
//
// Returns:
//   - *data.Policy: Loaded policy with compiled patterns, nil if not found
//   - error: Database errors or pattern compilation errors
func (s *DataStore) LoadPolicy(
	ctx context.Context, id string,
) (*data.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var policy data.Policy
	var encryptedSPIFFE, encryptedPath []byte
	var nonceSPIFFE, noncePath []byte
	var permissionsStr string
	var createdTime int64

	err := s.db.QueryRowContext(ctx, ddl.QueryLoadPolicy, id).Scan(
		&policy.Name,
		&encryptedSPIFFE,
		&encryptedPath,
		&permissionsStr,
		&nonceSPIFFE,
		&noncePath,
		&createdTime,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	// Decrypt
	decryptedSPIFFE, err := s.decrypt(encryptedSPIFFE, nonceSPIFFE)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SPIFFE pattern: %w", err)
	}
	decryptedPath, err := s.decrypt(encryptedPath, noncePath)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt Path pattern: %w", err)
	}

	// Set decrypted values
	policy.SPIFFEIDPattern = string(decryptedSPIFFE)
	policy.PathPattern = string(decryptedPath)
	policy.CreatedAt = time.Unix(createdTime, 0)

	// Deserialize permissions
	if permissionsStr != "" {
		perms := strings.Split(permissionsStr, ",")
		policy.Permissions = make([]data.PolicyPermission, len(perms))
		for i, p := range perms {
			policy.Permissions[i] = data.PolicyPermission(strings.TrimSpace(p))
		}
	}

	// Compile regex
	policy.IDRegex, err = regexp.Compile(policy.SPIFFEIDPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid spiffeid pattern: %w", err)
	}
	policy.PathRegex, err = regexp.Compile(policy.PathPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid path pattern: %w", err)
	}

	return &policy, nil
}

// LoadAllPolicies retrieves all policies from the backend storage.
//
// The function loads all policy data and compiles regex patterns for SPIFFE ID
// and path matching if they aren't wildcards (*).
//
// Parameters:
//   - ctx: Context for the database operation
//
// Returns:
//   - map[string]*data.Policy: Map of policy IDs to loaded policies with
//     compiled patterns
//   - error: Database errors or pattern compilation errors
func (s *DataStore) LoadAllPolicies(ctx context.Context) (map[string]*data.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, ddl.QueryAllPolicies)
	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	policies := make(map[string]*data.Policy)

	for rows.Next() {
		var policy data.Policy
		var encryptedSPIFFE, encryptedPath []byte
		var nonceSPIFFE, noncePath []byte
		var permissionsStr string
		var createdTime int64

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&encryptedSPIFFE,
			&encryptedPath,
			&permissionsStr,
			&nonceSPIFFE,
			&noncePath,
			&createdTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		// Decrypt
		decryptedSPIFFE, err := s.decrypt(encryptedSPIFFE, nonceSPIFFE)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt SPIFFE pattern for policy %s: %w", policy.ID, err)
		}
		decryptedPath, err := s.decrypt(encryptedPath, noncePath)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt path pattern for policy %s: %w", policy.ID, err)
		}

		policy.SPIFFEIDPattern = string(decryptedSPIFFE)
		policy.PathPattern = string(decryptedPath)
		policy.CreatedAt = time.Unix(createdTime, 0)

		// Deserialize permissions
		if permissionsStr != "" {
			perms := strings.Split(permissionsStr, ",")
			policy.Permissions = make([]data.PolicyPermission, len(perms))
			for i, p := range perms {
				policy.Permissions[i] = data.PolicyPermission(strings.TrimSpace(p))
			}
		}

		// Compile regex
		policy.IDRegex, err = regexp.Compile(policy.SPIFFEIDPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid spiffeid pattern for policy %s: %w", policy.ID, err)
		}
		policy.PathRegex, err = regexp.Compile(policy.PathPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid path pattern for policy %s: %w", policy.ID, err)
		}

		policies[policy.ID] = &policy
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate policy rows: %w", err)
	}

	return policies, nil
}
