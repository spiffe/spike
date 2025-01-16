//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"

	"github.com/spiffe/spike-sdk-go/api/entity/data"

	"github.com/spiffe/spike/pkg/store"
)

// NoopStore provides a no-op implementation of a storage backend.
// This implementation can be used for testing or as a placeholder
// where no actual storage is needed. NoopStore is also use when the
// backing store is configured to be in-memory.
type NoopStore struct {
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

// LoadSecret retrieves a secret from the store by its path.
// This implementation always returns nil secret and nil error.
func (s *NoopStore) LoadSecret(
	_ context.Context, _ string,
) (*store.Secret, error) {
	return nil, nil
}

// StoreSecret saves a secret to the store at the specified path.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) StoreSecret(
	_ context.Context, _ string, _ store.Secret,
) error {
	return nil
}

// StorePolicy stores a policy in the no-op store.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) StorePolicy(_ context.Context, _ data.Policy) error {
	return nil
}

// LoadPolicy retrieves a policy from the store by its ID.
// This implementation always returns nil and nil error.
func (s *NoopStore) LoadPolicy(
	_ context.Context, _ string,
) (*data.Policy, error) {
	return nil, nil
}

// DeletePolicy removes a policy from the no-op store by its ID.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) DeletePolicy(_ context.Context, _ string) error {
	return nil
}

// TODO: store likely does not need key recovery info as we keep the root
// key in memory; and without the root key, the recovery info is irrecoverable
// anyway. So it's more of a chicken-egg situation.

// StoreKeyRecoveryInfo saves key recovery information to the no-op store.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) StoreKeyRecoveryInfo(
	_ context.Context, _ store.KeyRecoveryData,
) error {
	return nil
}

// LoadKeyRecoveryInfo retrieves key recovery data from the no-op store.
// This implementation always returns nil for both data and error.
func (s *NoopStore) LoadKeyRecoveryInfo(
	_ context.Context,
) (meta *store.KeyRecoveryData, err error) {
	return nil, nil
}
