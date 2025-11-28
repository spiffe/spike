//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"sync"
	"sync/atomic"

	"github.com/spiffe/spike/app/nexus/internal/state/backend"
)

var (
	// be is the primary backend instance used for state persistence.
	// This variable is initialized once during InitializeBackend and should
	// not be accessed directly. Use Backend() to obtain a safe reference.
	be backend.Backend

	// backendMu protects the initialization of the backend.
	// It is used during InitializeBackend to ensure only one goroutine
	// initializes the backend at a time.
	backendMu sync.RWMutex

	// backendPtr is an atomic pointer to the current backend.
	// This allows safe concurrent reads after initialization without
	// holding a lock for the duration of backend operations.
	backendPtr atomic.Pointer[backend.Backend]
)
