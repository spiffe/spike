# Spec: Bring the SPIRE charts and the bare-metal SPIRE build up to date

## Problem Statement

The Kubernetes install scripts pin the `spire` Helm chart at 0.26.1
(August 2025, SPIRE 1.12.4) and `spire-crds` at 0.5.0 (October 2024);
upstream is at 0.30.2 (SPIRE 1.15.3) and 0.6.1. The bare-metal build
script compiles SPIRE v1.11.2. The CI integration path installs the
chart unpinned, so it already floats to 0.30.2 while the documented
install paths lag four minors behind.

Checking the pins surfaced broken windows:

- `hack/k8s/spike-install.sh` passes the `spire` chart version to the
  `spire-crds` chart, which has no such version; the script fails at
  its first helm command.
- `hack/k8s/spike-dev-install.sh` defaults to local charts behind a
  FIXME that says "until upstream merges" (it merged long ago) and
  tests the flag with a non-empty check, so `false` still selects local
  charts.
- The CI values set `extraInitContainers` on the Nexus and Keeper
  subcharts; neither chart version has ever rendered that key.
- The CI setup applies its own bootstrap Job and patches two
  environment variables into the Nexus StatefulSet behind a FIXME. The
  0.30.2 `spike-nexus` subchart runs a bootstrap Helm hook by default
  and exposes `backendStore` as a value, so the Job duplicates the hook
  and half the patch is redundant.
- `hack/qa/cover.sh` runs tests serially behind a FIXME about
  concurrent isolation that was resolved in July.
- `values-demo-mgmt.yaml` sets `spike-nexus.trustRoot.keepers`, which
  0.30.2 renamed to `trustRoot.keeper`.

## Proposed Solution

1. Pin `spire` 0.30.2 and `spire-crds` 0.6.1 in both install scripts
   (each chart with its own version variable) and in the CI setup.
2. `spike-dev-install.sh`: default `SPIKE_USE_LOCAL_CHARTS` to `false`
   and compare it to `"true"`; drop the FIXME.
3. CI: remove the dead `extraInitContainers` blocks; run the chart's
   own bootstrap hook with the dev image loaded into kind
   (`spike-nexus.bootstrap.image` with `pullPolicy: Never`, `tag: dev`),
   which also exercises the bootstrap image, which the old setup never
   did; delete `bootstrap.yaml`; keep only the
   `SPIKE_TRUST_ROOT_LITE_WORKLOAD` patch, which has no chart value yet
   (upstream ask tracked in TASKS.md), with a plain comment.
4. `build-spire.sh`: SPIRE v1.15.3, matching the chart's appVersion;
   the bare-metal doc clones the same tag.
5. `values-demo-mgmt.yaml`: `trustRoot.keeper`.
6. `cover.sh`: drop `-p 1` and the FIXME.
7. Quickstart: pin the two chart versions in the helm commands.

8. One source of truth: `hack/lib/versions.sh` declares the three
   versions; the two install scripts, the CI setup, and the SPIRE build
   script source it. `internal/layout` gains a test that fails when a
   script stops sourcing it or when the quickstart or bare-metal page
   shows a different version. `make spire-versions` prints the pins. A
   bump is one edit plus the doc lines the test names.

## File Surface

- `hack/lib/versions.sh` (new), `makefiles/Kubernetes.mk`
  (`spire-versions`), `internal/layout/versions_test.go` (new)
- `app/bootstrap/internal/net/readiness.go` (new, `waitForKeepers`),
  `broadcast.go`; `app/nexus/internal/initialization/recovery/`
  `recovery.go`, `transform.go` (`resetShards`); tests for both
- `ci/integration/minio-rolearn/test.yaml` (assets chosen by the pod's
  architecture), `test.sh` (CA file name matches `helper.conf`)
- `hack/k8s/spike-install.sh`, `hack/k8s/spike-dev-install.sh`
- `ci/integration/minio-rolearn/setup.sh`, `spire-values.yaml`;
  `bootstrap.yaml` removed
- `config/helm/values-demo-mgmt.yaml`
- `hack/bare-metal/build/build-spire.sh`, `hack/qa/cover.sh`
- `docs-src/content/getting-started/quickstart.md`,
  `docs-src/content/development/bare-metal.md`,
  `docs-src/content/recipes/bootstrapping-spike.md` (the Kubernetes
  step now describes the chart hook)

## Chart Behavior Change Found by Validation

From 0.30.x the `spire-server` subchart no longer enables the SPIKE
identities by default: `controllerManager.identities.clusterSPIFFEIDs.
spike-{nexus,keeper,pilot,bootstrap}` render only when a values file
sets `enabled: true` on each. Without them the SPIRE agent answers "No
identity issued" and Nexus and Keeper die on their startup probes. The
repository's `config/helm` values already set the flags; the CI values
and the quickstart example relied on the old defaults and now set them.
The ID templates also became per-pod (`spike/nexus/<pod>`), which the
SDK predicates accept by prefix.

## SPIKE Defects Found by Validation

The chart's post-install bootstrap hook starts the instant the install
returns, before the Keepers hold SVIDs. That exposed two defects in
SPIKE itself, both fixed here:

- **Bootstrap seeded a subset and gave up.** A first attempt reached only
  the Keepers already listening, exhausted its per-Keeper retries, and
  exited; the Job retry generated a new root key and seeded the rest. The
  Keepers then held shares of two keys. `BroadcastKeepers` now waits,
  before producing a single share, until every Keeper accepts a TCP
  connection (`waitForKeepers`, bounded by
  `SPIKE_BOOTSTRAP_KEEPER_TIMEOUT` per Keeper). A Keeper listens only
  once it has its SVID, so this is the earliest moment a contribution can
  succeed, and a timeout exits with nothing sent, so a retry is safe.
- **Nexus combined shares from different rounds.** The recovery loop kept
  every share it had ever fetched, so a share from an earlier bootstrap
  was combined with a fresh one, reconstructing a wrong key that Nexus
  then trusted, answering every verify call with "decryption failed".
  Each round now starts from an empty, zeroed set (`resetShards`).

With both fixes the chart hook is the bootstrap path for CI and for
users, and the hand-written CI Job is gone. The CI values point the
hook at the dev bootstrap image loaded into kind, so that image is under
test as well.

## Error / Edge Cases

- **arm64 hosts cannot run the CI path as is.** Both chart versions
  hard-code one sha256 for the `cel` credential composer plugin, and it
  is the amd64 binary's; on arm64 spire-server refuses the plugin and
  crash-loops. CI is amd64 and unaffected. Local validation on arm64
  needs a values overlay with the arm64 checksum (see LEARNINGS.md);
  the durable fix is per-architecture checksums upstream (TASKS.md).

## Compatibility Evidence

Checked against the 0.26.1 and 0.30.2 chart archives, not release
notes: every values key this repository uses still exists except the
renamed `trustRoot.keepers`; the free-form maps (cluster SPIFFE IDs,
federated trust domains) are unchanged; SPIRE removed `k8s_sat`,
`retry_rebootstrap`, and two rego configurables between 1.12 and 1.15,
none of which the bare-metal configuration uses (join_token, disk,
unix, sql, memory plugins). The 1.15 CLI JSON output change affects
nothing here.

## Non-Goals

- No change to the kind node image or Kubernetes version in CI.
- No change to the MinIO chart pin (`specs/minio-bitnami-image-
  relocation.md` owns it).

## Verification (2026-09-20, kind v0.29.0, node v1.33.1, arm64 host)

- `make test` (race, uncached): 25 packages ok, 377 passes, 0 failures;
  `make audit`: exit 0, 0 lint issues, no called vulnerabilities.
- Four runs of the CI integration path with the dev images:
  1. spire-server crash-looped: the chart's `cel` plugin checksum is the
     amd64 binary's (arm64 host limitation; a local overlay with the
     arm64 checksum was used from run 2 on, never committed).
  2. Nexus and Keeper died on their startup probes: no ClusterSPIFFEID
     for SPIKE components. Fixed by enabling the identities in the CI
     values and the quickstart example.
  3. SPIRE and SPIKE pods healthy, but the bootstrap hook timed out:
     eight bootstrap attempts, Keepers holding shares of two keys, Nexus
     answering verify with "decryption failed". Fixed in SPIKE (above).
  4. Release `deployed`, bootstrap hook complete on its first attempt
     with the marker ConfigMap written ~90 s after install, Nexus up
     with zero restarts, all three Keepers running. The test pod fetched
     its SVID and the encrypt then decrypt round trip through Nexus
     returned the plaintext, using the committed test script's own
     commands.
- Not validated here: the MinIO half of the harness (`mc` provisioning
  Job and the S3 copy steps). The MinIO provisioning Job exhausted its
  retries on this arm64 host, where the legacy Bitnami images run under
  emulation; CI runs on amd64. The harness fixes made in passing (the CA
  file name, architecture-aware asset downloads) apply on both.
