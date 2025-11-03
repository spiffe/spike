//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"crypto/cipher"
	"database/sql"
	"sync"
)

// DataStore implements the backend.Backend interface providing encrypted storage
// capabilities using SQLite as the underlying database. It uses AES-GCM for
// encryption and implements proper locking mechanisms for concurrent access.
type DataStore struct {
	db         *sql.DB      // Database connection handle
	Cipher     cipher.AEAD  // Encryption Cipher for data protection
	mu         sync.RWMutex // Mutex for thread-safe operations
	closeOnce  sync.Once    // Ensures the database is closed only once
	Opts       *Options     // Configuration options for the data store
	kekManager interface {  // KEK manager for envelope encryption (optional)
		GetCurrentKEKID() string
		GetKEK(kekID string) (*[32]byte, error)
		IncrementWrapsCount(kekID string) error
	}
}
