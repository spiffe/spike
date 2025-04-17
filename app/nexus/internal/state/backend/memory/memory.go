//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package memory implements an in-memory storage backend for managing
// secrets and policies in the SPIKE system. This package includes a
// no-op implementation, `NoopStore`, which acts as a placeholder or
// testing tool for scenarios where persistent storage isn't required.
//
// The `NoopStore` provides implementations for interfaces related to
// storing and retrieving secrets and policies but does not perform
// any actual storage operations. All methods in `NoopStore` are no-ops and
// always return `nil` or equivalent default values.
package memory

import (
	"context"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/kv"
)

// NoopStore provides a no-op implementation of a storage backend.
// This implementation can be used for testing or as a placeholder
// where no actual storage is needed. NoopStore is also use when the
// backing kv is configured to be in-memory.
type NoopStore struct {
}

// Close implements the closing operation for the kv.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) Close(_ context.Context) error {
	return nil
}

// Initialize prepares the kv for use.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) Initialize(_ context.Context) error {
	return nil
}

// LoadSecret retrieves a secret from the kv by its path.
// This implementation always returns nil secret and nil error.
func (s *NoopStore) LoadSecret(
	_ context.Context, _ string,
) (*kv.Value, error) {
	return nil, nil
}

// LoadAllSecrets retrieves all secrets stored in the kv.
// This implementation always returns nil and nil error.
func (s *NoopStore) LoadAllSecrets(_ context.Context) (
	map[string]*kv.Value, error,
) {
	return nil, nil
}

// StoreSecret saves a secret to the kv at the specified path.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) StoreSecret(
	_ context.Context, _ string, _ kv.Value,
) error {
	return nil
}

// StorePolicy stores a policy in the no-op kv.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) StorePolicy(_ context.Context, _ data.Policy) error {
	return nil
}

// LoadPolicy retrieves a policy from the kv by its ID.
// This implementation always returns nil and nil error.
func (s *NoopStore) LoadPolicy(
	_ context.Context, _ string,
) (*data.Policy, error) {
	return nil, nil
}

// LoadAllPolicies retrieves all policies from the no-op store.
// This implementation always returns nil and nil error.
func (s *NoopStore) LoadAllPolicies(
	_ context.Context,
) (map[string]*data.Policy, error) {
	return nil, nil
}

// DeletePolicy removes a policy from the no-op kv by its ID.
// This implementation is a no-op and always returns nil.
func (s *NoopStore) DeletePolicy(_ context.Context, _ string) error {
	return nil
}
