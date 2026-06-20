+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Break-the-Glass Disaster Recovery"
weight = 8
sort_by = "weight"
+++

# Break-the-Glass Disaster Recovery

## Problem

Normally SPIKE recovers itself: Nexus rebuilds its root key from the keepers on
startup. But if you lose enough keepers at once (the whole cluster, the node,
the data center) there is nothing left to auto-recover *from*. Break-the-glass
recovery is the human-held fallback: a set of recovery shards an operator
exports **ahead of time** and feeds back in **after** a catastrophe.

The catch is in the timing. The shards must be exported while the system is
healthy. If you wait until the outage, it is too late.

## TL;DR

Two operator commands, two different moments:

```bash
# BEFORE disaster, while Nexus is healthy: export recovery shards
spike operator recover

# AFTER disaster, when Nexus cannot auto-recover: feed shards back in
spike operator restore   # prompts for one shard at a time; repeat
```

`recover` needs the `recover` role; `restore` needs the `restore` role. Store
the exported shards encrypted, offline, and split across custodians.

## Workflow

### Phase 1: Export Shards (Do This Now, While Healthy)

1. As an operator with the `recover` role, run:

   ```bash
   spike operator recover
   ```

2. SPIKE writes the recovery shards to the recovery directory as
   `spike.recovery.0.txt`, `spike.recovery.1.txt`, ... Each file holds one
   shard in `spike:<index>:<hex>` format.

3. **Immediately secure them.** Encrypt each shard, move it to safe offline
   storage (ideally different custodians/locations), and securely erase the
   plaintext files from the recovery directory. SPIKE prints this reminder for
   a reason: if you lose these shards, a total crash is unrecoverable.

### Phase 2: Restore (Only After a Catastrophe)

4. When Nexus cannot auto-recover (keepers gone, no root key), an operator with
   the `restore` role runs:

   ```bash
   spike operator restore
   ```

5. Paste one recovery shard when prompted. Input is hidden. SPIKE reports
   progress:

   ```text
   Shards collected:  1
   Shards remaining:  1
   Please run `spike operator restore` again to provide the remaining shards.
   ```

6. Repeat with the next shard until SPIKE collects the threshold and prints
   `SPIKE is now restored and ready to use.`

## Tips

- **`recover` vs `restore`.** `recover` *exports* shards from a healthy system
  (proactive backup). `restore` *imports* them into a broken one (reactive
  rebuild). They are not opposites of one command; they are two halves of one
  drill.
- **Threshold, not all.** Restore needs `threshold` shards, not every shard, so
  you can tolerate losing some custodians. This is the same Shamir threshold
  that backs keeper auto-recovery. See
  [Where the root key lives](/recipes/root-key-keepers-recovery/).
- **Rehearse it.** Schedule a recovery drill: export shards, stand up a
  throwaway Nexus, and restore into it. A break-the-glass procedure no one has
  run is a guess, not a plan.
- **Roles are separate identities.** The `recover` and `restore` roles are
  distinct SPIFFE-ID roles, separate from day-to-day Pilot access. Provision
  them deliberately to the humans who hold the glass.

## Pitfalls

- **You cannot export after the disaster.** `recover` talks to a *healthy*
  Nexus. If you skipped Phase 1, there is no second chance once the keepers are
  gone. Export shards as part of going to production, not as an afterthought.
- **Shards are root-key material.** Anyone with `threshold` shards can rebuild
  the root key and decrypt everything. Treat them like the keys to the kingdom:
  encrypted, offline, split, audited.
- **Shard format is exact.** A shard is `spike:<index>:<hex>` where the hex is
  64 characters (32 bytes). Truncated or reformatted shards are rejected. Keep
  them byte-for-byte.
- **Re-keying invalidates old shards.** If you re-bootstrap with a new root key,
  previously exported shards no longer restore the current system. Re-export
  after any deliberate root-key rotation.

## Cross-Links

- [Where the root key lives: keepers, Shamir, and recovery](/recipes/root-key-keepers-recovery/)
- [Backup and restore](/recipes/backup-and-restore/)
- [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/)
- Reference: [Recovery operations](/operations/recovery/)

## What's Next

Pair key recovery with data backup:
[Backup and restore](/recipes/backup-and-restore/).
