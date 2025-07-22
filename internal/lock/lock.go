//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0
package lock

import (
    "os"
    "sync/atomic"
)

var locked atomic.Bool

const lockFile = "/var/lib/spike/locked" 

// LockFile is the file used to indicate that SPIKE is locked.
func init() {
    if _, err := os.Stat(lockFile); err == nil {
        locked.Store(true)
    }
}

// Lock sets the SPIKE lock by creating a lock file and marking it as locked.
func Lock() error {
    locked.Store(true)
    return os.WriteFile(lockFile, []byte("locked"), 0600)
}

// Unlock clears the SPIKE lock by removing the lock file.
func Unlock() error {
    locked.Store(false)
    return os.Remove(lockFile)
}

// IsLocked checks if SPIKE is currently locked.
func IsLocked() bool {
    return locked.Load()
}
