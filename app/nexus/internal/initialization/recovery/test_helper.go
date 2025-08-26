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
