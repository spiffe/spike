+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Deploying SPIKE (Kubernetes and Bare-Metal)"
weight = 10
sort_by = "weight"
+++

# Deploying SPIKE (Kubernetes and Bare-Metal)

## Problem

SPIKE is not a single binary you start. It is a small constellation of
components on top of SPIRE: Nexus (the secret store), one or more Keepers (root-
key shard holders), Pilot (the CLI), and a one-shot Bootstrap step. Standing
them up in the wrong order, or skipping bootstrap, leaves Nexus stuck in keeper
recovery and never Ready. This recipe is the map of *what* you deploy and in
*what order*; the linked guides have the full command-by-command walkthrough.

## TL;DR

The order is always: **SPIRE -> Keepers + Nexus -> Bootstrap -> ready**.

```text
1. SPIRE server + agent (identity substrate)
2. SPIKE Keepers and Nexus (Nexus waits in keeper recovery)
3. SPIKE Bootstrap (seeds the keepers with root-key shards)  <-- the easy miss
4. Nexus reconstructs the root key and becomes Ready
```

On Kubernetes that is the SPIFFE Helm chart plus a bootstrap Job; on bare-metal
it is the SPIRE binaries plus the SPIKE `make` targets. See
[Quickstart](/getting-started/quickstart/) (Kubernetes) and
[SPIKE on Linux](/development/bare-metal/) (bare-metal) for the exact commands.

## Workflow

### Kubernetes

1. **Install SPIRE** via the hardened SPIFFE Helm chart. This gives every
   component (including `spike/bootstrap`) a SPIFFE identity.
2. **Deploy the SPIKE components** (Keepers, Nexus). Install **without**
   `--wait`: Nexus cannot become Ready until the keepers are seeded, so a
   `--wait` install hangs.
3. **Run the Bootstrap Job** with the `spike/bootstrap` SVID. It generates the
   root key, splits it into Shamir shares, and seeds the keepers. Give its
   ServiceAccount RBAC for the `spike-bootstrap-state` ConfigMap so re-runs are
   idempotent.
4. **Wait on the Nexus rollout.** Once the keepers hold `threshold` shares,
   Nexus reconstructs the root key, initializes its backend, and goes Ready.

### Bare-Metal

1. **Run SPIRE** server and agent, and register the SPIKE component identities
   (the `hack/bare-metal/entry` scripts do this).
2. **Start the Keepers and Nexus** from the SPIKE binaries / `make` targets.
3. **Bootstrap** with `make bootstrap` to generate and distribute the root-key
   shares.
4. **Use Pilot.** `spike` talks to Nexus over mTLS once Nexus is Ready.

## Tips

- **Pick the backend deliberately.** `memory` for a throwaway dev box (no
  keepers, no bootstrap), `sqlite` for persistent production, `lite` for
  encryption-only. The mode changes whether you even need bootstrap. See
  [Choosing a backend store](/recipes/choosing-a-backend-store/).
- **Shares and threshold are deployment-wide.** Decide `SHARES` and `THRESHOLD`
  before bootstrap; they determine how many keepers you run and how many can
  fail. Production guidance is 5 shares / threshold 3.
- **Use fully-qualified DNS in Kubernetes.** Point `SPIKE_NEXUS_API_URL` at the
  `*.svc.cluster.local` name; the short `service.namespace` form is NXDOMAIN in
  some setups.
- **Set all the trust roots.** Nexus checks identities against the configured
  trust roots (including `SPIKE_TRUST_ROOT_NEXUS`); a missing one silently
  rejects callers.

## Pitfalls

- **Bootstrap is not optional for `lite`/`sqlite`.** Both need keepers and a
  root key. Forgetting bootstrap is the number-one reason a fresh deployment
  never reaches Ready. See
  [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/).
- **`helm install --wait` deadlocks.** Nexus is not Ready until the keepers are
  seeded, which happens after install. Install without `--wait`, bootstrap,
  then wait on the Nexus rollout.
- **Image cache traps (local clusters).** With `imagePullPolicy: Never`, kind
  and minikube reuse a cached image for the *same* tag even after a fresh load.
  Use a unique tag (or force a re-pull) when iterating on images.
- **Order is not negotiable.** Keepers and Nexus before bootstrap; bootstrap
  before expecting Ready. Out of order, you chase phantom failures.

## Cross-Links

- [Choosing a backend store](/recipes/choosing-a-backend-store/)
- [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/)
- [Production hardening](/recipes/production-hardening/)
- [Troubleshooting](/recipes/troubleshooting/)
- Reference: [Quickstart](/getting-started/quickstart/),
  [SPIKE on Linux](/development/bare-metal/),
  [Configuration](/usage/configuration/)

## What's Next

Lock the deployment down before it carries real secrets:
[Production hardening](/recipes/production-hardening/).
