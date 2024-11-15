//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package env

import (
	"os"
	"strconv"
)

func Pbkdf2IterationCount() int {
	p := os.Getenv("NEXUS_PBKDF2_ITERATION_COUNT")

	c, err := strconv.Atoi(p)
	if err == nil && c > 0 {
		return c
	}

	// Minimum OWASP recommendation for PBKDF2-SHA256
	return 600000
}

func ShaHashLength() int {
	p := os.Getenv("NEXUS_SHA_HASH_LENGTH")

	c, err := strconv.Atoi(p)
	if err == nil && c > 0 {
		return c
	}

	// Default to 256 bits
	return 32
}

// TODO: add to docs and env.sh too.
