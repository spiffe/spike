//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package store

// Get retrieves a versioned key-value data map from the store at the specified
// path.
//
// The function supports versioned data retrieval with the following behavior:
//   - If version is 0, returns the current version of the data
//   - If version is specified, returns that specific version if it exists
//   - Returns nil and false if the path doesn't exist
//   - Returns nil and false if the specified version doesn't exist
//   - Returns nil and false if the version has been deleted (DeletedTime is set)
//
// Parameters:
//   - path: The path to retrieve data from
//   - version: The specific version to retrieve (0 for current version)
//
// Returns:
//   - map[string]string: The key-value data at the specified path and version
//   - bool: true if data was found and is valid, false otherwise
//
// Example usage:
//
//	kv := &KV{}
//	// Get current version
//	data, exists := kv.Get("secret/myapp", 0)
//
//	// Get specific version
//	historicalData, exists := kv.Get("secret/myapp", 2)
func (kv *KV) Get(path string, version int) (map[string]string, bool) {
	secret, exists := kv.data[path]
	if !exists {
		return nil, false
	}

	// If version not specified, use current version
	if version == 0 {
		version = secret.Metadata.CurrentVersion
	}

	v, exists := secret.Versions[version]
	if !exists || v.DeletedTime != nil {
		return nil, false
	}

	return v.Data, true
}
