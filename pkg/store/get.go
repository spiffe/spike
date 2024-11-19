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
//   - Returns nil and false if the version has been deleted
//     (DeletedTime is set)
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
func (kv *KV) Get(path string, version int) (map[string]string, error) {
	secret, exists := kv.data[path]
	if !exists {
		return nil, ErrSecretNotFound
	}

	// #region debug
	// fmt.Println("########")
	// vv := secret.Versions
	// for i, v := range vv {
	// 	fmt.Println("version", i, "version:", v.Version, "created:",
	// 		v.CreatedTime, "deleted:", v.DeletedTime, "data:", v.Data)
	// }
	// fmt.Println("########")
	// #endregion

	// If version not specified, use current version
	if version == 0 {
		version = secret.Metadata.CurrentVersion
	}

	v, exists := secret.Versions[version]
	if !exists || v.DeletedTime != nil {
		return nil, ErrSecretSoftDeleted
	}

	return v.Data, nil
}

// GetRawSecret retrieves a raw secret from the store at the specified path.
// This function is similar to Get, but it returns the raw Secret object instead
// of the key-value data map.
//
// Parameters:
// - path: The path to retrieve the secret from
//
// Returns:
//   - *Secret: The secret at the specified path, or nil if it doesn't exist
//     or has been deleted.
func (kv *KV) GetRawSecret(path string) (*Secret, error) {
	secret, exists := kv.data[path]
	if !exists {
		return nil, ErrSecretNotFound
	}

	return secret, nil
}
