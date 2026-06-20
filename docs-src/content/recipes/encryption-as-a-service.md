+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Using SPIKE as an Encryption Service"
weight = 7
sort_by = "weight"
+++

# Using SPIKE as an Encryption Service

## Problem

Sometimes you do not want SPIKE to *store* your data; you want it to **encrypt**
data you store somewhere else (an object store, a database column, a file on
disk). You keep custody of the ciphertext; SPIKE holds the key and does the
crypto. This is "encryption as a service," and it pairs naturally with `lite`
mode, where Nexus has a root key and keepers but no secret store at all.

## TL;DR

The `spike cipher` command encrypts and decrypts through Nexus without
persisting anything:

```bash
# encrypt a file (stream mode), keep the ciphertext yourself
spike cipher encrypt -f plan.txt -o plan.enc

# decrypt it back
spike cipher decrypt -f plan.enc -o plan.txt
```

The plaintext is never stored in SPIKE; only the key (derived from the root
key) lives there. Run Nexus in `lite` mode when this is the *only* thing you
need from it.

## Workflow

1. **Encrypt.** Stream mode reads a file or stdin and writes ciphertext to a
   file or stdout; it handles binary data transparently:

   ```bash
   spike cipher encrypt -f secret-plan.txt -o secret-plan.enc
   echo "transient token" | spike cipher encrypt -o token.enc
   ```

2. **Store the ciphertext wherever you like**: S3/minio, a database BLOB, a
   git-crypt-style file. SPIKE is out of the loop until you need it back.

3. **Decrypt** by feeding the ciphertext back through Nexus:

   ```bash
   spike cipher decrypt -f secret-plan.enc -o secret-plan.txt
   cat token.enc | spike cipher decrypt
   ```

4. **For programmatic callers, use JSON mode.** Encrypt accepts base64
   `--plaintext` and returns the version byte, nonce, and ciphertext; decrypt
   takes those three back:

   ```bash
   spike cipher encrypt --plaintext "$(printf 'hello' | base64)"
   # -> JSON with {version, nonce, ciphertext} (all base64)

   spike cipher decrypt \
     --version 1 \
     --nonce "<base64-nonce>" \
     --ciphertext "<base64-ciphertext>"
   ```

## Tips

- **Stream mode for files, JSON mode for code.** Stream mode (`-f`/`-o` or
  stdin/stdout) is the easy path for files and pipelines. JSON mode (passing
  `--plaintext`, or any of `--version`/`--nonce`/`--ciphertext`) is for callers
  that want to persist the components separately.
- **Keep the version byte.** Decryption needs the version, nonce, and
  ciphertext that encryption produced. Store all three with your data; losing
  the nonce or version makes the ciphertext undecryptable.
- **Pair with `lite` mode.** If encryption is all you need, `lite` gives you the
  cipher routes (and the root key/keepers that back them) without a secret
  store to operate, back up, or persist.
- **Access still needs a policy.** The caller authenticates with its SPIFFE ID
  and needs permission to use the cipher routes, the same as any other SPIKE
  operation.

## Pitfalls

- **`lite` still needs keepers and a root key.** "No secret store" does not mean
  "no setup." `lite` is encryption-only, but the key that encrypts your data is
  the root key, which is reconstructed from the keepers. You must bootstrap it
  exactly like `sqlite`. See
  [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/).
- **You own the ciphertext durability.** SPIKE does not keep a copy. If you lose
  the ciphertext, SPIKE cannot recover the plaintext; it only holds the key.
- **Re-keying invalidates old ciphertext.** The key is derived from the root
  key. If you re-bootstrap with a new root key, data encrypted under the old key
  can no longer be decrypted. Treat root-key rotation as a deliberate migration.
- **`--plaintext` is base64.** In JSON mode the plaintext is base64-encoded, not
  raw text. Encode on the way in and decode on the way out.

## Cross-Links

- [Choosing a backend store](/recipes/choosing-a-backend-store/) (when `lite`
  is the right mode)
- [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/)
- [Writing access policies](/recipes/writing-access-policies/)
- Reference: [Configuration](/usage/configuration/) and the
  [command reference](/usage/commands/)

## What's Next

Make sure you can recover the key that all this depends on:
[Break-the-glass disaster recovery](/recipes/break-the-glass-recovery/).
