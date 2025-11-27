//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package memory

import (
	"context"
	"crypto/cipher"
	"errors"
	"sync"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	sdkErrors "github.com/spiffe/spike-sdk-go/errors"
	"github.com/spiffe/spike-sdk-go/kv"
)

// Store provides an in-memory implementation of a storage backend.
//
// This implementation stores data in memory using the kv package for secrets
// and a map for policies. It is fully functional and thread-safe, making it
// suitable for development, testing, or scenarios where persistent storage is
// not required. All data is lost when the process terminates.
//
// The store uses separate read-write mutexes for secrets and policies to
// allow concurrent reads while ensuring exclusive writes.
type Store struct {
	secretStore *kv.KV       // In-memory key-value store for secrets
	secretMu    sync.RWMutex // Mutex protecting secret operations

	policies map[string]*data.Policy // In-memory map of policies by ID
	policyMu sync.RWMutex            // Mutex protecting policy operations

	cipher cipher.AEAD // Encryption cipher (for interface compatibility)
}

// NewInMemoryStore creates a new in-memory store instance.
//
// The store is immediately ready for use and requires no additional
// initialization. Secret versioning is configured according to the
// maxVersions parameter.
//
// Parameters:
//   - cipher: The encryption cipher (stored for interface compatibility but
//     not used for in-memory encryption)
//   - maxVersions: Maximum number of versions to retain per secret
//
// Returns:
//   - *Store: An initialized in-memory store ready for use
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
//
// For the in-memory implementation, this is a no-op since the store is fully
// initialized in the constructor. This method exists to satisfy the backend
// interface.
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) Initialize(_ context.Context) *sdkErrors.SDKError {
	// Already initialized in constructor
	return nil
}

// Close implements the closing operation for the store.
//
// For the in-memory implementation, this is a no-op since there are no
// resources to release. All data is simply garbage collected when the store
// is no longer referenced. This method exists to satisfy the backend
// interface.
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil
func (s *Store) Close(_ context.Context) *sdkErrors.SDKError {
	// Nothing to close for in-memory store
	return nil
}

// StoreSecret saves a secret to the store at the specified path.
//
// This method is thread-safe and stores the complete secret structure,
// including all versions and metadata. If a secret already exists at the
// path, it is replaced entirely.
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//   - path: The secret path where the secret should be stored
//   - secret: The complete secret value including metadata and versions
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil for in-memory storage
func (s *Store) StoreSecret(
	_ context.Context, path string, secret kv.Value,
) *sdkErrors.SDKError {
	s.secretMu.Lock()
	defer s.secretMu.Unlock()

	// Store the entire secret structure
	s.secretStore.ImportSecrets(map[string]*kv.Value{
		path: &secret,
	})

	return nil
}

// LoadSecret retrieves a secret from the store by its path.
//
// This method is thread-safe and returns the complete secret structure,
// including all versions and metadata.
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//   - path: The secret path to retrieve
//
// Returns:
//   - *kv.Value: The secret with all its versions and metadata, or nil if not
//     found
//   - *sdkErrors.SDKError: nil on success, sdkErrors.ErrEntityNotFound if the
//     secret does not exist, or an error if retrieval fails
func (s *Store) LoadSecret(
	_ context.Context, path string,
) (*kv.Value, *sdkErrors.SDKError) {
	s.secretMu.RLock()
	defer s.secretMu.RUnlock()

	rawSecret, err := s.secretStore.GetRawSecret(path)
	if err != nil && errors.Is(err, sdkErrors.ErrEntityNotFound) {
		return nil, sdkErrors.ErrEntityNotFound
	} else if err != nil {
		return nil, err
	}

	return rawSecret, nil
}

// LoadAllSecrets retrieves all secrets stored in the store.
//
// This method is thread-safe and returns a map of all secrets currently in
// memory. If any individual secret fails to load (which should not happen in
// normal operation), it is silently skipped.
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//
// Returns:
//   - map[string]*kv.Value: A map of secret paths to their values
//   - *sdkErrors.SDKError: Always returns nil for in-memory storage
func (s *Store) LoadAllSecrets(_ context.Context) (
	map[string]*kv.Value, *sdkErrors.SDKError,
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
//
// This method is thread-safe and validates that the policy has a non-empty ID
// before storing. If a policy with the same ID already exists, it is
// replaced.
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//   - policy: The policy to store
//
// Returns:
//   - *sdkErrors.SDKError: nil on success, or sdkErrors.ErrEntityInvalid if
//     the policy ID is empty
func (s *Store) StorePolicy(
	_ context.Context, policy data.Policy,
) *sdkErrors.SDKError {
	s.policyMu.Lock()
	defer s.policyMu.Unlock()

	if policy.ID == "" {
		failErr := *sdkErrors.ErrEntityInvalid.Clone()
		failErr.Msg = "policy ID cannot be empty"
		return &failErr
	}

	s.policies[policy.ID] = &policy
	return nil
}

// LoadPolicy retrieves a policy from the store by its ID.
//
// This method is thread-safe and returns the policy if it exists.
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//   - id: The unique identifier of the policy to retrieve
//
// Returns:
//   - *data.Policy: The policy if found, nil otherwise
//   - *sdkErrors.SDKError: nil on success, or sdkErrors.ErrEntityNotFound if
//     the policy does not exist
func (s *Store) LoadPolicy(
	_ context.Context, id string,
) (*data.Policy, *sdkErrors.SDKError) {
	s.policyMu.RLock()
	defer s.policyMu.RUnlock()

	policy, exists := s.policies[id]
	if !exists {
		return nil, sdkErrors.ErrEntityNotFound
	}

	return policy, nil
}

// LoadAllPolicies retrieves all policies from the store.
//
// This method is thread-safe and returns a copy of the policies map to avoid
// race conditions if the caller modifies the returned map.
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//
// Returns:
//   - map[string]*data.Policy: A map of policy IDs to policies
//   - *sdkErrors.SDKError: Always returns nil for in-memory storage
func (s *Store) LoadAllPolicies(
	_ context.Context,
) (map[string]*data.Policy, *sdkErrors.SDKError) {
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
//
// This method is thread-safe and removes the policy if it exists. If the
// policy does not exist, this is a no-op (no error is returned).
//
// Parameters:
//   - context.Context: Context for cancellation (ignored in this
//     implementation)
//   - id: The unique identifier of the policy to delete
//
// Returns:
//   - *sdkErrors.SDKError: Always returns nil for in-memory storage
func (s *Store) DeletePolicy(_ context.Context, id string) *sdkErrors.SDKError {
	s.policyMu.Lock()
	defer s.policyMu.Unlock()

	delete(s.policies, id)
	return nil
}

// GetCipher returns the cipher used for encryption/decryption.
//
// For the in-memory implementation, this cipher is stored for interface
// compatibility but is not used for encryption since data is kept
// in memory in plaintext.
//
// Returns:
//   - cipher.AEAD: The cipher provided during initialization
func (s *Store) GetCipher() cipher.AEAD {
	return s.cipher
}
