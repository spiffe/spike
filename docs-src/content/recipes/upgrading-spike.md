+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Upgrading SPIKE"
weight = 14
sort_by = "weight"
+++

# Upgrading SPIKE

## Problem

SPIKE is actively hardened and patched, so you want to stay current, but two
SPIKE-specific facts make upgrades more than a binary swap. First, there is no
built-in database migration between versions, so you cannot assume a new Nexus
reads an old store. Second, component identity is often bound to the binary's
SHA-256 in SPIRE, so a new binary fails attestation until you update the
registration entry. Upgrade without accounting for both and the system either
will not start or will not trust itself.

## TL;DR

```text
1. Read the changelog for breaking changes (config, key lengths, schema)
2. Back up first: database + root-key shards
3. Verify the new binaries' SHA-256 checksums
4. Update SPIRE entries if they pin the binary SHA selector
5. Upgrade in a staging environment, verify, then production
```

Back up before, verify checksums, update the SHA selectors, and rehearse in
staging. Never upgrade production as the first place you try the new version.

## Workflow

1. **Read the changelog and release notes** for the target version. Look
   specifically for changes to configuration, environment variables, default
   key lengths or cipher suites, and the storage schema. See
   [Changelog](/tracking/changelog/).

2. **Back up everything first.** Capture the database and export fresh
   root-key shards before you touch any binary:

   ```bash
   sqlite3 ~/.spike/data/spike.db "PRAGMA wal_checkpoint(FULL);"
   sqlite3 ~/.spike/data/spike.db \
     ".backup '/backup/pre-upgrade-$(date +%Y%m%d).sqlite'"
   spike operator recover   # export + secure the shards
   ```

   See [Backup and restore](/recipes/backup-and-restore/).

3. **Verify the new binaries.** Official SPIKE binaries publish SHA-256
   checksums. Verify each before installing:

   ```bash
   sha256sum -c spike-<version>.sha256
   ```

4. **Update SPIRE registration entries if they pin the SHA.** If your entries
   bind a `unix:sha256:<hash>` selector (recommended hardening), the new binary
   has a new hash and will fail attestation until you re-register it:

   ```bash
   spire-server entry create \
     -spiffeID spiffe://spike.ist/spike/keeper \
     -parentID "spiffe://spike.ist/spire-agent" \
     -selector unix:uid:"$KEEPER_UID" \
     -selector unix:path:"$KEEPER_PATH" \
     -selector unix:sha256:"$NEW_KEEPER_SHA"
   ```

5. **Upgrade in staging, verify, then production.** Roll the new version into a
   staging environment that mirrors production. Confirm Nexus reaches Ready and
   secrets round-trip:

   ```bash
   sqlite3 ~/.spike/data/spike.db "PRAGMA integrity_check;"
   spike secret get path/to/test/secret
   ```

   Only then promote to production.

## Tips

- **Keep a frequent cadence.** Small, regular upgrades are far less risky than
  rare jumps across many versions, where breaking changes accumulate.
- **Upgrade order.** Bring keepers and Nexus to the new version together, and
  expect Nexus to recover its root key from the keepers on restart, exactly as
  on first boot. If it stalls, see [Troubleshooting](/recipes/troubleshooting/).
- **The root key is version-independent.** Upgrading binaries does not re-key.
  Your existing shards and encrypted data remain valid as long as you do not
  re-bootstrap.
- **Automate checksum verification** in your deployment pipeline so an
  unverified binary can never reach a node.

## Pitfalls

- **No cross-version DB migration.** SPIKE ships no tool to migrate the store
  between incompatible versions. If a release changes the schema, plan a
  deliberate migration; do not point new Nexus at an old database and hope.
- **Stale SHA selectors lock you out.** Forgetting step 4 makes the upgraded
  component fail SPIRE attestation, so it gets no SVID and the rest of the
  system refuses to talk to it. This looks like a connectivity failure but is an
  identity failure.
- **No backup, no rollback.** If the upgrade goes wrong and you skipped the
  backup, there is nothing to restore to. The backup in step 2 is the rollback
  plan.
- **Defaults can change silently.** A new version may tighten key lengths or
  cipher suites. Re-read the configuration reference after upgrading rather than
  assuming your old settings are still optimal. See
  [Configuration](/usage/configuration/).

## Cross-Links

- [Backup and restore](/recipes/backup-and-restore/)
- [Production hardening](/recipes/production-hardening/)
- [Troubleshooting](/recipes/troubleshooting/)
- Reference: [Changelog](/tracking/changelog/),
  [Configuration](/usage/configuration/)

## What's Next

You have completed the recipe set. Revisit the
[concepts](/recipes/choosing-a-backend-store/) any time the fundamentals feel
fuzzy, or browse all [Recipes](/recipes/).
