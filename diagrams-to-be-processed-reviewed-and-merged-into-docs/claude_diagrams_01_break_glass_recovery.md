# Break-the-Glass Recovery and Restore Procedures

## 1. Recovery Flow (spike recover)

Generate recovery shards from the running SPIKE Nexus instance.

```mermaid
sequenceDiagram
    participant Operator as SPIKE Pilot<br/>(spike recover)
    participant Nexus as SPIKE Nexus
    participant RootKey as Root Key<br/>(in-memory)
    participant Shamir as Shamir Secret<br/>Sharing
    participant Disk as ~/.spike/recover

    Note over Operator: Operator initiates<br/>recovery procedure

    Operator->>Nexus: POST /v1/operator/recover
    Note right of Operator: SPIFFE ID:<br/>spike/pilot/role/recover

    Nexus->>Nexus: Validate SPIFFE ID

    Nexus->>RootKey: Get root key
    RootKey-->>Nexus: [32]byte root key

    Nexus->>Shamir: computeShares(rootKey)
    Note right of Shamir: threshold t = ShamirThreshold - 1<br/>total shares n = ShamirShares<br/>Example: t=2, n=5<br/>(need 3 of 5 shards)

    Shamir->>Shamir: Generate deterministic shards
    Note right of Shamir: Uses P256 elliptic curve<br/>Deterministic from root key

    Shamir-->>Nexus: []ShamirShard (first t+1 shards)

    Nexus->>Nexus: Validate shards
    Note right of Nexus: Check for:<br/>- Duplicates<br/>- Zero values<br/>- Valid indices

    Nexus-->>Operator: RecoveryResponse<br/>{shards: []ShamirShard}

    Operator->>Disk: Save shards to files
    Note right of Disk: Files: shard-1, shard-2, ...<br/>Format: JSON with index and value

    Operator->>Operator: mem.ClearRawBytes(shards)
    Note right of Operator: Zero out sensitive data

    Note over Operator,Disk: Recovery shards safely stored.<br/>System can now be restored if Nexus fails.
```

**Key Files:**
- `app/spike/internal/cmd/operator/recover.go` - CLI command
- `app/nexus/internal/route/operator/recover.go` - HTTP handler
- `app/nexus/internal/initialization/recovery/recovery.go::NewPilotRecoveryShards()`
- `app/nexus/internal/initialization/recovery/shamir.go::computeShares()`

**Security Notes:**
- Only operators with SPIFFE ID `spiffe://$trustRoot/spike/pilot/role/recover` can
  recover
- Shards are deterministic (same root key produces same shards)
- Shards are zeroed from memory after save
- Threshold scheme: need ShamirThreshold shards to reconstruct

---

## 2. Restore Flow (spike restore)

Restore the root key in SPIKE Nexus from recovery shards.

```mermaid
sequenceDiagram
    participant Operator as SPIKE Pilot<br/>(spike restore)
    participant Disk as ~/.spike/recover
    participant Nexus as SPIKE Nexus
    participant ShardStore as Shard Store<br/>(in-memory)
    participant Shamir as Shamir Secret<br/>Sharing
    participant RootKey as Root Key<br/>(in-memory)
    participant State as Nexus State<br/>(SQLite)
    participant Keepers as SPIKE Keepers

    Note over Nexus: SPIKE Nexus is running<br/>but has no root key

    loop For each shard (until threshold met)
        Operator->>Disk: Read shard file
        Disk-->>Operator: ShamirShard

        Operator->>Nexus: POST /v1/operator/restore<br/>{shard: ShamirShard}
        Note right of Operator: SPIFFE ID:<br/>spike/pilot/role/restore

        Nexus->>Nexus: Validate SPIFFE ID

        Nexus->>ShardStore: Append shard to global slice
        Note right of ShardStore: Protected by mutex<br/>Global: shards []ShamirShard

        ShardStore->>ShardStore: Check duplicates

        alt Threshold not met
            Nexus-->>Operator: RestoreResponse<br/>{Restored: false,<br/>ShardsCollected: n,<br/>ShardsRemaining: t+1-n}
        else Threshold met
            Nexus->>Shamir: ComputeRootKeyFromShards(shards)
            Note right of Shamir: Use t+1 shards<br/>to reconstruct secret

            Shamir->>Shamir: secretsharing.Recover(t, shares)
            Shamir-->>Nexus: [32]byte root key

            Nexus->>RootKey: SetRootKey(rootKey)
            Note right of RootKey: Store in memory<br/>Protected by mutex

            Nexus->>State: Initialize(rootKey)
            Note right of State: Load encrypted data<br/>from SQLite database

            State->>State: Decrypt policies and secrets

            Nexus->>Keepers: SendShardsPeriodically()
            Note right of Keepers: Start background goroutine<br/>Send shards every 5 minutes

            Nexus-->>Operator: RestoreResponse<br/>{Restored: true,<br/>ShardsCollected: t+1,<br/>ShardsRemaining: 0}

            Nexus->>ShardStore: mem.ClearRawBytes(shards)
            Note right of ShardStore: Zero out all shards
        end
    end

    Note over Operator,Keepers: System fully restored and operational.<br/>Keepers receive shards for redundancy.
```

**Key Files:**
- `app/spike/internal/cmd/operator/restore.go` - CLI command
- `app/nexus/internal/route/operator/restore.go` - HTTP handler
- `app/nexus/internal/initialization/recovery/recovery.go::RestoreBackingStoreFromPilotShards()`
- `app/nexus/internal/initialization/recovery/root_key.go::ComputeRootKeyFromShards()`

**Flow Details:**
- Shards submitted one at a time (stateful accumulation)
- Global `shards` variable protected by mutex
- When threshold reached, immediate restoration
- Shards sent to all keepers immediately after restore
- All shards zeroed from memory after reconstruction

**Configuration:**
- `SPIKE_SHAMIR_THRESHOLD`: Number of shards needed (default: 3)
- `SPIKE_SHAMIR_SHARES`: Total shards generated (default: 5)
