![SPIKE](../assets/spike-banner-lg.png)

## SPIKE Nexus Root Key Restoration (`spike restore`)

Restore the root key in SPIKE Nexus from recovery shards.

**Flow Details:**
- Shards submitted one at a time (stateful accumulation)
- Global `shards` variable protected by mutex
- When the threshold is reached, immediate restoration
- Shards sent to all keepers immediately after restore
- All shards zeroed from memory after reconstruction

**Configuration:**
- `SPIKE_NEXUS_SHAMIR_THRESHOLD`: Number of shards needed (default: 2)
- `SPIKE_NEXUS_SHAMIR_SHARES`: Total shards generated (default: 3)

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