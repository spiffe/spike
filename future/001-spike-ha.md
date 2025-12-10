# SPIKE High Availability Strategy

## Status: Draft

## Overview

This document describes high availability (HA) strategies for SPIKE deployments.
Because of SPIKE's architecture, achieving HA is straightforward: SPIKE Nexus
is essentially stateless (the root key is reconstructed from Keepers), and
SPIKE Keepers are HA by design through Shamir threshold-based redundancy.

## Architecture Summary

### SPIKE Nexus Statelessness

SPIKE Nexus does not hold persistent state in memory:

- **Root key**: Reconstructed at startup by collecting shards from Keepers
- **Secrets/Policies**: Read directly from the backing store (no in-memory
  cache per ADR-0024)
- **Consequence**: Any Nexus instance with access to:
  1. Enough Keeper shards (threshold)
  2. The backing store

  ...can serve requests. Multiple Nexus instances can run simultaneously.

### SPIKE Keeper Redundancy

SPIKE Keepers provide fault tolerance through Shamir's Secret Sharing:

- **Default configuration**: 3 shards, threshold of 2
- **Recovery**: As long as `threshold` Keepers are available, the root key
  can be reconstructed
- **Rehydration**: SPIKE Nexus periodically sends shards to Keepers, so a
  restarted Keeper is automatically rehydrated

This means one Keeper can be down indefinitely without affecting availability.

## HA Strategies by Environment

### Kubernetes Deployments

For Kubernetes deployments, use standard Kubernetes primitives:

#### Option 1: Multiple Nexus Replicas with Service Load Balancing (Recommended)

```yaml
# SPIKE Nexus StatefulSet or Deployment with replicas > 1
spec:
  replicas: 2  # or more
---
# Service with standard load balancing
kind: Service
metadata:
  name: spike-nexus
spec:
  type: ClusterIP
  selector:
    app: spike-nexus
  ports:
    - port: 8553
      targetPort: 8553
```

All replicas share:
- The same backing store (via PersistentVolumeClaim or external database)
- The same set of Keepers

No leader election is needed because Nexus instances are stateless.

#### Option 2: Lease-Based Leader Election (Not Recommended)

Kubernetes Lease objects can provide leader election:

- Reference: https://msalinas92.medium.com/deep-dive-into-kubernetes-leases-robust-leader-election-for-daemonsets-with-go-examples-f3b9a8858c49

However, this adds complexity without significant benefit since multiple
Nexus instances can serve requests simultaneously. Leader election is only
useful if you want exactly one active instance (active-passive), which
provides no advantage over active-active for SPIKE.

#### Keeper Configuration

Run Keepers as a StatefulSet with at least `SPIKE_NEXUS_SHAMIR_SHARES`
replicas (default: 3):

```yaml
kind: StatefulSet
metadata:
  name: spike-keeper
spec:
  replicas: 3  # Must match SPIKE_NEXUS_SHAMIR_SHARES
  # ...
```

### Non-Kubernetes (Bare Metal / VM) Deployments

For non-Kubernetes deployments, the strategy depends on your infrastructure.

#### Option 1: Load Balancer with Health Checks (Recommended)

Use any load balancer (HAProxy, nginx, cloud LB) in front of multiple
Nexus instances:

```
                    ┌─────────────────┐
                    │  Load Balancer  │
                    │  (HAProxy/nginx)│
                    └────────┬────────┘
                             │
           ┌─────────────────┼─────────────────┐
           │                 │                 │
           ▼                 ▼                 ▼
    ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
    │ Nexus #1    │   │ Nexus #2    │   │ Nexus #3    │
    └─────────────┘   └─────────────┘   └─────────────┘
           │                 │                 │
           └─────────────────┼─────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ Shared Backend  │
                    │ (PostgreSQL/S3) │
                    └─────────────────┘
```

**HAProxy example configuration:**

```haproxy
frontend spike_nexus
    bind *:8553 ssl crt /path/to/cert.pem
    default_backend spike_nexus_servers

backend spike_nexus_servers
    balance roundrobin
    option httpchk GET /health
    http-check expect status 200

    server nexus1 192.168.1.10:8553 check ssl verify required
    server nexus2 192.168.1.11:8553 check ssl verify required
    server nexus3 192.168.1.12:8553 check ssl verify required
```

**nginx example configuration:**

```nginx
upstream spike_nexus {
    least_conn;
    server 192.168.1.10:8553;
    server 192.168.1.11:8553;
    server 192.168.1.12:8553;
}

server {
    listen 8553 ssl;

    location / {
        proxy_pass https://spike_nexus;
        proxy_ssl_verify on;
        # ... additional SSL configuration
    }

    location /health {
        proxy_pass https://spike_nexus;
        proxy_connect_timeout 5s;
        proxy_read_timeout 5s;
    }
}
```

#### Option 2: DNS Round-Robin (Simple but Less Reliable)

Configure multiple A records for the Nexus hostname:

```
spike-nexus.example.com.  A  192.168.1.10
spike-nexus.example.com.  A  192.168.1.11
spike-nexus.example.com.  A  192.168.1.12
```

Limitations:
- No health checking (failed instances still receive traffic)
- DNS caching can cause sticky sessions
- No graceful failover

#### Option 3: Systemd with Watchdog (Single-Node HA)

For single-node deployments where you want automatic restart on failure:

```ini
# /etc/systemd/system/spike-nexus.service
[Unit]
Description=SPIKE Nexus
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nexus
Restart=always
RestartSec=5
WatchdogSec=30

[Install]
WantedBy=multi-user.target
```

This provides process-level HA but not node-level HA.

### Backing Store Considerations

The backing store must also be HA. SQLite is single-node only.

#### SQLite (Default) - Not Recommended for HA

- Single-node only
- Use for development/testing or single-instance deployments
- For HA, migrate to PostgreSQL or use shared storage

#### PostgreSQL - Recommended for Production HA

- Native replication and failover support
- Use with pgpool-II or Patroni for automatic failover
- Connection string: Set via `SPIKE_NEXUS_*` environment variables

#### S3-Compatible Storage

- Inherently HA (managed by cloud provider)
- Good for secrets that don't change frequently
- Higher latency than local database

### Keeper Deployment for Non-Kubernetes

Run Keepers on separate nodes for true fault tolerance:

```
Node 1: spike-keeper (port 8443)
Node 2: spike-keeper (port 8443)
Node 3: spike-keeper (port 8443)
```

Configure Nexus to know about all Keepers:

```bash
export SPIKE_NEXUS_KEEPER_PEERS="https://keeper1:8443,https://keeper2:8443,https://keeper3:8443"
export SPIKE_NEXUS_SHAMIR_SHARES="3"
export SPIKE_NEXUS_SHAMIR_THRESHOLD="2"
```

Use a process manager (systemd, supervisord) to ensure Keepers restart
automatically:

```ini
# /etc/systemd/system/spike-keeper.service
[Unit]
Description=SPIKE Keeper
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/keeper
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## What You Do NOT Need

Because of SPIKE's architecture, you do NOT need:

1. **Leader election** - Multiple Nexus instances can serve requests
   simultaneously (ADR-0024)
2. **Distributed locking** - The backing store handles concurrency
   (ADR-0023, ADR-0024)
3. **Cache invalidation** - No in-memory cache exists (ADR-0024)
4. **Consensus protocols** - State is in the backing store, not in Nexus
5. **Sticky sessions** - Any Nexus can handle any request

## Failure Scenarios

| Component Failed | Impact | Recovery |
|-----------------|--------|----------|
| 1 Keeper | None (if threshold met) | Automatic rehydration |
| All Keepers | Nexus cannot start | Restore from recovery shards |
| 1 Nexus | None (if multiple) | LB removes from pool |
| All Nexus | Service unavailable | Restart any Nexus instance |
| Backing store | Service unavailable | Restore backing store |

## Summary

SPIKE's HA model is simple:

1. **Run multiple Nexus instances** behind a load balancer
2. **Run enough Keepers** to meet the threshold requirement
3. **Use an HA-capable backing store** (PostgreSQL, S3, or shared storage)
4. **No special coordination needed** between Nexus instances

This "shared-nothing" architecture (per Nexus instance) combined with
a shared backing store provides straightforward horizontal scaling and
fault tolerance without the complexity of distributed consensus.

## References

- [ADR-0021: SPIKE Keeper as a Stateless Shard Holder]
- [ADR-0023: Decision Against Implementing Lock/Unlock Mechanism]
- [ADR-0024: Transition from In-Memory Cache to Direct Backend Storage for HA]
