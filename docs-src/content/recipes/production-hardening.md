+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Production Hardening"
weight = 11
sort_by = "weight"
+++

# Production Hardening

## Problem

SPIKE keeps its most sensitive material (the root key, decrypted secrets) in
memory by design. That makes the *host* the trust boundary: if an attacker can
read process memory, swap, or core dumps, encryption at rest no longer helps.
Going to production means closing those host-level gaps and hardening the SPIRE
identity plane SPIKE depends on. This recipe is the prioritized checklist; the
[Production Setup guide](/operations/production/) has the full rationale.

## TL;DR

Protect memory, run unprivileged, lock down the identity plane, and restrict
the backing store:

```bash
# Keep secrets out of disk-backed memory
swapoff -a                                   # disable swap
ulimit -c 0                                  # disable core dumps (RLIMIT_CORE)
# allow mlock for the spike user (limits.conf / LimitMEMLOCK=infinity)
# restrict debugging
echo 'kernel.yama.ptrace_scope = 2' | sudo tee /etc/sysctl.d/10-ptrace.conf
```

Then: non-root service accounts, mTLS-only SPIRE, granular registration
entries (no wildcards), and least-privilege access to `~/.spike/data`.

## Workflow

1. **Protect memory (highest value).** SPIKE attempts `mlockall` to keep keys
   out of swap; give the `spike` user the memlock limit so it succeeds.
   - Disable swap (`swapoff -a`), or use encrypted swap if you cannot.
   - Disable core dumps (`RLIMIT_CORE=0`, or `LimitCORE=0` in systemd).
   - Restrict `ptrace`: set `kernel.yama.ptrace_scope` to `2` (root-only) or `3`
     (off). This applies to both Nexus and the Keepers.
   - Keep ASLR on (`kernel.randomize_va_space = 2`).

2. **Run unprivileged and isolated.**
   - Run every component as a dedicated **non-root** user with minimal
     permissions. Never run as root.
   - Single-tenant the Nexus host (node affinity in Kubernetes); it should be
     the only significant process on the box.
   - In Kubernetes: `allowPrivilegeEscalation: false`, `privileged: false`,
     `readOnlyRootFilesystem: true`, non-root `runAsUser`/`runAsGroup`, Pod
     Security Admission, and NetworkPolicies.

3. **Harden the SPIRE identity plane.** SPIKE's security rests on it.
   - Isolate the SPIRE server (separate cluster or dedicated hardware); consider
     a KMS plugin for its keys.
   - Write **granular** registration entries; avoid wildcard selectors.
   - Bind component identities to binary `sha256` selectors so a swapped binary
     fails attestation and gets no SVID.
   - mTLS everywhere; rotate workload certs frequently.

4. **Restrict the backing store and backups.**
   - Limit write access to `~/.spike/data` to the Nexus process only.
   - For external stores (S3/minio/DB), assume untrusted: TLS in transit,
     restricted access, encrypted at rest.
   - Encrypt backups; guard root-key shards far more strictly than the (already
     encrypted) database. See
     [Backup and restore](/recipes/backup-and-restore/).

5. **Set the Shamir parameters for your scale.**

   ```bash
   export SPIKE_NEXUS_SHAMIR_SHARES=5      # total keepers
   export SPIKE_NEXUS_SHAMIR_THRESHOLD=3   # needed to reconstruct
   ```

   | Deployment | Threshold | Shares |
   |------------|-----------|--------|
   | Dev/Test   | 2         | 3      |
   | Small Prod | 3         | 5      |
   | Large Prod | 5         | 7      |
   | Critical   | 7         | 10     |

## Tips

- **Defense in depth.** No single control is sufficient. Memory locking plus no
  swap plus no core dumps plus ptrace restriction together make memory
  extraction genuinely hard.
- **Verify binary integrity.** SPIKE binaries ship with SHA-256 checksums.
  Verify on install and re-check periodically; binding the SHA into SPIRE
  entries makes this load-bearing, not just advisory.
- **Audit logs are evidence.** Audit entries are prefixed `[AUDIT]:` on stdout
  (not yet a separate stream). Ship them somewhere tamper-evident with a
  retention policy that matches your compliance needs.
- **Containers need mlock support.** To use `mlock` inside a container, use a
  storage driver that supports it (e.g. `overlay2`) and raise the container
  runtime's `LimitMEMLOCK`.

## Pitfalls

- **The host is the trust boundary.** Encryption at rest does not protect a key
  sitting in swap or a core dump. Skipping the memory-protection steps is the
  most consequential omission.
- **Misconfigured SPIRE silently weakens everything.** A wildcard selector or a
  privileged user running a component undoes attestation. The protections exist
  only if configured correctly.
- **Storage tampering is still possible.** SPIKE encrypts data at rest, but an
  attacker with write access to the store can corrupt or delete it. Restrict
  store access independent of encryption.
- **Don't treat this as one-time.** Cipher suites, key lengths, and defaults
  change across versions. Hardening is continuous; keep a frequent upgrade
  cadence. See [Upgrading SPIKE](/recipes/upgrading-spike/).

## Cross-Links

- [Deploying SPIKE](/recipes/deploying-spike/)
- [Backup and restore](/recipes/backup-and-restore/)
- [Break-the-glass disaster recovery](/recipes/break-the-glass-recovery/)
- Reference: [Production Setup guide](/operations/production/),
  [Security model](/architecture/security-model/)

## What's Next

When something still will not come up, work the symptoms:
[Troubleshooting](/recipes/troubleshooting/).
