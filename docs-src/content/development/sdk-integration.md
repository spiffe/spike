+++
#    \\ SPIKE: Secure your secrets with SPIFFE. — https://spike.ist/
#  \\\\\ Copyright 2024-present SPIKE contributors.
# \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SDK Integration Guide"
weight = 4
sort_by = "weight"
+++

# SDK Integration Guide

This guide demonstrates how to integrate the **SPIKE SDK** into your
applications for secret management. It covers the complete workflow from
SPIRE registration to secret operations.

## Prerequisites

Before integrating the SPIKE SDK, ensure the following are in place:

### 1. SPIRE Registration

Your workload must be registered in SPIRE with a SPIFFE ID:

**Kubernetes example:**
```bash
spire-server entry create \
  -spiffeID spiffe://example.org/myapp \
  -parentID spiffe://example.org/spire/agent/k8s_psat/node1 \
  -selector k8s:ns:default \
  -selector k8s:pod-label:app:myapp
```

**Bare-metal example:**
```bash
spire-server entry create \
  -spiffeID spiffe://example.org/myapp \
  -parentID spiffe://example.org/spire/agent/unix/hostname \
  -selector unix:uid:1001
```

### 2. SPIKE Policy

A policy must grant your workload access to the secrets it needs:

```bash
spike policy create myapp-policy \
  --spiffe-id-pattern "spiffe://example\.org/myapp" \
  --path-pattern "tenants/myapp/.*" \
  --permissions read,write
```

### 3. SPIKE Nexus Running

Ensure SPIKE Nexus is running and accessible from your workload.

## Basic Integration

Here is a minimal example showing how to use the SPIKE SDK:

```go
package main

import (
    "fmt"

    spike "github.com/spiffe/spike-sdk-go/api"
)

func main() {
    // Create a new SPIKE API client
    // Uses the default Workload API Socket
    api, err := spike.New()
    if err != nil {
        fmt.Println("Error connecting to SPIKE Nexus:", err.Error())
        return
    }

    // Close the connection when done
    defer api.Close()

    // Store a secret
    path := "tenants/myapp/db/creds"
    err = api.PutSecret(path, map[string]string{
        "username": "dbuser",
        "password": "dbpass123",
    })
    if err != nil {
        fmt.Println("Error writing secret:", err.Error())
        return
    }

    // Retrieve the secret
    secret, err := api.GetSecret(path)
    if err != nil {
        fmt.Println("Error reading secret:", err.Error())
        return
    }

    fmt.Printf("Username: %s\n", secret.Data["username"])
}
```

## Deployment

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:latest
        env:
        - name: SPIFFE_ENDPOINT_SOCKET
          value: "unix:///run/spire/sockets/agent.sock"
        - name: SPIKE_NEXUS_URL
          value: "https://spike-nexus:8553"
        volumeMounts:
        - name: spire-agent-socket
          mountPath: /run/spire/sockets
          readOnly: true
      volumes:
      - name: spire-agent-socket
        hostPath:
          path: /run/spire/sockets
          type: Directory
```

**Key configuration:**
* Mount the SPIRE Agent socket
* Set `SPIFFE_ENDPOINT_SOCKET` environment variable
* Set `SPIKE_NEXUS_URL` to the Nexus service endpoint

### Bare-Metal Deployment

```bash
# 1. Ensure SPIRE Agent is running
systemctl status spire-agent

# 2. Set environment variables
export SPIFFE_ENDPOINT_SOCKET=unix:///tmp/spire-agent/public/api.sock
export SPIKE_NEXUS_URL=https://localhost:8553

# 3. Run your application
./myapp
```

## Integration Patterns

### Pattern 1: Initialization Secret Fetch

Fetch all required secrets at application startup:

```go
func main() {
    api, _ := spike.New()
    defer api.Close()

    // Fetch all required secrets at startup
    dbCreds, _ := api.GetSecret("tenants/myapp/db/creds")
    apiKey, _ := api.GetSecret("tenants/myapp/api/key")

    // Initialize services with secrets
    db := connectDB(dbCreds.Data["username"], dbCreds.Data["password"])
    client := initAPIClient(apiKey.Data["key"])

    // Run application
    serve(db, client)
}
```

### Pattern 2: On-Demand Secret Fetch

Fetch secrets when needed for specific operations:

```go
func handleRequest(req Request) Response {
    api, _ := spike.New()
    defer api.Close()

    // Fetch secret for this specific request
    secret, _ := api.GetSecret("tenants/myapp/api/key")

    // Use secret
    response := callExternalAPI(secret.Data["key"])

    return response
}
```

### Pattern 3: Cached Secrets with Refresh

Cache secrets and refresh them periodically:

```go
type SecretCache struct {
    api   *spike.API
    cache map[string]map[string]string
    mu    sync.RWMutex
}

func (c *SecretCache) Get(path string) map[string]string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.cache[path]
}

func (c *SecretCache) Refresh(path string) {
    secret, _ := c.api.GetSecret(path)

    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[path] = secret.Data
}

// Refresh every 5 minutes
func (c *SecretCache) StartRefresh(path string) {
    ticker := time.NewTicker(5 * time.Minute)
    go func() {
        for range ticker.C {
            c.Refresh(path)
        }
    }()
}
```

## Secret Versioning

SPIKE supports secret versioning. You can access previous versions:

```go
// Store initial secret (version 1)
api.PutSecret("path", map[string]string{"password": "old"})

// Update secret (version 2)
api.PutSecret("path", map[string]string{"password": "new"})

// Get current version
current, _ := api.GetSecret("path")

// Get specific version
opts := &spike.GetSecretOptions{Version: 1}
oldVersion, _ := api.GetSecretWithOptions("path", opts)
```

## Error Handling

### Common Errors and Solutions

**Workload not registered in SPIRE:**
```
Error: Failed to acquire SVID: no registration entry found
```
*Solution:* Register your workload in SPIRE with correct selectors.

**No policy granting access:**
```
Error: 403 Forbidden - Permission denied
```
*Solution:* Create a policy granting your workload access to the secret path.

**SPIKE Nexus unreachable:**
```
Error: connection refused
```
*Solution:* Verify SPIKE Nexus is running and check network connectivity.

**SPIRE Agent not running:**
```
Error: Failed to acquire SVID: connection refused
```
*Solution:* Start SPIRE Agent and verify the socket path.

## What the SDK Handles

The SPIKE SDK handles all the complexity of secure secret management:

* **SVID acquisition** from SPIRE Agent
* **mTLS setup** with automatic certificate rotation
* **API communication** with SPIKE Nexus
* **Error handling** and retries

Your application focuses on business logic, not secret management
infrastructure.

<p>&nbsp;</p>

----

{{ toc_development() }}

----

{{ toc_top() }}