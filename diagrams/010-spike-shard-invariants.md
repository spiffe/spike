![SPIKE](../assets/spike-banner-lg.png)

## Key Properties
* Deterministic shards for consistency
* Continuous redistribution for resilience
* Retry logic for fault tolerance
* mTLS with SPIFFE ID validation
* Memory-only storage (no disk persistence)
* Threshold scheme: need `t` of `n` shards

## Deterministic Shards

**Why Deterministic?**
* The same root key always produces the same shards
* Critical for Nexus restarts
* Keepers can verify shard consistency
* Enables idempotent operations

## Consistency

**Invariants:**
1. All Keepers receive the same shard index consistently
2. Shard indices are stable across Nexus restarts
3. Any threshold number of shards to reconstruct the same root key
4. Keepers can independently verify shard validity

## Flow

**Bidirectional Flow:**
1. **Nexus → Keeper**: Periodic distribution for redundancy
2. **Keeper → Nexus**: On-demand retrieval for recovery
