//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import "net/url"

// Helper function for URL path checking
func containsPathUpdate(fullURL, path string) bool {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return false
	}

	// Clean the path from leading/trailing slashes for comparison
	cleanPath := path
	if len(cleanPath) > 0 && cleanPath[0] == '/' {
		cleanPath = cleanPath[1:]
	}

	return len(parsedURL.Path) > 0 &&
		(parsedURL.Path[len(parsedURL.Path)-len(cleanPath):] == cleanPath ||
			parsedURL.Path == "/"+cleanPath)
}

// containsBase reports whether a full URL starts with the given base URL.
//
// Parameters:
//   - fullURL: The complete URL under test.
//   - base: The expected prefix.
//
// Returns:
//   - bool: True when fullURL begins with base.
func containsBase(fullURL, base string) bool {
	return len(fullURL) >= len(base) && fullURL[:len(base)] == base
}

// containsPath reports whether a full URL's path ends with the given path,
// ignoring a leading slash on the expected path.
//
// Parameters:
//   - fullURL: The complete URL under test.
//   - path: The expected path suffix.
//
// Returns:
//   - bool: True when the URL parses and its path carries the suffix.
func containsPath(fullURL, path string) bool {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return false
	}

	// Clean the path from leading/trailing slashes for comparison
	cleanPath := path
	if len(cleanPath) > 0 && cleanPath[0] == '/' {
		cleanPath = cleanPath[1:]
	}

	return len(parsedURL.Path) > 0 &&
		(parsedURL.Path[len(parsedURL.Path)-len(cleanPath):] == cleanPath ||
			parsedURL.Path == "/"+cleanPath)
}
