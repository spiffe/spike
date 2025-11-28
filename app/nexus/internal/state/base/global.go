//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"sync"

	"github.com/spiffe/spike-sdk-go/crypto"
)

// Global variables related to the root key with thread-safety protection.
var (
	// rootKey is a 32-byte array that stores the cryptographic root key.
	// It is initialized to zeroes by default.
	rootKey [crypto.AES256KeySize]byte
	// rootKeyMu provides mutual exclusion for access to the root key.
	rootKeyMu sync.RWMutex
)
