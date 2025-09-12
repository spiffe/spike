//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/log"

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
	const fName = "DeletePolicy"
	if ctx == nil {
		log.FatalLn(fName, "message", "nil context")
	}

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

func generateNonce(s *DataStore) ([]byte, error) {
	nonce := make([]byte, s.Cipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

func encryptWithNonce(s *DataStore, nonce []byte, data []byte) ([]byte, error) {
	if len(nonce) != s.Cipher.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size: got %d, want %d", len(nonce), s.Cipher.NonceSize())
	}
	ciphertext := s.Cipher.Seal(nil, nonce, data, nil)
	return ciphertext, nil
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
	const fName = "StorePolicy"
	if ctx == nil {
		log.FatalLn(fName, "message", "nil context")
	}

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
	nonce, err := generateNonce(s)
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	encryptedSpiffeID, err := encryptWithNonce(s, nonce, []byte(policy.SPIFFEIDPattern))
	if err != nil {
		return fmt.Errorf("failed to encrypt SPIFFE ID: %w", err)
	}

	encryptedPathPattern, err := encryptWithNonce(s, nonce, []byte(policy.PathPattern))
	if err != nil {
		return fmt.Errorf("failed to encrypt path pattern: %w", err)
	}
	encryptedPermissions, err := encryptWithNonce(s, nonce, []byte(permissionsStr))
	if err != nil {
		return fmt.Errorf("failed to encrypt permissions: %w", err)
	}

	_, err = tx.ExecContext(ctx, ddl.QueryUpsertPolicy,
		policy.ID,
		policy.Name,
		nonce,
		encryptedSpiffeID,
		encryptedPathPattern,
		encryptedPermissions,
		time.Now().Unix(),
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
	const fName = "LoadPolicy"
	if ctx == nil {
		log.FatalLn(fName, "message", "nil context")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var policy data.Policy
	var encryptedSPIFFEIDPattern, encryptedPathPattern, encryptedPermissions, nonce []byte
	var createdTime int64

	err := s.db.QueryRowContext(ctx, ddl.QueryLoadPolicy, id).Scan(
		&policy.ID,
		&policy.Name,
		&encryptedSPIFFEIDPattern,
		&encryptedPathPattern,
		&encryptedPermissions,
		&nonce,
		&createdTime,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	// Decrypt
	decryptedSPIFFEIDPattern, err := s.decrypt(encryptedSPIFFEIDPattern, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SPIFFE ID pattern: %w", err)
	}
	decryptedPathPattern, err := s.decrypt(encryptedPathPattern, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt path pattern: %w", err)
	}

	decryptedPermissions, err := s.decrypt(encryptedPermissions, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt permissions: %w", err)
	}

	// Set decrypted values
	policy.SPIFFEIDPattern = string(decryptedSPIFFEIDPattern)
	policy.PathPattern = string(decryptedPathPattern)
	policy.CreatedAt = time.Unix(createdTime, 0)

	permissionsStr := string(decryptedPermissions)
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
func (s *DataStore) LoadAllPolicies(
	ctx context.Context,
) (map[string]*data.Policy, error) {
	const fName = "LoadAllPolicies"
	if ctx == nil {
		log.FatalLn(fName, "message", "nil context")
	}

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
		var encryptedSPIFFEIDPattern, encryptedPathPattern, encryptedPermissions, nonce []byte
		var createdTime int64

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&encryptedSPIFFEIDPattern,
			&encryptedPathPattern,
			&encryptedPermissions,
			&nonce,
			&createdTime,
		); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		// Decrypt
		decryptedSPIFFEIDPattern, err := s.decrypt(encryptedSPIFFEIDPattern, nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt SPIFFE ID pattern for policy %s: %w", policy.ID, err)
		}
		decryptedPathPattern, err := s.decrypt(encryptedPathPattern, nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt path pattern for policy %s: %w", policy.ID, err)
		}
		decryptedPermissions, err := s.decrypt(encryptedPermissions, nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt permissions for policy %s: %w", policy.ID, err)
		}

		policy.SPIFFEIDPattern = string(decryptedSPIFFEIDPattern)
		policy.PathPattern = string(decryptedPathPattern)
		policy.CreatedAt = time.Unix(createdTime, 0)

		// Deserialize permissions
		permissionsStr := string(decryptedPermissions)
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
