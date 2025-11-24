//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package noop implements an in-memory storage backend for managing
// secrets and policies in the SPIKE system. This package includes a
// no-op implementation, `Store`, which acts as a placeholder or
// testing tool for scenarios where persistent storage isn't required.
//
// The `Store` provides implementations for interfaces related to
// storing and retrieving secrets and policies but does not perform
// any actual storage operations. All methods in `Store` are no-ops and
// always return `nil` or equivalent default values.
package noop

import (
	"context"
	"crypto/cipher"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
)

// Store provides a no-op implementation of a storage backend.
// This implementation can be used for testing or as a placeholder
// where no actual storage is needed. Store is also used when the
// backing kv is configured to be in-memory.
type Store struct {
}

// Close implements the closing operation for the store.
//
// This is a no-op implementation that always succeeds. It exists to satisfy
// the backend interface but performs no actual cleanup operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) Close(_ context.Context) *sdkErrors.SDKError {
	return nil
}

// Initialize prepares the store for use.
//
// This is a no-op implementation that always succeeds. It exists to satisfy
// the backend interface but performs no actual initialization operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) Initialize(_ context.Context) *sdkErrors.SDKError {
	return nil
}

// LoadSecret retrieves a secret from the store by its path.
//
// This is a no-op implementation that always returns nil values. It exists to
// satisfy the backend interface but performs no actual retrieval operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//   - string: The secret path (ignored in this implementation)
//
// Returns:
//   - *kv.Value: Always returns nil
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) LoadSecret(
	_ context.Context, _ string,
) (*kv.Value, *sdkErrors.SDKError) {
	return nil, nil
}

// LoadAllSecrets retrieves all secrets stored in the store.
//
// This is a no-op implementation that always returns nil. It exists to
// satisfy the backend interface but performs no actual retrieval operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//
// Returns:
//   - map[string]*kv.Value: Always returns nil
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) LoadAllSecrets(_ context.Context) (
	map[string]*kv.Value, *sdkErrors.SDKError,
) {
	return nil, nil
}

// StoreSecret saves a secret to the store at the specified path.
//
// This is a no-op implementation that always succeeds. It exists to satisfy
// the backend interface but performs no actual storage operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//   - string: The secret path (ignored in this implementation)
//   - kv.Value: The secret value (ignored in this implementation)
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) StoreSecret(
	_ context.Context, _ string, _ kv.Value,
) *sdkErrors.SDKError {
	return nil
}

// StorePolicy stores a policy in the store.
//
// This is a no-op implementation that always succeeds. It exists to satisfy
// the backend interface but performs no actual storage operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//   - data.Policy: The policy to store (ignored in this implementation)
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) StorePolicy(
	_ context.Context, _ data.Policy,
) *sdkErrors.SDKError {
	return nil
}

// LoadPolicy retrieves a policy from the store by its ID.
//
// This is a no-op implementation that always returns nil values. It exists to
// satisfy the backend interface but performs no actual retrieval operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//   - string: The policy ID (ignored in this implementation)
//
// Returns:
//   - *data.Policy: Always returns nil
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) LoadPolicy(
	_ context.Context, _ string,
) (*data.Policy, *sdkErrors.SDKError) {
	return nil, nil
}

// LoadAllPolicies retrieves all policies from the store.
//
// This is a no-op implementation that always returns nil. It exists to
// satisfy the backend interface but performs no actual retrieval operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//
// Returns:
//   - map[string]*data.Policy: Always returns nil
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) LoadAllPolicies(
	_ context.Context,
) (map[string]*data.Policy, *sdkErrors.SDKError) {
	return nil, nil
}

// DeletePolicy removes a policy from the store by its ID.
//
// This is a no-op implementation that always succeeds. It exists to satisfy
// the backend interface but performs no actual deletion operations.
//
// Parameters:
//   - context.Context: Ignored in this implementation
//   - string: The policy ID (ignored in this implementation)
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) DeletePolicy(_ context.Context, _ string) *sdkErrors.SDKError {
	return nil
}

// GetCipher returns the cipher used for encryption/decryption.
//
// This is a no-op implementation that always returns nil. It exists to
// satisfy the backend interface but provides no cipher since no actual
// encryption operations are performed.
//
// Returns:
//   - cipher.AEAD: Always returns nil
func (s *Store) GetCipher() cipher.AEAD {
	return nil
}
