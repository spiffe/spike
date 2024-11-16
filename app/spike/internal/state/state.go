//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"errors"
	"github.com/spiffe/spike/internal/config"
	"os"
	"sync"
)

var tokenMutex sync.RWMutex

// AdminToken retrieves the admin token from the SpikeAdminTokenFile().
// The function is thread-safe through a read mutex lock.
//
// Returns:
//   - string: The admin token read from the file
//   - error: An error if the file read fails, which includes both the context
//     "failed to read token from file" and the underlying file system error
//
// The function uses os.ReadFile to read the entire contents of the token file.
// Callers should handle both return values as the operation may fail if the
// file is inaccessible or doesn't exist.
//
// The token file is expected to be in the current working directory with
// the name SpikeAdminTokenFile().
func AdminToken() (string, error) {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()

	tokenBytes, err := os.ReadFile(config.SpikePilotAdminTokenFile())
	if err != nil {
		return "", errors.Join(
			errors.New("failed to read token from file"),
			err,
		)
	}

	return string(tokenBytes), nil
}
