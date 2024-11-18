//    \\ SPIKE: Secure your secrets with SPIFFE.
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import "github.com/spiffe/spike/app/nexus/internal/state/persist"

func SetAdminCredentials(passwordHash, salt string) {
	adminCredentialsMu.Lock()
	adminCredentials = Credentials{
		PasswordHash: passwordHash,
		Salt:         salt,
	}
	adminCredentialsMu.Unlock()

	persist.AsyncPersistAdminCredentials(Credentials{
		PasswordHash: passwordHash,
		Salt:         salt,
	})
}

func AdminCredentials() Credentials {
	adminCredentialsMu.RLock()
	creds := adminCredentials
	adminCredentialsMu.RUnlock()

	salt := creds.Salt
	hash := creds.PasswordHash

	if salt == "" || hash == "" {
		cachedCreds := persist.ReadAdminCredentials()
		if cachedCreds == nil {
			return creds
		}

		adminCredentialsMu.Lock()
		adminCredentials = *cachedCreds
		adminCredentialsMu.Unlock()

		return *cachedCreds
	}

	// implement database lookup too.
	return creds
}
