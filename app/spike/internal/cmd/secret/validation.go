//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package secret

import "regexp"

// validPath is the regular expression pattern used to validate secret path
// formats. It allows alphanumeric characters, dots, underscores, hyphens,
// slashes, and various special characters commonly used in path notation.
const validPath = `^[a-zA-Z0-9._\-/()?+*|[\]{}\\]+$`

// validSecretPath validates whether a given path conforms to the allowed
// secret path format. The path must match the validPath regular expression,
// which permits alphanumeric characters and common path separators.
//
// Parameters:
//   - path: The secret path string to validate
//
// Returns:
//   - true if the path is valid, according to the pattern
//   - false if the path contains invalid characters or is malformed
//
// Note: This validation is performed client-side for early error detection.
// The server may perform additional validation.
func validSecretPath(path string) bool {
	return regexp.MustCompile(validPath).MatchString(path)
}
