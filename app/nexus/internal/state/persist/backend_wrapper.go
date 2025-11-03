//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"context"
	"crypto/cipher"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/kv"
	"github.com/spiffe/spike-sdk-go/log"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
	"github.com/spiffe/spike/app/nexus/internal/state/backend/sqlite/persist"
	"github.com/spiffe/spike/internal/config"
)

// EnvelopeAwareBackend wraps a backend and automatically uses envelope
// encryption when KEK manager is available
type EnvelopeAwareBackend struct {
	backend backend.Backend
}

// NewEnvelopeAwareBackend creates a backend wrapper that uses envelope
// encryption when available
func NewEnvelopeAwareBackend(be backend.Backend) *EnvelopeAwareBackend {
	return &EnvelopeAwareBackend{backend: be}
}

// LoadSecret loads a secret, automatically using envelope encryption if available
func (e *EnvelopeAwareBackend) LoadSecret(ctx context.Context, path string) (*kv.Value, error) {
	const fName = "EnvelopeAwareBackend.LoadSecret"

	// Check if KEK manager is initialized and enabled
	if config.KEKRotationEnabled() && IsKEKManagerInitialized() {
		// Try to use envelope encryption
		if sqliteBackend, ok := e.backend.(*persist.DataStore); ok {
			secret, needsRewrap, err := sqliteBackend.LoadSecretWithEnvelope(ctx, path)
			if err != nil {
				return nil, err
			}

			// If lazy rewrap is enabled and needed, schedule it
			if needsRewrap && config.KEKLazyRewrapEnabled() {
				// Get secret metadata to find which versions need rewrapping
				if secret != nil {
					for version := range secret.Versions {
						// Schedule background rewrap
						go func(v int) {
							if err := sqliteBackend.RewrapSecret(ctx, path, v); err != nil {
								log.Log().Error(fName,
									"message", "background rewrap failed",
									"path", path,
									"version", v,
									"err", err.Error())
							}
						}(version)
					}
				}
			}

			return secret, nil
		}
	}

	// Fall back to regular load (no envelope encryption)
	return e.backend.LoadSecret(ctx, path)
}

// StoreSecret stores a secret, automatically using envelope encryption if available
func (e *EnvelopeAwareBackend) StoreSecret(ctx context.Context, path string, secret kv.Value) error {
	// Check if KEK manager is initialized and enabled
	if config.KEKRotationEnabled() && IsKEKManagerInitialized() {
		// Try to use envelope encryption
		if sqliteBackend, ok := e.backend.(*persist.DataStore); ok {
			return sqliteBackend.StoreSecretWithEnvelope(ctx, path, secret)
		}
	}

	// Fall back to regular store (no envelope encryption)
	return e.backend.StoreSecret(ctx, path, secret)
}

// LoadAllSecrets delegates to the wrapped backend
func (e *EnvelopeAwareBackend) LoadAllSecrets(ctx context.Context) (map[string]*kv.Value, error) {
	return e.backend.LoadAllSecrets(ctx)
}

// LoadPolicy delegates to the wrapped backend
func (e *EnvelopeAwareBackend) LoadPolicy(ctx context.Context, id string) (*data.Policy, error) {
	return e.backend.LoadPolicy(ctx, id)
}

// StorePolicy delegates to the wrapped backend
func (e *EnvelopeAwareBackend) StorePolicy(ctx context.Context, policy data.Policy) error {
	return e.backend.StorePolicy(ctx, policy)
}

// LoadAllPolicies delegates to the wrapped backend
func (e *EnvelopeAwareBackend) LoadAllPolicies(ctx context.Context) (map[string]*data.Policy, error) {
	return e.backend.LoadAllPolicies(ctx)
}

// DeletePolicy delegates to the wrapped backend
func (e *EnvelopeAwareBackend) DeletePolicy(ctx context.Context, id string) error {
	return e.backend.DeletePolicy(ctx, id)
}

// Initialize delegates to the wrapped backend
func (e *EnvelopeAwareBackend) Initialize(ctx context.Context) error {
	return e.backend.Initialize(ctx)
}

// Close delegates to the wrapped backend
func (e *EnvelopeAwareBackend) Close(ctx context.Context) error {
	return e.backend.Close(ctx)
}

// GetCipher delegates to the wrapped backend
func (e *EnvelopeAwareBackend) GetCipher() cipher.AEAD {
	return e.backend.GetCipher()
}
