+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Choosing a Backend Store: Memory, Lite, or SQLite"
weight = 1
sort_by = "weight"
+++

# Choosing a Backend Store: Memory, Lite, or SQLite

## Problem

SPIKE Nexus can run with one of three backend stores, set by
`SPIKE_NEXUS_BACKEND_STORE`: `memory`, `lite`, or `sqlite`. They differ on two
**independent** axes: whether secrets are *persisted*, and whether a *root
key* (and therefore SPIKE Keepers) is required. Picking the wrong one leads to
confusing failures: a Nexus that waits forever for keepers that were never
deployed, or an in-memory store that silently loses everything on restart.

## TL;DR

| Backend  | Persists secrets? | Root key? | Keepers required? | Use it for |
|----------|-------------------|-----------|-------------------|------------|
| `memory` | No (in-process)   | **No** (must be nil) | **No** | local dev / tests |
| `lite`   | No (no store at all) | **Yes** | **Yes** | encryption-as-a-service; secrets live elsewhere (e.g. S3) |
| `sqlite` | Yes (`~/.spike/data/spike.db`, AES-256-GCM at rest) | **Yes** | **Yes** | production (default) |

Rule of thumb: **`sqlite`** for production, **`memory`** for throwaway dev, and
**`lite`** only when SPIKE is your encrypt/decrypt service and something else
stores the ciphertext.

## Workflow

Set the mode on SPIKE Nexus:

```bash
export SPIKE_NEXUS_BACKEND_STORE=sqlite   # or: lite | memory
```

- **memory**: Nexus initializes a volatile in-process store and **does not
  contact keepers**. No bootstrap, no keepers, no root key. Nexus logs a
  "not for production" warning at startup. Restarting Nexus wipes all secrets.
- **lite** and **sqlite**: Nexus recovers its root key from the SPIKE Keepers
  on startup (Shamir reconstruction), so you must deploy keepers **and** seed
  them once via SPIKE Bootstrap. See
  [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/).
  - **sqlite** then opens/creates the encrypted database and serves the full
    secret + policy API.
  - **lite** keeps **no** local store (it embeds a no-op backend) and serves
    only the cipher API; it is an *encryption-only* service.

## Tips

- The default is `sqlite`; you only need to set the variable to choose
  `lite` or `memory`.
- `memory` is the only mode that runs standalone: no keepers, no bootstrap.
  Reach for it in unit/integration tests and quick local experiments.
- Use `lite` when secrets are stored externally (e.g. S3-compatible storage)
  and you just need SPIKE to encrypt/decrypt with a SPIFFE-gated key. See
  [Using SPIKE as an encryption service](/recipes/encryption-as-a-service/).
- For `sqlite`, point `SPIKE_NEXUS_DATA_DIR` at durable, access-controlled
  storage and back it up. See [Backup and restore](/recipes/backup-and-restore/).

## Pitfalls

- **"lite doesn't need keepers."** It does. `lite` and `sqlite` both recover the
  root key from keepers on startup; only `memory` is keeper-free. If you deploy
  `lite` without seeded keepers, Nexus loops forever in keeper recovery and
  never becomes ready (see [Troubleshooting](/recipes/troubleshooting/)).
- **"lite is an in-memory store."** It isn't; `lite` has *no* store (it's
  encryption-only). The in-memory secret store is `memory`.
- **`memory` and a root key are mutually exclusive.** In `memory` mode the root
  key must be nil; passing one is treated as an initialization bug. Conversely
  `lite`/`sqlite` refuse to start with a nil/empty root key.
- **`memory` loses data on restart.** Never use it where you expect secrets to
  survive a process restart.

## Cross-Links

- [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/)
- [Where the root key lives: keepers, Shamir, and recovery](/recipes/root-key-keepers-recovery/)
- [Using SPIKE as an encryption service](/recipes/encryption-as-a-service/)
- Reference: [Configuration](/usage/configuration/) ·
  Architecture: [System overview](/architecture/system-overview/)

## What's Next

If you chose `lite` or `sqlite`, set up keepers and seed them:
[Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/).
