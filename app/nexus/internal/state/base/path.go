//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

// ListKeys returns a slice of strings containing all keys currently stored
// in the key-value store. The function acquires a lock on the store to ensure
// a consistent view of the keys during enumeration.
//
// The returned slice contains the paths to all keys, regardless of their
// version status (active or deleted). The paths are returned in lexicographical
// order.
//
// Returns:
//   - []string: A slice containing all key paths in the store
//
// Example:
//
//	keys := ListKeys()
//	for _, key := range keys {
//	    fmt.Printf("Found key: %s\n", key)
//	}
func ListKeys() []string {
	secretStoreMu.Lock()
	defer secretStoreMu.Unlock()
	return secretStore.List()
}
