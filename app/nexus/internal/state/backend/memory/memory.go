//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

// Package memory implements an in-memory storage backend for managing
// secrets and policies in the SPIKE system. This package provides a
// fully functional in-memory implementation, `Store`, which is suitable
// for development, testing, or scenarios where persistent storage is
// not required.
//
// The `Store` provides thread-safe implementations for interfaces related
// to storing and retrieving secrets and policies. Unlike the noop backend,
// this implementation actually stores data in memory using the kv package
// and maintains the proper state throughout the application lifecycle.
package memory

import (
	"context"
	"crypto/cipher"
	"errors"
	"fmt"
	"sync"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
)

// Store provides an in-memory implementation of a storage backend.
// This implementation actually stores data in memory using the kv package,
// suitable for development, testing, or scenarios where persistence isn't needed.
type Store struct {
	secretStore *kv.KV
	secretMu    sync.RWMutex

	policies map[string]*data.Policy
	policyMu sync.RWMutex

	cipher cipher.AEAD
}

// NewInMemoryStore creates a new in-memory store instance
func NewInMemoryStore(cipher cipher.AEAD, maxVersions int) *Store {
	return &Store{
		secretStore: kv.New(kv.Config{
			MaxSecretVersions: maxVersions,
		}),
		policies: make(map[string]*data.Policy),
		cipher:   cipher,
	}
}

// Initialize prepares the store for use.
func (s *Store) Initialize(_ context.Context) *sdkErrors.SDKError {
	// Already initialized in constructor
	return nil
}

// Close implements the closing operation for the store.
func (s *Store) Close(_ context.Context) *sdkErrors.SDKError {
	// Nothing to close for in-memory store
	return nil
}

// StoreSecret saves a secret to the store at the specified path.
func (s *Store) StoreSecret(
	_ context.Context, path string, secret kv.Value,
) error {
	s.secretMu.Lock()
	defer s.secretMu.Unlock()

	// Store the entire secret structure
	s.secretStore.ImportSecrets(map[string]*kv.Value{
		path: &secret,
	})

	return nil
}

// LoadSecret retrieves a secret from the store by its path.
func (s *Store) LoadSecret(
	_ context.Context, path string,
) (*kv.Value, error) {
	s.secretMu.RLock()
	defer s.secretMu.RUnlock()

	rawSecret, err := s.secretStore.GetRawSecret(path)
	if err != nil && errors.Is(err, sdkErrors.ErrEntityNotFound) {
		// To align with the SQLite implementation, don't return an error for
		// "not found" items and just return a `nil` secret.
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return rawSecret, nil
}

// LoadAllSecrets retrieves all secrets stored in the store.
func (s *Store) LoadAllSecrets(_ context.Context) (
	map[string]*kv.Value, error,
) {
	s.secretMu.RLock()
	defer s.secretMu.RUnlock()

	result := make(map[string]*kv.Value)

	// Get all paths
	paths := s.secretStore.List()

	// Load each secret
	for _, path := range paths {
		secret, err := s.secretStore.GetRawSecret(path)
		if err != nil {
			continue // Skip secrets that can't be loaded
		}
		result[path] = secret
	}

	return result, nil
}

// StorePolicy stores a policy in the store.
func (s *Store) StorePolicy(_ context.Context, policy data.Policy) error {
	s.policyMu.Lock()
	defer s.policyMu.Unlock()

	if policy.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}

	s.policies[policy.ID] = &policy
	return nil
}

// LoadPolicy retrieves a policy from the store by its ID.
func (s *Store) LoadPolicy(
	_ context.Context, id string,
) (*data.Policy, error) {
	s.policyMu.RLock()
	defer s.policyMu.RUnlock()

	policy, exists := s.policies[id]
	if !exists {
		return nil, nil // Return nil, nil for not found (matching Store behavior)
	}

	return policy, nil
}

// LoadAllPolicies retrieves all policies from the store.
func (s *Store) LoadAllPolicies(
	_ context.Context,
) (map[string]*data.Policy, error) {
	s.policyMu.RLock()
	defer s.policyMu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*data.Policy, len(s.policies))
	for id, policy := range s.policies {
		result[id] = policy
	}

	return result, nil
}

// DeletePolicy removes a policy from the store by its ID.
func (s *Store) DeletePolicy(_ context.Context, id string) error {
	s.policyMu.Lock()
	defer s.policyMu.Unlock()

	delete(s.policies, id)
	return nil
}

// GetCipher returns the cipher used for encryption/decryption
func (s *Store) GetCipher() cipher.AEAD {
	return s.cipher
}
