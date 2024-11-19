//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

// KV represents an in-memory key-value store with versioning
type KV struct {
	maxSecretVersions int
	data              map[string]*Secret
}

// KVConfig represents the configuration for a KV instance
type KVConfig struct {
	MaxSecretVersions int
}

// NewKV creates a new KV instance
func NewKV(config KVConfig) *KV {
	return &KV{
		maxSecretVersions: config.MaxSecretVersions,
		data:              make(map[string]*Secret),
	}
}
