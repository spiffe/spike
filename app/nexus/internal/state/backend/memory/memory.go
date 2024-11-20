//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	"github.com/spiffe/spike/app/nexus/internal/state/entity/data"
	"github.com/spiffe/spike/pkg/store"
)

// NoopStore provides a no-op implementation of a storage backend.
// This implementation can be used for testing or as a placeholder
// where no actual storage is needed. NoopStore is also use when the
// backing store is configured to be in-memory.
type NoopStore struct {
}

func (s *NoopStore) StoreAdminRecoveryMetadata(
	ctx context.Context, credentials data.RecoveryMetadata,
) error {
	return nil
}

func (s *NoopStore) LoadAdminRecoveryMetadata(
	ctx context.Context,
) (data.RecoveryMetadata, error) {
	return data.RecoveryMetadata{}, nil
}

// Close implements the closing operation for the store.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) Close(_ context.Context) error {
	return nil
}

// Initialize prepares the store for use.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) Initialize(_ context.Context) error {
	return nil
}

// LoadAdminToken retrieves the admin token from the store.
// This implementation always returns an empty string and nil error.
func (s *NoopStore) LoadAdminToken(_ context.Context) (string, error) {
	return "", nil
}

// LoadSecret retrieves a secret from the store by its path.
// This implementation always returns nil secret and nil error.
func (s *NoopStore) LoadSecret(
	_ context.Context, _ string,
) (*store.Secret, error) {
	return nil, nil
}

// StoreAdminToken saves an admin token to the store.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) StoreAdminToken(_ context.Context, _ string) error {
	return nil
}

// StoreSecret saves a secret to the store at the specified path.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) StoreSecret(
	_ context.Context, _ string, _ store.Secret,
) error {
	return nil
}
