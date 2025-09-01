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

	_, err = tx.ExecContext(ctx, ddl.QueryUpsertPolicy,
		policy.ID,
		policy.Name,
		policy.SPIFFEIDPattern,
		policy.PathPattern,
		permissionsStr,
		policy.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store policy: %w", err)
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	var policy data.Policy
	policy.ID = id // Set the ID since we queried with it

	var permissionsStr string
	err := s.db.QueryRowContext(ctx, ddl.QueryLoadPolicy, id).Scan(
		&policy.Name,
		&policy.SPIFFEIDPattern,
		&policy.PathPattern,
		&permissionsStr,
		&policy.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	// Deserialize permissions from comma-separated string
	if permissionsStr != "" {
		permissionStrs := strings.Split(permissionsStr, ",")
		policy.Permissions = make([]data.PolicyPermission, len(permissionStrs))
		for i, permStr := range permissionStrs {
			policy.Permissions[i] = data.PolicyPermission(strings.TrimSpace(permStr))
		}
	}

	idRegex, err := regexp.Compile(policy.SPIFFEIDPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid spiffeid pattern: %w", err)
	}
	policy.IDRegex = idRegex

	pathRegex, err := regexp.Compile(policy.PathPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid path pattern: %w", err)
	}
	policy.PathRegex = pathRegex

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
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Query to get all policies
	rows, err := s.db.QueryContext(ctx, ddl.QueryAllPolicies)

	if err != nil {
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("failed to close rows: %v\n", err)
		}
	}(rows)

	// Map to store all policies
	policies := make(map[string]*data.Policy)

	// Iterate over policy rows
	for rows.Next() {
		var policy data.Policy
		var permissionsStr string

		if err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.SPIFFEIDPattern,
			&policy.PathPattern,
			&permissionsStr,
			&policy.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}

		// Deserialize permissions from comma-separated string
		if permissionsStr != "" {
			permissionStrs := strings.Split(permissionsStr, ",")
			policy.Permissions = make([]data.PolicyPermission, len(permissionStrs))
			for i, permStr := range permissionStrs {
				policy.Permissions[i] = data.PolicyPermission(strings.TrimSpace(permStr))
			}
		}

		// Compile the patterns just like in LoadPolicy
		idRegex, err := regexp.Compile(policy.SPIFFEIDPattern)
		if err != nil {
			return nil,
				fmt.Errorf("invalid spiffeid pattern for policy %s: %w", policy.ID, err)
		}
		policy.IDRegex = idRegex

		pathRegex, err := regexp.Compile(policy.PathPattern)
		if err != nil {
			return nil,
				fmt.Errorf("invalid path pattern for policy %s: %w", policy.ID, err)
		}
		policy.PathRegex = pathRegex

		policies[policy.ID] = &policy
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate policy rows: %w", err)
	}

	return policies, nil
}
