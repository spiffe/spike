//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package recovery

import (
	"testing"
	"time"

	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/security/mem"

	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// TestInitializeBackingStoreFromKeepers_RootKeyAlreadyPresent verifies
// that the Keeper recovery loop stands down promptly when the root key
// is already present, which is what happens when an operator restores
// the system through the emergency restore route while the loop is
// still retrying. Before the standdown check, this call would retry
// forever (the nil source alone never stops it), so the test guards
// the boot path that keeps the restore flow deadlock-free.
func TestInitializeBackingStoreFromKeepers_RootKeyAlreadyPresent(
	t *testing.T,
) {
	rk := &[crypto.AES256KeySize]byte{}
	for i := range rk {
		rk[i] = byte(i + 1)
	}
	state.SetRootKey(rk)

	// SetRootKey rejects zero keys by design, so zero the key directly
	// under the lock, the same way the state package's own tests reset it.
	t.Cleanup(func() {
		state.LockRootKey()
		defer state.UnlockRootKey()
		mem.ClearRawBytes(state.RootKeyNoLock())
	})

	done := make(chan struct{})
	go func() {
		InitializeBackingStoreFromKeepers(nil)
		close(done)
	}()

	select {
	case <-done:
		// The loop noticed the root key and returned.
	case <-time.After(5 * time.Second):
		t.Fatal("InitializeBackingStoreFromKeepers did not stand down" +
			" although the root key is present")
	}
}
