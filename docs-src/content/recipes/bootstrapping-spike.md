+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Bootstrapping a Fresh SPIKE"
weight = 2
sort_by = "weight"
+++

# Bootstrapping a Fresh SPIKE

## Problem

On a brand-new `lite` or `sqlite` deployment the SPIKE Keepers start **empty**.
SPIKE Nexus recovers its root key *from* the keepers, so until something
generates a root key and seeds the keepers, Nexus can never initialize; it
loops in keeper recovery and its readiness probe keeps killing it. Bootstrap is
the one-time step that breaks this chicken-and-egg.

> `memory` mode needs no bootstrap, so skip this recipe.

## TL;DR

SPIKE Bootstrap generates a 256-bit root key, splits it into Shamir shares, and
distributes one share to each keeper. Run it once after the keepers are up:

```bash
# bare-metal
make bootstrap

# Kubernetes: run the spike-bootstrap workload (Job) with the bootstrap SVID
kubectl apply -f bootstrap.yaml
```

After it succeeds, Nexus reconstructs the root key from the keepers and becomes
ready.

## Workflow

1. **Deploy keepers first.** Bring up `SPIKE_NEXUS_SHAMIR_SHARES` keepers
   (default 3) and make sure each is reachable at its
   `SPIKE_NEXUS_KEEPER_PEERS` URL.
2. **Run bootstrap** with the `spike/bootstrap` SPIFFE identity and the shared
   config (`SPIKE_NEXUS_KEEPER_PEERS`, `SPIKE_NEXUS_SHAMIR_SHARES`,
   `SPIKE_NEXUS_SHAMIR_THRESHOLD`, the trust roots). It:
   - generates the root key and splits it (Shamir over P-256, via CIRCL);
   - `POST`s one share to each keeper's `/v1/store/contribute` (with retries);
   - verifies initialization by asking Nexus to decrypt a probe encrypted with
     the canonical root key;
   - records completion in the `spike-bootstrap-state` ConfigMap (Kubernetes).
3. **Nexus recovers.** On the next loop Nexus collects `threshold` shares,
   reconstructs the root key, initializes its backend, and becomes ready. It
   then re-syncs shares to keepers periodically.

## Tips

- **Shares vs threshold:** `SHARES` keepers each hold one share; any
  `THRESHOLD` of them can reconstruct the key (e.g. 3 shares / threshold 2
  tolerates one keeper down). Production guidance is 5 shares / threshold 3.
- **Idempotency:** the `spike-bootstrap-state` ConfigMap makes re-runs no-ops.
  Grant the bootstrap ServiceAccount RBAC to read/write that ConfigMap, or a
  retried Job could re-bootstrap with a *new* root key and orphan the data
  encrypted under the old one. Set `SPIKE_BOOTSTRAP_FORCE=true` only when you
  deliberately want to re-key.
- Bootstrap waits for the SPIRE agent socket (init container in Kubernetes)
  before it runs.

## Pitfalls

- **No `--wait` race.** Don't `helm install --wait` the chart and expect Nexus
  Ready *before* bootstrap runs; Nexus can't be ready until the keepers are
  seeded. Install without `--wait`, run bootstrap, then wait on the Nexus
  rollout.
- **Keepers not all up.** Bootstrap requires exactly `SHARES` reachable keepers;
  if some aren't resolvable yet it retries, but a wrong peer list fails it.
- **Verify needs to reach Nexus.** The post-seed verification calls Nexus's
  API; set `SPIKE_NEXUS_API_URL` to a resolvable address (in Kubernetes, the
  fully-qualified service DNS).

## Cross-Links

- [Choosing a backend store](/recipes/choosing-a-backend-store/)
- [Where the root key lives: keepers, Shamir, and recovery](/recipes/root-key-keepers-recovery/)
- [Troubleshooting](/recipes/troubleshooting/) (Nexus stuck in keeper recovery)
- Reference: [Configuration](/usage/configuration/)

## What's Next

Understand the moving parts you just wired up:
[Where the root key lives](/recipes/root-key-keepers-recovery/).
