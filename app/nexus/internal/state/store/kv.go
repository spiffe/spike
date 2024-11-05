//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

// KV represents an in-memory key-value store with versioning
type KV struct {
	data map[string]*Secret
}

// NewKV creates a new KV instance
func NewKV() *KV {
	return &KV{
		data: make(map[string]*Secret),
	}
}
