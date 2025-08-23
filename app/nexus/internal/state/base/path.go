//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"github.com/spiffe/spike/app/nexus/internal/state/persist"
	"sort"
)

// ListKeys returns a slice of strings containing all secret paths currently
// stored in the persistence backend. The function loads all secrets from the
// backend and extracts their paths for enumeration.
//
// The function uses a background context for the backend operation. If an error
// occurs while loading secrets from the backend, an empty slice is returned.
// The returned paths are sorted in lexicographical order for consistent
// ordering.
//
// Returns:
//   - []string: A slice containing all secret paths in the backend, sorted
//     lexicographically.
//     Returns an empty slice if there are no secrets or if an error occurs.
//
// Example:
//
//	keys := ListKeys()
//	for _, key := range keys {
//	    fmt.Printf("Found key: %s\n", key)
//	}
func ListKeys() []string {
	ctx := context.Background()

	secrets, err := persist.Backend().LoadAllSecrets(ctx)
	if err != nil {
		return []string{}
	}

	// Extract just the keys
	keys := make([]string, 0, len(secrets))
	for path := range secrets {
		keys = append(keys, path)
	}

	// Sort for consistent ordering (lexicographical)
	sort.Strings(keys)

	return keys
}
