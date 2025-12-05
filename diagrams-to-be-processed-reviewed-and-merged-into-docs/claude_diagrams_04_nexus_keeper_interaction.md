# SPIKE Nexus and Keeper Interaction

## Overview

SPIKE Keepers store shards of the root key for redundancy and disaster
recovery. The interaction is bidirectional:
1. **Nexus → Keeper**: Periodic shard distribution
2. **Nexus ← Keeper**: Shard retrieval during recovery

---

## 1. Periodic Shard Distribution (Nexus → Keeper)

SPIKE Nexus periodically sends shards to all Keepers to ensure redundancy.

```mermaid
sequenceDiagram
    participant Nexus as SPIKE Nexus
    participant Timer as Timer Goroutine
    participant RootKey as Root Key<br/>(in-memory)
    participant Shamir as Shamir Secret<br/>Sharing
    participant MTLSClient as mTLS Client
    participant Keepers as SPIKE Keepers<br/>(all instances)

    Note over Nexus: SPIKE Nexus starts

    Nexus->>Timer: Start SendShardsPeriodically()
    Note right of Timer: Background goroutine<br/>Runs every 5 minutes

    loop Every SPIKE_RECOVERY_KEEPER_UPDATE_INTERVAL
        Timer->>RootKey: GetRootKey()

        alt Root key not available
            RootKey-->>Timer: empty key
            Note over Timer: Skip this iteration<br/>Wait for next interval
        else Root key available
            RootKey-->>Timer: [32]byte root key

            Timer->>Shamir: computeShares(rootKey)
            Note right of Shamir: Generate deterministic shards<br/>t = ShamirThreshold - 1<br/>n = ShamirShares<br/>Example: 5 shards, need 3

            Shamir->>Shamir: Create P256 scalar from key
            Shamir->>Shamir: Deterministic share generation
            Note right of Shamir: Same key = same shards<br/>Critical for consistency

            Shamir-->>Timer: []ShamirShard

            Timer->>Timer: Get keeper URLs from env
            Note right of Timer: env.KeepersVal()<br/>Comma-separated list

            par Send to each Keeper
                Timer->>MTLSClient: Create client with source
                Note right of MTLSClient: X509Source from SPIFFE<br/>Predicate: AllowKeeper

                MTLSClient->>Keepers: POST /v1/store/contribute<br/>{shard: ShamirShard}
                Note right of MTLSClient: SPIFFE ID: spike/nexus<br/>mTLS authentication

                Keepers->>Keepers: Validate peer SPIFFE ID
                Note right of Keepers: Must be spike/nexus<br/>or spike/bootstrap

                alt Valid peer
                    Keepers->>Keepers: state.SetShard(shard)
                    Note right of Keepers: Store in memory<br/>Protected by mutex<br/>Global: keeperShard variable

                    Keepers-->>MTLSClient: 200 OK

                    MTLSClient-->>Timer: Success
                else Invalid peer
                    Keepers-->>MTLSClient: 403 Forbidden

                    MTLSClient-->>Timer: Error
                    Note over Timer: Log error, continue<br/>Will retry next interval
                end
            end

            Timer->>Timer: mem.ClearRawBytes(shards)
            Note right of Timer: Zero out sensitive data

            Note over Timer: Wait for next interval<br/>(5 minutes by default)
        end
    end
```

**Key Files:**
- `app/nexus/internal/initialization/recovery/recovery.go::SendShardsPeriodically()`
- `app/nexus/internal/initialization/recovery/update.go::sendShardsToKeepers()`
- `app/nexus/internal/initialization/recovery/shamir.go::computeShares()`
- `app/keeper/internal/route/store/contribute.go::RouteContribute()`

**Flow Details:**
- Timer starts when Nexus initializes
- Runs in background goroutine
- Interval configurable via `SPIKE_RECOVERY_KEEPER_UPDATE_INTERVAL`
- Default: 5 minutes
- Continues even if some Keepers are unreachable
- Deterministic shards: same root key produces same shards
- Critical for handling Nexus restarts

**Keeper Storage:**
```go
var (
    keeperShard *ShamirShard
    keeperShardMu sync.RWMutex
)

func SetShard(shard *ShamirShard) {
    keeperShardMu.Lock()
    defer keeperShardMu.Unlock()
    keeperShard = shard
}
```

---

## 2. Shard Retrieval During Recovery (Nexus ← Keeper)

When SPIKE Nexus starts without a root key, it retrieves shards from Keepers.

```mermaid
sequenceDiagram
    participant Nexus as SPIKE Nexus
    participant Recovery as Recovery<br/>Logic
    participant MTLSClient as mTLS Client
    participant Keepers as SPIKE Keepers<br/>(all instances)
    participant ShardStore as Shard Collection<br/>(in-memory)
    participant Shamir as Shamir Secret<br/>Sharing
    participant RootKey as Root Key<br/>(in-memory)
    participant State as Nexus State

    Note over Nexus: SPIKE Nexus starts<br/>No root key in memory

    Nexus->>Recovery: InitializeBackingStoreFromKeepers()

    Recovery->>Recovery: Get keeper URLs from env
    Note right of Recovery: env.KeepersVal()<br/>["https://keeper1:8554",<br/>"https://keeper2:8554", ...]

    loop Until threshold shards collected
        loop For each keeper URL
            Recovery->>MTLSClient: Create client with source
            Note right of MTLSClient: X509Source from SPIFFE<br/>Predicate: AllowKeeper

            MTLSClient->>Keepers: POST /v1/store/shard<br/>ShardGetRequest{}
            Note right of MTLSClient: SPIFFE ID: spike/nexus<br/>mTLS authentication

            Keepers->>Keepers: Validate peer SPIFFE ID
            Note right of Keepers: Must be spike/nexus

            alt Valid peer and shard available
                Keepers->>Keepers: state.GetShard()
                Note right of Keepers: Retrieve from memory<br/>Protected by mutex

                Keepers-->>MTLSClient: 200 OK<br/>ShardGetResponse{shard: ShamirShard}

                MTLSClient-->>Recovery: ShamirShard

                Recovery->>ShardStore: Append shard to collection
                Note right of ShardStore: []ShamirShard<br/>Validate no duplicates

                Recovery->>Recovery: Check if threshold met
                Note right of Recovery: Need ShamirThreshold shards<br/>Example: 3 of 5

                alt Threshold met
                    Note over Recovery: Enough shards collected.<br/>Proceed to reconstruction.
                else Threshold not met
                    Note over Recovery: Continue querying Keepers
                end
            else Keeper doesn't have shard yet
                Keepers-->>MTLSClient: 404 Not Found

                MTLSClient-->>Recovery: No shard available
                Note over Recovery: Try next keeper
            else Connection failed
                MTLSClient-->>Recovery: Network error

                Recovery->>Recovery: retry.Forever() with backoff
                Note right of Recovery: Max delay: 5 seconds<br/>Exponential backoff
            end
        end

        alt Threshold still not met after all keepers
            Recovery->>Recovery: Wait and retry all keepers
            Note right of Recovery: Exponential backoff<br/>Continue forever until success
        end
    end

    Note over Recovery: Threshold shards collected

    Recovery->>Shamir: ComputeRootKeyFromShards(shards)

    Shamir->>Shamir: Convert to secretsharing.Share
    Shamir->>Shamir: secretsharing.Recover(t, shares)
    Note right of Shamir: t = ShamirThreshold - 1<br/>Reconstruct secret

    Shamir->>Shamir: Marshal scalar to [32]byte
    Shamir-->>Recovery: [32]byte root key

    Recovery->>RootKey: SetRootKey(rootKey)
    Note right of RootKey: Store in memory<br/>Protected by mutex

    Recovery->>State: Initialize(rootKey)
    Note right of State: Load and decrypt<br/>secrets and policies<br/>from SQLite

    State-->>Recovery: Initialized

    Recovery->>Recovery: mem.ClearRawBytes(shards)
    Note right of Recovery: Zero out all shards

    Recovery->>Recovery: Start SendShardsPeriodically()
    Note right of Recovery: Begin periodic distribution<br/>to Keepers

    Recovery-->>Nexus: Initialization complete

    Note over Nexus: SPIKE Nexus fully operational
```

**Key Files:**
- `app/nexus/internal/initialization/recovery/recovery.go::InitializeBackingStoreFromKeepers()`
- `app/nexus/internal/initialization/recovery/keeper.go::iterateKeepersAndInitializeState()`
- `app/nexus/internal/initialization/recovery/shard.go::shardGetResponse()`
- `app/keeper/internal/route/store/shard.go::RouteShard()`

**Flow Details:**
- Nexus queries Keepers on startup if no root key
- Uses retry logic with exponential backoff
- Continues forever until threshold met
- Max retry delay: 5 seconds
- Validates shards (no duplicates, valid indices)
- Immediately starts periodic distribution after recovery

---

## 3. Keeper Shard Management

### Shard Storage in Keeper

```go
// Global state in SPIKE Keeper
var (
    keeperShard   *ShamirShard
    keeperShardMu sync.RWMutex
)

func SetShard(shard *ShamirShard) {
    keeperShardMu.Lock()
    defer keeperShardMu.Unlock()
    keeperShard = shard
}

func GetShard() *ShamirShard {
    keeperShardMu.RLock()
    defer keeperShardMu.RUnlock()
    return keeperShard
}
```

**Characteristics:**
- Stored only in memory (not persisted)
- Protected by read-write mutex
- Single shard per Keeper
- Shard index corresponds to Keeper instance
- Lost on Keeper restart (recovered from Nexus)

---

## 4. Authorization and Security

### SPIFFE ID Validation

**Nexus → Keeper (Contribute):**
- Peer MUST be `spiffe://$trustRoot/spike/nexus` OR
- Peer MUST be `spiffe://$trustRoot/spike/bootstrap`
- Validated in `app/keeper/internal/route/store/contribute.go`

**Nexus ← Keeper (Shard Retrieval):**
- Peer MUST be `spiffe://$trustRoot/spike/nexus`
- Validated in `app/keeper/internal/route/store/shard.go`

### mTLS Configuration

```go
// Nexus creates client to Keeper
client := network.CreateMTLSClientWithPredicate(
    source,
    predicate.AllowKeeper,  // Only talk to Keepers
)

// Keeper server configuration
net.ServeWithPredicate(
    source,
    routeHandler,
    predicate.AllowKeeperPeer,  // Only allow Nexus/Bootstrap
    env.KeeperTLSPortVal(),
)
```

---

## 5. Failure Scenarios

### Scenario 1: Keeper Unavailable During Distribution

```
Nexus tries to send shard → Keeper unreachable
↓
Log error, continue to next Keeper
↓
Wait for next interval (5 minutes)
↓
Retry all Keepers
```

**Impact:** Minimal. Other Keepers still receive shards.

### Scenario 2: Insufficient Keepers During Recovery

```
Nexus starts, needs 3 shards
↓
Only 2 Keepers are online
↓
Retrieve 2 shards, threshold not met
↓
Retry with exponential backoff
↓
Continue until 3rd Keeper comes online
↓
Retrieve 3rd shard, reconstruct root key
```

**Impact:** Nexus waits until sufficient Keepers online.

### Scenario 3: All Keepers Lost Shards (Restart)

```
All Keepers restart (shards in memory lost)
↓
Nexus periodic distribution sends new shards
↓
Keepers store shards in memory
↓
System recovers within 5 minutes
```

**Impact:** Temporary. Nexus continuously redistributes.

---

## 6. Configuration

Environment variables:

**SPIKE Nexus:**
- `SPIKE_KEEPERS`: Comma-separated Keeper URLs
  - Example: `https://keeper1:8554,https://keeper2:8554`
- `SPIKE_RECOVERY_KEEPER_UPDATE_INTERVAL`: Shard distribution frequency
  - Default: `5m` (5 minutes)
  - Format: Go duration string (e.g., `10m`, `1h`)

**SPIKE Keeper:**
- `SPIKE_KEEPER_TLS_PORT`: mTLS listen port
  - Default: `8554`

---

## 7. Shard Distribution Guarantees

### Deterministic Shards

**Why Deterministic?**
- Same root key always produces same shards
- Critical for Nexus restarts
- Keepers can verify shard consistency
- Enables idempotent operations

**Implementation:**
```go
// Use root key as seed for deterministic reader
reader := deterministicReader(rootKey)

// Generate shares with deterministic randomness
ss := shamir.New(reader, threshold, secret)
shares := ss.Share(numShares)
```

### Consistency

**Invariants:**
1. All Keepers receive same shard index consistently
2. Shard indices are stable across Nexus restarts
3. Any threshold number of shards reconstruct same root key
4. Keepers can independently verify shard validity

---

## 8. Operational Visibility

### Logging

**Nexus logs:**
- Shard distribution start/complete
- Individual Keeper successes/failures
- Threshold checks during recovery
- Root key reconstruction success

**Keeper logs:**
- Shard receipt from Nexus/Bootstrap
- SPIFFE ID validation results
- Shard requests from Nexus

### Monitoring

**Key metrics to monitor:**
- Shard distribution interval
- Keeper availability
- Failed shard distributions
- Recovery attempts and duration
- Threshold status during startup

---

## Summary

**Bidirectional Flow:**
1. **Nexus → Keeper**: Periodic distribution for redundancy
2. **Keeper → Nexus**: On-demand retrieval for recovery

**Key Properties:**
- Deterministic shards for consistency
- Continuous redistribution for resilience
- Retry logic for fault tolerance
- mTLS with SPIFFE ID validation
- Memory-only storage (no disk persistence)
- Threshold scheme: need `t` of `n` shards
