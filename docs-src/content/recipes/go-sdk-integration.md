+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.0

title = "Integrating the Go SDK"
weight = 13
sort_by = "weight"
+++

# Integrating the Go SDK

## Problem

You want your Go application to read (and maybe write) secrets from SPIKE
directly, without shelling out to the `spike` CLI. The SDK does this, and it
hides the hard parts: acquiring the SVID from the SPIRE Agent, setting up mTLS,
rotating certificates, and talking to Nexus. Your code just asks for a path.
The work is mostly making sure the *environment* around the code is right.

## TL;DR

```go
import spike "github.com/spiffe/spike-sdk-go/api"

api, err := spike.New()          // uses the default Workload API socket
if err != nil { /* handle */ }
defer api.Close()

secret, err := api.GetSecret("tenants/myapp/db/creds")
if err != nil { /* handle */ }
fmt.Println(secret.Data["password"])
```

The workload needs a SPIRE entry, a SPIKE policy granting access to the path,
and `SPIFFE_ENDPOINT_SOCKET` / `SPIKE_NEXUS_API_URL` set. See
[Granting a workload access](/recipes/granting-a-workload-access/) for the
identity and policy half.

## Workflow

1. **Add the dependency.**

   ```bash
   go get github.com/spiffe/spike-sdk-go/api
   ```

2. **Create the client and read a secret.** `spike.New()` connects via the
   default Workload API socket; `GetSecret` returns a struct whose `Data` is the
   key-value map stored at the path:

   ```go
   package main

   import (
       "fmt"

       spike "github.com/spiffe/spike-sdk-go/api"
   )

   func main() {
       api, err := spike.New()
       if err != nil {
           fmt.Println("connect:", err)
           return
       }
       defer api.Close()

       secret, err := api.GetSecret("tenants/myapp/db/creds")
       if err != nil {
           fmt.Println("read:", err)
           return
       }
       fmt.Println("user:", secret.Data["username"])
   }
   ```

3. **Write a secret** (if the workload's policy grants `write`):

   ```go
   err = api.PutSecret("tenants/myapp/db/creds", map[string]string{
       "username": "dbuser",
       "password": "s3cr3t",
   })
   ```

4. **Read a specific version** with options:

   ```go
   opts := &spike.GetSecretOptions{Version: 1}
   old, err := api.GetSecretWithOptions("tenants/myapp/db/creds", opts)
   ```

5. **Wire the runtime** so the SDK can find SPIRE and Nexus:

   ```bash
   export SPIFFE_ENDPOINT_SOCKET=unix:///run/spire/sockets/agent.sock
   export SPIKE_NEXUS_API_URL=https://spike-nexus:8553
   ```

## Tips

- **Pick a fetch pattern that fits the workload.**
  - *Startup fetch*: read all secrets once at boot. Simple; the app restarts
    to pick up changes.
  - *On-demand fetch*: read per request. Always fresh; more calls to Nexus.
  - *Cached with refresh*: cache and refresh on a ticker. Balances freshness
    and load; guard the cache with a mutex.
- **Reuse the client.** `spike.New()` sets up the SVID source and mTLS; create
  it once and reuse it, and `defer api.Close()`.
- **Versioning is built in.** Every `PutSecret` to a path creates a new version;
  read old ones with `GetSecretWithOptions`. See
  [Storing and reading secrets](/recipes/storing-and-reading-secrets/).
- **In Kubernetes**, mount the SPIRE Agent socket into the pod and set the two
  environment variables; the
  [SDK Integration Guide](/development/sdk-integration/) has a full Deployment
  manifest.

## Pitfalls

- **The error usually names the layer.** Map it before debugging SPIKE:
  - `no registration entry found` -> SPIRE entry / selectors.
  - `403 Forbidden` -> missing or mismatched SPIKE policy.
  - `connection refused` (SVID) -> SPIRE Agent down or wrong socket.
  - `connection refused` (Nexus) -> `SPIKE_NEXUS_API_URL` wrong or Nexus down.
- **`Data` is a map, read the field.** `GetSecret` returns the whole key-value
  map; pull `secret.Data["password"]`, not the struct itself.
- **Paths are namespaces.** `tenants/myapp/db/creds`, never with a leading
  slash, and identical to the path in the policy.
- **Don't log secrets.** It is easy to `fmt.Println(secret.Data)` while
  debugging and leave it in. The value is sensitive; keep it out of logs.

## Cross-Links

- [Granting a workload access to secrets](/recipes/granting-a-workload-access/)
- [Storing and reading secrets](/recipes/storing-and-reading-secrets/)
- [Writing access policies](/recipes/writing-access-policies/)
- Reference: [SDK Integration Guide](/development/sdk-integration/)

## What's Next

Keep your deployment current and patched:
[Upgrading SPIKE](/recipes/upgrading-spike/).
