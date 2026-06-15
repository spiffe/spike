+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Troubleshooting"
weight = 12
sort_by = "weight"
+++

# Troubleshooting

## Problem

Something is wrong and the error message points at a symptom, not a cause.
Nexus will not go Ready, a workload gets `403` even though you wrote a policy,
or bootstrap hangs. This recipe is organized **symptom first**: find the line
that matches what you see, then work the checklist under it.

## TL;DR

| Symptom | Most likely cause |
|---------|-------------------|
| Nexus never Ready / loops in keeper recovery | Keepers unseeded; bootstrap |
| `403 Forbidden` on a secret | Missing policy, or pattern is a glob not regex |
| Bootstrap hangs / never completes | A keeper unreachable, or verify can't reach Nexus |
| `no registration entry found` | SPIRE entry missing or selectors mismatch |
| `connection refused` (SVID) | SPIRE Agent down or wrong socket path |

## Nexus Never Becomes Ready (Stuck in Keeper Recovery)

This is the classic fresh-deployment failure. Nexus rebuilds its root key from
the keepers on startup and **waits forever** until it can; if the keepers are
empty, it loops and the readiness probe keeps restarting it.

Work down this list:

1. **Did bootstrap run?** On a fresh `lite`/`sqlite` deployment the keepers
   start empty. Until bootstrap seeds them, Nexus *cannot* recover. This is the
   number-one cause. Run it. See
   [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/).
2. **Did you `helm install --wait`?** That deadlocks: Nexus is not Ready until
   the keepers are seeded, which happens *after* install. Install without
   `--wait`, run bootstrap, then wait on the Nexus rollout.
3. **Are all keepers reachable?** Bootstrap needs every keeper in
   `SPIKE_NEXUS_KEEPER_PEERS` reachable, and Nexus needs `threshold` of them to
   reconstruct. Check the peer list and that each keeper is up.
4. **Is `SPIKE_NEXUS_API_URL` resolvable?** In Kubernetes use the
   fully-qualified `*.svc.cluster.local` name; the short `service.namespace`
   form can be NXDOMAIN.
5. **Are all trust roots set?** Nexus validates callers against the configured
   trust roots, including `SPIKE_TRUST_ROOT_NEXUS`. A missing root silently
   rejects otherwise-valid identities.

## Policy Created but Access Denied (403)

The workload has an identity but is not authorized for what it tried.

1. **Does a policy actually match?** Both the `spiffeid-pattern` **and** the
   `path-pattern` must match. A policy that "looks right" usually has one too
   narrow or unanchored.
2. **Are the patterns regex, not globs?** `^tenants/acme/.*$`, not
   `tenants/acme/*`. A glob-style `*` silently matches the wrong set or nothing.
   See [Writing access policies](/recipes/writing-access-policies/).
3. **Did you escape the dots?** `example\.org`, not `example.org` (an unescaped
   `.` matches any character, sometimes masking the real problem).
4. **Right permission?** `read` to read, `write` to create/update/delete,
   `list` to enumerate. A reader with only `list` still cannot `get`.
5. **Path is a namespace.** No leading slash; `tenants/acme/db/creds`, and the
   policy path and request path must agree.

## Bootstrap Hangs or Never Completes

1. **Keepers not all up.** Bootstrap requires exactly `SHARES` reachable
   keepers; a wrong or incomplete peer list stalls or fails it.
2. **Verify cannot reach Nexus.** The post-seed verification calls the Nexus
   API; if `SPIKE_NEXUS_API_URL` is wrong, verification fails even though the
   shares landed.
3. **Re-run did nothing.** That is idempotency working: the
   `spike-bootstrap-state` ConfigMap records completion. To deliberately re-key,
   set `SPIKE_BOOTSTRAP_FORCE=true` (this orphans data under the old key).

## SVID / SPIRE Errors

- **`no registration entry found`**: the SPIRE entry is missing, or its
  selectors do not match the pod/process. Create the entry and confirm the
  selectors. See
  [Granting a workload access](/recipes/granting-a-workload-access/).
- **`connection refused` acquiring an SVID**: the SPIRE Agent is down or
  `SPIFFE_ENDPOINT_SOCKET` points at the wrong socket. Start the agent and fix
  the path.

## Tips

- **Read the audit log carefully.** Audit entries (`[AUDIT]:`) can log
  "enter/exit success" for a route even when the request fell through to a
  fallback and returned an error. Cross-check the actual HTTP status, not just
  the audit line.
- **Local clusters cache images by tag.** With `imagePullPolicy: Never`, kind
  and minikube reuse the cached image for the *same* tag even after a fresh
  load. Use a unique tag (or force a re-pull) when an image change "isn't taking
  effect."
- **Isolate the layer.** Confirm SPIRE issues the SVID, *then* that SPIKE
  authorizes it, *then* that the data operation works. Most "SPIKE" failures are
  actually identity or policy failures one layer down.

## Pitfalls

- **Treating the symptom, not the cause.** A restarting Nexus pod is almost
  never a Nexus bug; it is usually unseeded keepers. Fix the cause (bootstrap),
  not the symptom (probe tuning).
- **Assuming a 403 is a storage problem.** It is authorization. Check the
  policy and its patterns before touching the secret store.
- **Forgetting `lite` needs keepers too.** "Encryption-only" does not mean
  "no bootstrap." `lite` still recovers a root key from keepers and shows the
  exact same stuck-in-recovery symptom if unseeded.

## Cross-Links

- [Bootstrapping a fresh SPIKE](/recipes/bootstrapping-spike/)
- [Where the root key lives](/recipes/root-key-keepers-recovery/)
- [Writing access policies](/recipes/writing-access-policies/)
- [Deploying SPIKE](/recipes/deploying-spike/)
- Reference: [Configuration](/usage/configuration/)

## What's Next

Once it is healthy, read secrets from your own code:
[Integrating the Go SDK](/recipes/go-sdk-integration/).
