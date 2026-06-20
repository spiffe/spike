# Spec: Fix the minio-rolearn integration test (keeper bootstrap)

## Problem Statement

The `Integration Test` CI job (`ci/integration/minio-rolearn`) is red and
has been on `main` for many commits. SPIKE Nexus never becomes Ready, so
`helm --wait` / the rollout times out.

Root cause chain (reproduced locally on kind):

1. **Keepers are never seeded.** Nexus's
   `InitializeBackingStoreFromKeepers` loops forever (`retry.Forever`)
   waiting for keeper shards. The spire chart registers the
   `spike/bootstrap` identity but ships **no** bootstrap workload, and
   `setup.sh` never seeds the keepers. So Nexus never initializes,
   never binds `:8443`, and the startup probe kills it on a loop.
2. Once a bootstrap workload is added, three further bugs surfaced:
   - **Deadlock:** `bootstrap.VerifyInitialization` took the write lock
     (`LockRootKeySeed`) then a read lock (`RootKeySeed`) on the same
     non-reentrant `sync.RWMutex` -> self-deadlock.
   - **Wrong verify key:** the verify probe was encrypted with the raw
     seed, but Keepers/Nexus key their cipher with the *canonical* root
     key (`ComputeShares(seed)` scalar marshalled to bytes). The seed
     and the canonical key differ (the seed is reduced mod the P256
     order), so Nexus could never decrypt the probe.
   - **Missing route in lite mode:** Nexus registers
     `/v1/bootstrap/verify` only in `routeWithBackingStore`. In lite
     mode (`routeWithNoBackingStore`) the route fell through to
     `net.Fallback` -> HTTP 400, even though lite mode holds the root
     key and already exposes the other root-key routes (cipher,
     operator recover/restore). The bootstrap verifies unconditionally,
     so the omission left every verify failing.

## Proposed Solution

1. **Seed the keepers.** Add a `spike-bootstrap` Job
   (`ci/integration/minio-rolearn/bootstrap.yaml`) with the
   `component=spike-bootstrap` labels (so SPIRE issues the
   `spike/bootstrap` SVID), the SPIRE CSI socket, keeper-peer / Shamir /
   trust-root / `SPIKE_NEXUS_API_URL` env, and a ServiceAccount + Role +
   RoleBinding for its `spike-bootstrap-state` idempotency ConfigMap.
   Wire it into `setup.sh`: drop `--wait` on the spire install (Nexus
   can't be Ready pre-seed) and `kubectl apply` the Job.
2. **Fix the deadlock** in `app/bootstrap/internal/net/broadcast.go`:
   use `RootKeySeedNoLock` (the write lock is already held).
3. **Fix the verify key**: derive the canonical root key with
   `crypto.ComputeShares(seed)` + `MarshalBinary()` and encrypt the
   probe with it.
4. **Register verify in lite mode**: add `NexusBootstrapVerify ->
   bootstrap.RouteVerify` to `routeWithNoBackingStore` in
   `app/nexus/internal/route/base/impl.go`.

## File Surface

- `ci/integration/minio-rolearn/bootstrap.yaml` (new): bootstrap Job +
  SA/RBAC.
- `ci/integration/minio-rolearn/setup.sh`: drop `--wait`, apply Job.
- `app/bootstrap/internal/net/broadcast.go`: deadlock + canonical-key.
- `app/nexus/internal/route/base/impl.go`: lite-mode verify route.

## Non-Goals

- The `cel` CredentialComposer fails on local arm64 (upstream chart);
  only relevant to local reproduction, not committed. CI runs amd64.

## Verification

- Local kind run: bootstrap Job `succeeded=1`, `spike-bootstrap-state`
  ConfigMap `bootstrap-completed=true`, Nexus `1/1` Ready, in ~30s with
  production images. Verified end to end.
- `make test` + `make audit` pass.
