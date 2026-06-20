+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Granting a Workload Access to Secrets"
weight = 6
sort_by = "weight"
+++

# Granting a Workload Access to Secrets

## Problem

You have a running application and a secret in SPIKE, and now you want the app
to read that secret. This is the recipe that connects the others end to end:
the workload needs a SPIFFE identity (from SPIRE), a SPIKE policy that
authorizes that identity, and a few lines of SDK code. Miss any one of the
three and the read fails, usually with a confusing error.

## TL;DR

Three links in the chain, in order:

```bash
# 1. SPIRE: give the workload an identity
spire-server entry create \
  -spiffeID spiffe://example.org/acme/web \
  -parentID spiffe://example.org/spire/agent/k8s_psat/node1 \
  -selector k8s:ns:default \
  -selector k8s:pod-label:app:web

# 2. SPIKE: authorize that identity on the path
spike policy create \
  --name "acme-web" \
  --spiffeid-pattern '^spiffe://example\.org/acme/web$' \
  --path-pattern     '^tenants/acme/db/.*$' \
  --permissions      'read'
```

```go
// 3. App: read the secret with the SDK
api, _ := spike.New()
defer api.Close()
secret, _ := api.GetSecret("tenants/acme/db/creds")
fmt.Println(secret.Data["password"])
```

## Workflow

1. **Register the workload in SPIRE.** The SPIFFE ID is what SPIKE matches a
   policy against; the selectors are how SPIRE decides which process gets that
   ID.

   ```bash
   # Kubernetes
   spire-server entry create \
     -spiffeID spiffe://example.org/acme/web \
     -parentID spiffe://example.org/spire/agent/k8s_psat/node1 \
     -selector k8s:ns:default \
     -selector k8s:pod-label:app:web

   # Bare-metal
   spire-server entry create \
     -spiffeID spiffe://example.org/acme/web \
     -parentID spiffe://example.org/spire/agent/unix/hostname \
     -selector unix:uid:1001
   ```

2. **Write the SPIKE policy** that grants this identity access to the path.
   Grant the least privilege the workload needs (here, `read` only):

   ```bash
   spike policy create \
     --name "acme-web" \
     --spiffeid-pattern '^spiffe://example\.org/acme/web$' \
     --path-pattern     '^tenants/acme/db/.*$' \
     --permissions      'read'
   ```

3. **Read from the workload using the SDK.** The SDK acquires the SVID from the
   SPIRE Agent, sets up mTLS, and talks to Nexus. Your code just asks for the
   path:

   ```go
   package main

   import (
       "fmt"

       spike "github.com/spiffe/spike-sdk-go/api"
   )

   func main() {
       api, err := spike.New() // uses the default Workload API socket
       if err != nil {
           fmt.Println("connect:", err)
           return
       }
       defer api.Close()

       secret, err := api.GetSecret("tenants/acme/db/creds")
       if err != nil {
           fmt.Println("read:", err)
           return
       }
       fmt.Println("password:", secret.Data["password"])
   }
   ```

4. **Wire the runtime.** The SDK needs to find the SPIRE Agent socket and
   Nexus:

   ```bash
   export SPIFFE_ENDPOINT_SOCKET=unix:///run/spire/sockets/agent.sock
   export SPIKE_NEXUS_API_URL=https://spike-nexus:8553
   ./web
   ```

## Tips

- **Match the identity exactly, then widen if needed.** Pin a single workload
  with an anchored pattern (`^spiffe://example\.org/acme/web$`). Use `.*` only
  when you deliberately want a whole family of SVIDs to share the policy.
- **Least privilege.** A reader only needs `read`. Add `write`/`list` only for
  workloads that store or enumerate secrets, and never hand a workload `super`.
- **The SDK handles the hard parts**: SVID acquisition, mTLS, certificate
  rotation, retries. Your app focuses on business logic, not transport.
- **In Kubernetes**, mount the SPIRE Agent socket into the pod and set
  `SPIFFE_ENDPOINT_SOCKET` and `SPIKE_NEXUS_API_URL`. See the SDK guide for a
  full Deployment manifest.

## Pitfalls

- **All three links are required.** A missing SPIRE entry, a missing policy, or
  an unset socket each break the chain independently:
  - `no registration entry found` -> the SPIRE entry is missing or its
    selectors do not match the pod/process.
  - `403 Forbidden` / permission denied -> the workload has an SVID but no
    policy authorizes it on that path.
  - `connection refused` on SVID acquisition -> the SPIRE Agent socket is wrong
    or the agent is down.
- **Policy patterns are regex.** `^spiffe://example\.org/acme/web$`, not
  `spiffe://example.org/acme/web*`. Escape the dots; anchor the ends. See
  [Writing access policies](/recipes/writing-access-policies/).
- **Paths are namespaces.** The policy path and the SDK path must agree, and
  neither starts with a slash: `tenants/acme/db/creds`.
- **`GetSecret` returns a map.** Read the field you want from `secret.Data`
  (e.g. `secret.Data["password"]`); the value is the whole key-value map stored
  at that path.

## Cross-Links

- [Storing and reading secrets](/recipes/storing-and-reading-secrets/)
- [Writing access policies](/recipes/writing-access-policies/)
- [Integrating the Go SDK](/recipes/go-sdk-integration/)
- Reference: [SDK Integration Guide](/development/sdk-integration/)

## What's Next

Skip storing secrets entirely and use SPIKE to encrypt your own data:
[Using SPIKE as an encryption service](/recipes/encryption-as-a-service/).
