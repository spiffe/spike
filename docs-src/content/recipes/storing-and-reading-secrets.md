+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Storing and Reading Secrets"
weight = 4
sort_by = "weight"
+++

# Storing and Reading Secrets

## Problem

You have SPIKE up and running and you want to do the everyday thing: write a
secret, read it back, list what is there, and clean up old ones. You also want
to know what happens when you overwrite a secret (does the old value vanish?)
and how to recover one you deleted by mistake.

## TL;DR

Secrets are **versioned key-value maps** stored at a namespaced path. Use
`spike secret` from SPIKE Pilot:

```bash
spike secret put   tenants/acme/db/creds user=acme pass=SPIKERocks
spike secret get   tenants/acme/db/creds
spike secret list
spike secret delete tenants/acme/db/creds      # soft delete (recoverable)
spike secret undelete tenants/acme/db/creds    # bring it back
```

Every `put` to an existing path creates a **new version**; old versions are
retained until you explicitly delete them.

## Workflow

1. **Write a secret.** A secret is one or more `key=value` pairs at a path:

   ```bash
   spike secret put tenants/acme/db/creds user=acme pass=SPIKERocks
   ```

   This is an upsert. Writing the same path again stores a new version rather
   than mutating the old one.

2. **Read it back.** Get the whole map, a single key, or a specific version:

   ```bash
   spike secret get tenants/acme/db/creds          # all keys, current version
   spike secret get tenants/acme/db/creds pass      # just one key
   spike secret get tenants/acme/db/creds -v 2      # version 2 (0 = current)
   spike secret get tenants/acme/db/creds -f json   # plain | yaml | json
   ```

3. **List paths.** `list` shows every secret path you are allowed to see:

   ```bash
   spike secret list
   spike secret list -f json
   ```

4. **Inspect metadata** without revealing the value (versions, timestamps,
   current version):

   ```bash
   spike secret metadata get tenants/acme/db/creds
   spike secret metadata get tenants/acme/db/creds -v 2
   ```

5. **Delete and undelete.** Delete is a *soft* delete: the version is marked
   deleted, not destroyed, so it can be restored. Versions are given as a
   comma-separated list with `-v`; `0` means the current version (the default
   when `-v` is omitted):

   ```bash
   spike secret delete   tenants/acme/db/creds          # current version
   spike secret delete   tenants/acme/db/creds -v 1,2,3 # specific versions
   spike secret undelete tenants/acme/db/creds -v 1,2,3 # restore them
   ```

## Tips

- **Multiple keys per path.** One path can hold a whole map
  (`put .../creds user=acme pass=… host=db.internal`). Group related fields
  under one path instead of scattering them.
- **`-v` differs by command.** For `get` and `metadata get`, `-v/--version`
  takes a single integer. For `delete` and `undelete`, `-v/--versions` takes a
  comma-separated list. `0` always means the current version.
- **Output formats.** `get`, `list`, and `metadata get` accept
  `-f/--format` with `plain`/`p`, `yaml`/`y`, or `json`/`j`. Use `json` when
  piping into scripts.
- **Persistence depends on the backend.** In `memory` mode everything is gone
  on restart; `sqlite` persists to disk; `lite` keeps no secrets at all (it is
  encryption-only). See
  [Choosing a backend store](/recipes/choosing-a-backend-store/).

## Pitfalls

- **Paths are namespaces, not filesystem paths.** Use `tenants/acme/db/creds`,
  never `/tenants/acme/db/creds`. A leading slash is wrong, and trailing
  slashes are discouraged.
- **`list` takes no path argument.** It lists all paths you can access; it does
  not filter by prefix. Filter in your shell if you need a subset.
- **Reading needs a policy.** SPIKE Pilot authenticates with its SPIFFE ID, but
  it still needs a policy granting `read`/`write` on the path. "I can `put` but
  another workload can't `get`" is almost always a missing policy, not a
  storage problem. See
  [Writing access policies](/recipes/writing-access-policies/).
- **Overwrite does not destroy history.** `put` over an existing path keeps the
  old version. If you must scrub a value, delete the specific versions.

## Cross-Links

- [Choosing a backend store](/recipes/choosing-a-backend-store/)
- [Writing access policies](/recipes/writing-access-policies/)
- [Granting a workload access to secrets](/recipes/granting-a-workload-access/)
- Reference: [SPIKE CLI](/usage/cli/) and the
  [command reference](/usage/commands/)

## What's Next

Control who can read and write these secrets:
[Writing access policies](/recipes/writing-access-policies/).
