+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Writing Access Policies"
weight = 5
sort_by = "weight"
+++

# Writing Access Policies

## Problem

A workload can authenticate to SPIKE with its SPIFFE ID, but authentication is
not authorization. Until a **policy** says "this identity may do these things
on these paths," every read and write is denied. You need to write that policy,
and the two fields that decide who and what (`spiffeid-pattern` and
`path-pattern`) are **regular expressions**, which is the single most common
place people trip.

## TL;DR

A policy binds a **SPIFFE ID regex** and a **path regex** to a set of
**permissions** (`read`, `write`, `list`, `super`):

```bash
spike policy create \
  --name "acme-db" \
  --spiffeid-pattern '^spiffe://example\.org/acme/db/.*$' \
  --path-pattern     '^tenants/acme/db/.*$' \
  --permissions      'read,write'
```

`create` fails if the name already exists. Use `apply` for upsert (and for
applying a YAML file).

## Workflow

1. **Decide the four inputs:**
   - **name**: a unique label for the policy.
   - **spiffeid-pattern**: a regex matching the workload SVIDs it applies to.
   - **path-pattern**: a regex matching the secret paths it covers.
   - **permissions**: any of `read`, `write`, `list`, `super`.

2. **Create the policy:**

   ```bash
   spike policy create \
     --name "acme-db" \
     --spiffeid-pattern '^spiffe://example\.org/acme/db/.*$' \
     --path-pattern     '^tenants/acme/db/.*$' \
     --permissions      'read,write'
   ```

3. **Or apply from YAML** (upsert: creates if new, updates if the name exists):

   ```yaml
   # acme-db.yaml
   name: acme-db
   spiffeidPattern: ^spiffe://example\.org/acme/db/.*$
   pathPattern: ^tenants/acme/db/.*$
   permissions:
     - read
     - write
   ```

   ```bash
   spike policy apply --file acme-db.yaml
   ```

4. **Verify and inspect:**

   ```bash
   spike policy list
   spike policy list --path-pattern '^tenants/acme/.*$'
   spike policy get --name acme-db
   spike policy get --name acme-db --format json
   ```

5. **Delete when no longer needed** (by name or ID; prompts to confirm):

   ```bash
   spike policy delete --name acme-db
   ```

## Tips

- **`create` vs `apply`.** `create` is strict: it errors if a policy with that
  name already exists, which is what you want in scripts that must not clobber.
  `apply` is upsert and is the only form that reads a `--file`. Use `apply` for
  declarative, re-runnable config.
- **Permissions.** `read` reads secrets, `write` creates/updates/deletes,
  `list` lists resources, and `super` is administrative (grants all). Grant the
  narrowest set that works; reserve `super` for operators.
- **Patterns can be broad or pinned.** `^tenants/acme/db/.*$` covers a subtree;
  `^tenants/acme/db/creds$` pins one exact path. Anchor with `^` and `$` so a
  pattern does not match more than you intend.
- **`policy list` filters.** Filter by `--path-pattern` or `--spiffeid-pattern`
  (not both at once) to find which policies touch a path or an identity.

## Pitfalls

- **Patterns are regex, not globs.** This is the big one.
  - Correct: `^tenants/acme/db/.*$`, `^spiffe://example\.org/web/.*$`
  - Wrong: `tenants/acme/db/*`, `spiffe://example.org/web/*`

  In a glob, `*` means "anything"; in regex it means "zero or more of the
  previous character." A glob-style pattern silently matches the wrong set.
- **Escape the dots in SPIFFE IDs.** `.` matches any character in regex. Write
  `example\.org`, not `example.org`, or the pattern will also match
  `exampleXorg`.
- **Paths are namespaces, not filesystem paths.** Match `^tenants/acme/.*$`,
  never `^/tenants/acme/.*$`. A leading slash is wrong here just as it is when
  storing secrets.
- **Both patterns must match.** Access is granted only when the caller's
  SPIFFE ID matches `spiffeid-pattern` **and** the target path matches
  `path-pattern`. A policy that "looks right" but denies access usually has one
  of the two patterns too narrow or unanchored.

## Cross-Links

- [Storing and reading secrets](/recipes/storing-and-reading-secrets/)
- [Granting a workload access to secrets](/recipes/granting-a-workload-access/)
- Reference: [Configuration](/usage/configuration/) and the
  [command reference](/usage/commands/)

## What's Next

Put the policy to work end to end:
[Granting a workload access to secrets](/recipes/granting-a-workload-access/).
