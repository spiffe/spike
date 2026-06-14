+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Where the root key lives"
weight = 3
sort_by = "weight"
+++

# Where the root key lives: keepers, Shamir, and recovery

## Problem

SPIKE Nexus encrypts everything with a single **root key**. Where is that key,
how does it survive a Nexus restart without a human typing it in, and why does
Nexus talk to "keepers" on startup? Understanding this explains why `lite`/
`sqlite` need keepers and why a missing keeper isn't fatal.

## TL;DR

The root key is never written to disk. It's split with **Shamir's Secret
Sharing** into N shares; each **SPIKE Keeper** holds one share in memory. On
startup Nexus collects any `threshold` shares from the keepers and
reconstructs the key in memory. Lose one keeper and the key still recovers;
lose more than `N − threshold` and you fall back to
[break-the-glass recovery](/recipes/break-the-glass-recovery/).

## Workflow (what happens automatically)

1. **Bootstrap** generates the root key, splits it into `SPIKE_NEXUS_SHAMIR_SHARES`
   shares, and seeds the keepers (one share each). See
   [Bootstrapping](/recipes/bootstrapping-spike/).
2. **Nexus startup** (`lite`/`sqlite`): `InitializeBackingStoreFromKeepers`
   iterates the keepers, gathers shares until it has
   `SPIKE_NEXUS_SHAMIR_THRESHOLD` of them, and reconstructs the canonical root
   key (the seed reduced into a P-256 scalar). It then keys its cipher /
   opens its store and serves the API.
3. **Ongoing sync:** Nexus runs `SendShardsPeriodically`, re-pushing shares to
   the keepers on an interval so restarted/replaced keepers get re-hydrated.
4. **Keeper restart:** a keeper holds its share only in memory, so a restarted
   keeper is empty until Nexus re-syncs it — which is why a single keeper
   bouncing is harmless as long as `threshold` others are up.

## Tips

- **`memory` mode has no root key and no keepers** — it's the only standalone
  mode. `lite` and `sqlite` always recover from keepers.
- Pick `shares`/`threshold` for your failure tolerance: you can lose up to
  `shares − threshold` keepers and still recover automatically.
- The reconstructed key lives only in Nexus memory; harden the host
  accordingly (see [Production hardening](/recipes/production-hardening/)).

## Pitfalls

- **Canonical key ≠ raw seed.** The AES key is the Shamir secret *scalar*
  marshalled to bytes, not the raw random seed (the seed is reduced mod the
  P-256 group order). Anything that needs the actual key (e.g. the bootstrap
  verify probe) must derive it the same way, not use the seed.
- **All keepers empty = stuck Nexus, not a crash.** If no keeper has a share
  (fresh deploy without bootstrap, or all keepers restarted before re-sync),
  Nexus retries recovery indefinitely by design. See
  [Troubleshooting](/recipes/troubleshooting/).
- **Keepers are not a secret store.** They hold only root-key shares, never
  your secrets.

## Cross-links

- [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/)
- [Break-the-glass disaster recovery](/recipes/break-the-glass-recovery/)
- [Backup and restore](/recipes/backup-and-restore/)
- Architecture: [System overview](/architecture/system-overview/) ·
  [Security model](/architecture/security-model/)

## What's next

Plan for the day the keepers can't recover the key:
[Break-the-glass disaster recovery](/recipes/break-the-glass-recovery/).
