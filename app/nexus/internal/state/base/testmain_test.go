//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package base

import (
	"fmt"
	"os"
	"testing"

	"github.com/spiffe/spike-sdk-go/config/env"
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
		fmt.Fprintln(os.Stderr,
			"failed to create a temporary data directory:", mkErr)
		os.Exit(1)
	}

	if setErr := os.Setenv(env.NexusDataDir, dir); setErr != nil {
		_ = os.RemoveAll(dir)
		fmt.Fprintln(os.Stderr, "failed to set "+env.NexusDataDir+":", setErr)
		os.Exit(1)
	}

	code := m.Run()

	_ = os.RemoveAll(dir)
	os.Exit(code)
}
