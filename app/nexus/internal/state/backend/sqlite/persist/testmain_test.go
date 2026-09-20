//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package persist

import (
	"os"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/log"
)

// TestMain points SPIKE_NEXUS_DATA_DIR at a per-run temporary directory
// before any test resolves the Nexus data folder. fs.NexusDataFolder
// memoizes its result with sync.Once, so the override must happen before
// the first call; doing it here isolates the whole package run from the
// real ~/.spike/data directory, whose database these tests used to
// delete out from under a live dev environment.
func TestMain(m *testing.M) {
	const fName = "TestMain"

	dir, mkErr := os.MkdirTemp("", "spike-sqlite-persist-test-*")
	if mkErr != nil {
		log.FatalLn(fName,
			"failed to create a temporary data directory:", mkErr)
	}

	if setErr := os.Setenv(env.NexusDataDir, dir); setErr != nil {
		removeTempDir(dir)
		log.FatalLn(fName, "failed to set "+env.NexusDataDir+":", setErr)
	}

	code := m.Run()

	removeTempDir(dir)
	os.Exit(code)
}
