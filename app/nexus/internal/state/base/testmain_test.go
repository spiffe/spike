//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

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
	dir, mkErr := os.MkdirTemp("", "spike-state-base-test-*")
	if mkErr != nil {
		log.FatalLn("TestMain",
			"message", "failed to create a temporary data directory",
			"err", mkErr.Error())
	}

	if setErr := os.Setenv(env.NexusDataDir, dir); setErr != nil {
		if rmErr := os.RemoveAll(dir); rmErr != nil {
			log.Warn("TestMain",
				"message", "failed to remove the temporary data directory",
				"path", dir, "err", rmErr.Error())
		}
		log.FatalLn("TestMain",
			"message", "failed to set "+env.NexusDataDir,
			"err", setErr.Error())
	}

	code := m.Run()

	if rmErr := os.RemoveAll(dir); rmErr != nil {
		log.Warn("TestMain",
			"message", "failed to remove the temporary data directory",
			"path", dir, "err", rmErr.Error())
	}
	os.Exit(code)
}
