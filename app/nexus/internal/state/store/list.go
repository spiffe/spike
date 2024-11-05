//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

// List returns a slice containing all keys stored in the key-value store.
// The order of keys in the returned slice is not guaranteed to be stable
// between calls.
//
// Returns:
//   - []string: A slice containing all keys present in the store
func (kv *KV) List() []string {
	keys := make([]string, 0, len(kv.data))
	for k := range kv.data {
		keys = append(keys, k)
	}
	return keys
}
