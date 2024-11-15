//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"strconv"
)

// Pbkdf2IterationCount returns the number of iterations to use for PBKDF2
// password hashing. It retrieves the value from the
// NEXUS_PBKDF2_ITERATION_COUNT environment variable.
// If the environment variable is not set or invalid, it returns the
// OWASP-recommended minimum of 600,000 iterations for PBKDF2-SHA256.
//
// Returns:
//   - int: The number of PBKDF2 iterations to use, minimum 600,000
func Pbkdf2IterationCount() int {
	p := os.Getenv("SPIKE_NEXUS_PBKDF2_ITERATION_COUNT")

	c, err := strconv.Atoi(p)
	if err == nil && c > 0 {
		return c
	}

	// Minimum OWASP recommendation for PBKDF2-SHA256
	return 600000
}

// ShaHashLength returns the length in bytes to use for SHA hash operations.
// It retrieves the value from the NEXUS_SHA_HASH_LENGTH environment variable.
// If the environment variable is not set or invalid, it returns the default
// value of 32 bytes (256 bits).
//
// Returns:
//   - int: The SHA hash length in bytes, default 32 (256 bits)
func ShaHashLength() int {
	p := os.Getenv("SPIKE_NEXUS_SHA_HASH_LENGTH")

	c, err := strconv.Atoi(p)
	if err == nil && c > 0 {
		return c
	}

	// Default to 256 bits
	return 32
}
