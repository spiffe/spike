//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"sync"

	"github.com/spiffe/spike-sdk-go/crypto"
)

var (
	// rootKeySeed stores the root key seed generated during initialization.
	// It is kept in memory to allow encryption operations during bootstrap.
	rootKeySeed [crypto.AES256KeySize]byte
	// rootKeySeedMu provides mutual exclusion for access to the root key seed.
	rootKeySeedMu sync.RWMutex
)
