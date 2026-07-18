//    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
//  \\\\\ Copyright 2024-present SPIKE contributors.
// \\\\\\\ SPDX-License-Identifier: Apache-2.0

package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/spiffe/spike-sdk-go/api/entity/data"
	"github.com/spiffe/spike-sdk-go/config/env"
	"github.com/spiffe/spike-sdk-go/crypto"
	"github.com/spiffe/spike-sdk-go/security/mem"

	"github.com/spiffe/spike/app/nexus/internal/initialization/recovery"
	state "github.com/spiffe/spike/app/nexus/internal/state/base"
)

// TestMain isolates the package run: the sqlite backend writes into a
// per-run temporary directory (fs.NexusDataFolder memoizes its result,
// so the override must precede the first resolution), and the backend
// store type is pinned to sqlite explicitly since the lifecycle under
// test only exists for persistent backends.
func TestMain(m *testing.M) {
	dir, mkErr := os.MkdirTemp("", "spike-state-integration-test-*")
	if mkErr != nil {
		fmt.Fprintln(os.Stderr,
			"failed to create a temporary data directory:", mkErr)
		os.Exit(1)
	}

	for key, value := range map[string]string{
		env.NexusDataDir:      dir,
		env.NexusBackendStore: "sqlite",
	} {
		if setErr := os.Setenv(key, value); setErr != nil {
			_ = os.RemoveAll(dir)
			fmt.Fprintln(os.Stderr, "failed to set "+key+":", setErr)
			os.Exit(1)
		}
	}

	code := m.Run()

	_ = os.RemoveAll(dir)
	os.Exit(code)
}

const (
	secretPath = "integration/db/creds"
	policyName = "integration-workload-can-read"
)

// TestStateLifecycle walks the state layer through its whole life:
// initialization, secret and policy writes, a duplicate initialization
// (which must not recompute or corrupt anything), the export of
// operator recovery shards, a simulated root-key loss, a shard-based
// restore, and finally proof that the pre-crash data is readable again.
// The stages depend on each other and run in order.
func TestStateLifecycle(t *testing.T) {
	rootKey := &[crypto.AES256KeySize]byte{}
	for i := range rootKey {
		rootKey[i] = byte(i + 1)
	}

	// Stage 1: initialize and verify the root key is cached.
	state.Initialize(rootKey)
	if state.RootKeyZero() {
		t.Fatal("root key is not cached after Initialize")
		return
	}

	// Stage 2: write a secret and a policy through the real stack.
	secretValues := map[string]string{
		"username": "spike",
		"password": "integration-v1",
	}
	if upsertErr := state.UpsertSecret(secretPath, secretValues); upsertErr != nil {
		t.Fatalf("failed to upsert the secret: %v", upsertErr)
		return
	}

	got, getErr := state.GetSecret(secretPath, 0)
	if getErr != nil {
		t.Fatalf("failed to read the secret back: %v", getErr)
		return
	}
	if got["password"] != secretValues["password"] {
		t.Fatalf("secret round trip mismatch: got %q", got["password"])
		return
	}

	if _, policyErr := state.UpsertPolicy(data.Policy{
		Name:            policyName,
		SPIFFEIDPattern: `^spiffe://spike\.ist/workload/.*$`,
		PathPattern:     `^integration/.*$`,
		Permissions:     []data.PolicyPermission{"read"},
	}); policyErr != nil {
		t.Fatalf("failed to upsert the policy: %v", policyErr)
		return
	}

	policy, policyGetErr := state.GetPolicy(policyName)
	if policyGetErr != nil {
		t.Fatalf("failed to read the policy back: %v", policyGetErr)
		return
	}
	if policy.PathPattern != `^integration/.*$` {
		t.Fatalf("policy round trip mismatch: got %q", policy.PathPattern)
		return
	}

	// Stage 3: a duplicate initialization must not recompute root key
	// material or recreate the backend. The observable invariant: the
	// secret written before the duplicate call stays readable, which
	// proves the backend (and the cipher derived from the original
	// key) survived intact.
	state.Initialize(rootKey)
	if state.RootKeyZero() {
		t.Fatal("root key lost after a duplicate Initialize")
		return
	}
	if _, rereadErr := state.GetSecret(secretPath, 0); rereadErr != nil {
		t.Fatalf("secret unreadable after duplicate Initialize: %v", rereadErr)
		return
	}

	// Stage 4: export recovery shards while healthy, as the operator
	// recover flow does, and keep only a threshold-sized subset to
	// prove reconstruction does not need every share.
	shardMap := recovery.NewPilotRecoveryShards()
	threshold := env.ShamirThresholdVal()
	if len(shardMap) < threshold {
		t.Fatalf("expected at least %d shards, got %d",
			threshold, len(shardMap))
		return
	}

	shards := make([]crypto.ShamirShard, 0, threshold)
	for idx, value := range shardMap {
		if len(shards) == threshold {
			break
		}
		shards = append(shards, crypto.ShamirShard{
			ID:    uint64(idx),
			Value: value,
		})
	}

	// Stage 5: simulate the crash by zeroing the cached root key, the
	// in-process equivalent of losing Nexus and every Keeper.
	state.LockRootKey()
	mem.ClearRawBytes(state.RootKeyNoLock())
	state.UnlockRootKey()
	if !state.RootKeyZero() {
		t.Fatal("root key still cached after the simulated crash")
		return
	}

	// Stage 6: restore from the shard subset. The restore path
	// initializes the state first and only then reaches for a SPIFFE
	// source to hydrate the Keepers; that step fails via log.FatalErr,
	// which the panic mode converts into a recoverable panic. The panic
	// is therefore expected here, and it fires after the part under
	// test has completed.
	//
	// A malformed workload API address makes the source creation fail
	// at validation time. With no SPIFFE_ENDPOINT_SOCKET at all,
	// go-spiffe would instead dial the default socket with an
	// unbounded context and hang the test forever (the missing
	// SVID-acquisition timeout is tracked as its own task).
	t.Setenv("SPIFFE_ENDPOINT_SOCKET", "bogus://fail-fast")
	t.Setenv("SPIKE_STACK_TRACES_ON_LOG_FATAL", "true")
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected a panic at the SPIFFE-source boundary" +
					" after the state restore")
			}
		}()
		recovery.RestoreBackingStoreFromPilotShards(shards)
	}()

	// Stage 7: the restore must have recomputed the original root key
	// and the pre-crash data must be readable again.
	if state.RootKeyZero() {
		t.Fatal("root key not restored from shards")
		return
	}

	state.LockRootKey()
	restoredMatches := *state.RootKeyNoLock() == *rootKey
	state.UnlockRootKey()
	if !restoredMatches {
		t.Fatal("restored root key differs from the original")
		return
	}

	restored, restoredErr := state.GetSecret(secretPath, 0)
	if restoredErr != nil {
		t.Fatalf("secret unreadable after the restore: %v", restoredErr)
		return
	}
	if restored["password"] != secretValues["password"] {
		t.Fatalf("secret mismatch after the restore: got %q",
			restored["password"])
		return
	}

	if _, policyRereadErr := state.GetPolicy(policyName); policyRereadErr != nil {
		t.Fatalf("policy unreadable after the restore: %v", policyRereadErr)
		return
	}

	// Stage 8: deletion and undeletion survive the restored state.
	if delErr := state.DeleteSecret(secretPath, []int{1}); delErr != nil {
		t.Fatalf("failed to delete the secret: %v", delErr)
		return
	}
	if _, deletedErr := state.GetSecret(secretPath, 1); deletedErr == nil {
		t.Error("expected an error reading a deleted secret version")
	}

	if undelErr := state.UndeleteSecret(secretPath, []int{1}); undelErr != nil {
		t.Fatalf("failed to undelete the secret: %v", undelErr)
		return
	}
	if _, revivedErr := state.GetSecret(secretPath, 1); revivedErr != nil {
		t.Fatalf("secret unreadable after undelete: %v", revivedErr)
		return
	}
}
