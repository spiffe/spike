//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"crypto/cipher"
	"errors"
	"fmt"
	"sync"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/kv"
)

// InMemoryStore provides an in-memory implementation of a storage backend.
// This implementation actually stores data in memory using the kv package,
// suitable for development, testing, or scenarios where persistence isn't needed.
type InMemoryStore struct {
	secretStore *kv.KV
	secretMu    sync.RWMutex

	policies map[string]*data.Policy
	policyMu sync.RWMutex

	cipher cipher.AEAD
}

// NewInMemoryStore creates a new in-memory store instance
func NewInMemoryStore(cipher cipher.AEAD, maxVersions int) *InMemoryStore {
	return &InMemoryStore{
		secretStore: kv.New(kv.Config{
			MaxSecretVersions: maxVersions,
		}),
		policies: make(map[string]*data.Policy),
		cipher:   cipher,
	}
}

// Initialize prepares the store for use.
func (s *InMemoryStore) Initialize(_ context.Context) error {
	// Already initialized in constructor
	return nil
}

// Close implements the closing operation for the store.
func (s *InMemoryStore) Close(_ context.Context) error {
	// Nothing to close for in-memory store
	return nil
}

// StoreSecret saves a secret to the store at the specified path.
func (s *InMemoryStore) StoreSecret(
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
func (s *InMemoryStore) LoadSecret(
	_ context.Context, path string,
) (*kv.Value, error) {
	s.secretMu.RLock()
	defer s.secretMu.RUnlock()

	rawSecret, err := s.secretStore.GetRawSecret(path)
	if err != nil && errors.Is(err, kv.ErrItemNotFound) {
		// To align with the SQLite implementation, don't return an error for
		// "not found" items and just return a `nil` secret.
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return rawSecret, nil
}

// LoadAllSecrets retrieves all secrets stored in the store.
func (s *InMemoryStore) LoadAllSecrets(_ context.Context) (
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
func (s *InMemoryStore) StorePolicy(_ context.Context, policy data.Policy) error {
	s.policyMu.Lock()
	defer s.policyMu.Unlock()

	if policy.ID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}

	s.policies[policy.ID] = &policy
	return nil
}

// LoadPolicy retrieves a policy from the store by its ID.
func (s *InMemoryStore) LoadPolicy(
	_ context.Context, id string,
) (*data.Policy, error) {
	s.policyMu.RLock()
	defer s.policyMu.RUnlock()

	policy, exists := s.policies[id]
	if !exists {
		return nil, nil // Return nil, nil for not found (matching NoopStore behavior)
	}

	return policy, nil
}

// LoadAllPolicies retrieves all policies from the store.
func (s *InMemoryStore) LoadAllPolicies(
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
func (s *InMemoryStore) DeletePolicy(_ context.Context, id string) error {
	s.policyMu.Lock()
	defer s.policyMu.Unlock()

	delete(s.policies, id)
	return nil
}

// GetCipher returns the cipher used for encryption/decryption
func (s *InMemoryStore) GetCipher() cipher.AEAD {
	return s.cipher
}
