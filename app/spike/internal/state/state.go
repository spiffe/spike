//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package state

import (
	"errors"
	"os"
	"sync"

	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/spiffe/spike/app/spike/internal/net"
)

var tokenMutex sync.RWMutex

// AdminToken
// @deprecated
func AdminToken() (string, error) {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()

	// Try to read from file:
	tokenBytes, err := os.ReadFile(".spike-admin-token")
	if err != nil {
		return "", errors.Join(
			errors.New("failed to read token from file"),
			err,
		)
	}

	return string(tokenBytes), nil
}

// SaveAdminToken
// @deprecated
func SaveAdminToken(source *workloadapi.X509Source, token string) error {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	// Save token to file:
	err := os.WriteFile(".spike-admin-token", []byte(token), 0600)
	if err != nil {
		return errors.Join(errors.New("failed to save token to file"), err)
	}

	// Save the token to SPIKE Nexus
	// This token will be used for Nexus to generate
	// short-lived session tokens for the admin user.
	err = net.SendInitRequest(source, token)
	if err != nil {
		return errors.Join(errors.New("failed to save token to nexus"), err)
	}

	return nil
}

func AdminTokenExists() bool {
	tokenMutex.RLock()
	defer tokenMutex.RUnlock()
	token, _ := AdminToken()
	return token != ""
}
