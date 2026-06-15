+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Backup and Restore"
weight = 9
sort_by = "weight"
+++

# Backup and Restore

## Problem

A full SPIKE backup is **two independent things**, and people routinely save
one and forget the other. The SQLite database holds your encrypted secrets; the
root-key shards hold the only thing that can decrypt them. A database backup
without the key is undecryptable noise. The key without the database has nothing
to unlock. You need both, backed up on their own schedules, and a restore
procedure that puts them back in the right order.

> `memory` and `lite` modes have no database to back up. For them, only the
> root-key shards matter.

## TL;DR

```bash
# 1. Back up the encrypted secret store (sqlite mode)
sqlite3 ~/.spike/data/spike.db "PRAGMA wal_checkpoint(FULL);"
sqlite3 ~/.spike/data/spike.db \
  ".backup '/backup/spike-$(date +%Y%m%d).sqlite'"

# 2. Back up the root key as recovery shards (all modes)
spike operator recover   # then encrypt + store the shards offline
```

Restore is the reverse: put the database back, then reconstruct the root key
from shards if Nexus cannot auto-recover.

## Workflow

### Backup

1. **Database (sqlite mode).** Use SQLite's online `.backup`, not a file copy,
   so you get a consistent snapshot. Checkpoint the WAL first:

   ```bash
   sqlite3 ~/.spike/data/spike.db "PRAGMA wal_checkpoint(FULL);"
   sqlite3 ~/.spike/data/spike.db \
     ".backup '/backup/spike-$(date +%Y%m%d_%H%M%S).sqlite'"
   sqlite3 /backup/spike-*.sqlite "PRAGMA integrity_check;"
   ```

2. **Root-key shards (all modes).** Export and secure them as covered in
   [break-the-glass recovery](/recipes/break-the-glass-recovery/):

   ```bash
   spike operator recover   # writes spike.recovery.N.txt; then encrypt + erase
   ```

3. **Supporting state.** Also capture what you need to rebuild identity:

   ```bash
   spire-server entry show > /backup/spire-entries-$(date +%Y%m%d).txt
   # plus SPIRE server/agent config and SPIKE configuration
   ```

### Restore

4. **Database.** Stop Nexus, swap the file in, lock it down, restart:

   ```bash
   cp /backup/spike-TIMESTAMP.sqlite ~/.spike/data/spike.db
   chmod 600 ~/.spike/data/spike.db
   ```

5. **Root key.** If Nexus cannot auto-recover from the keepers, reconstruct it
   from shards (needs the `restore` role):

   ```bash
   spike operator restore   # paste shards until the threshold is met
   ```

6. **Verify.** Confirm the store is intact and crypto round-trips:

   ```bash
   sqlite3 ~/.spike/data/spike.db "PRAGMA integrity_check;"
   spike secret get path/to/test/secret
   ```

## Tips

- **Two assets, two cadences.** Back up the database daily (it changes with
  every secret write). Re-export root-key shards only after initial setup and
  after any deliberate root-key rotation; the key does not change otherwise.
- **The database is encrypted at rest.** Its contents are useless without the
  root key, so the database backup is far less sensitive than the shards. Guard
  the shards like the crown jewels; the database like ordinary backups.
- **Test restores, not just backups.** A backup you have never restored is a
  hypothesis. Rehearse the full restore into a throwaway environment on a
  schedule.
- **DB location.** The SQLite store lives at `~/.spike/data/spike.db` on the
  Nexus host.

## Pitfalls

- **Database without key is unrecoverable.** The most common mistake is backing
  up `spike.db` and never running `spike operator recover`. Encrypted secrets
  with no key are gone. Always pair the two.
- **File-copy backups corrupt.** Copying `spike.db` while Nexus is running can
  capture a torn write. Use SQLite's `.backup` (and checkpoint the WAL) for a
  consistent snapshot.
- **Restore order matters.** Put the database in place first, then restore the
  root key. Restoring the key into an empty store leaves nothing to decrypt.
- **No cross-version DB migration.** SPIKE has no built-in migration between
  versions. Restore into the same (or a compatible) SPIKE version you backed up
  from; plan version upgrades separately. See
  [Upgrading SPIKE](/recipes/upgrading-spike/).

## Cross-Links

- [Break-the-glass disaster recovery](/recipes/break-the-glass-recovery/)
- [Choosing a backend store](/recipes/choosing-a-backend-store/)
- [Where the root key lives](/recipes/root-key-keepers-recovery/)
- Reference: [Backup and Restore guide](/operations/backup/)

## What's Next

Stand the whole thing up cleanly in production:
[Deploying SPIKE](/recipes/deploying-spike/).
