![SPIKE](../assets/spike-banner-lg.png)

```mermaid
sequenceDiagram
    participant Nexus as SPIKE Nexus
    participant Init as Initialization<br/>Logic
    participant Recovery as Recovery<br/>Module
    participant Keepers as SPIKE Keepers
    participant Shamir as Shamir Secret<br/>Sharing
    participant RootKey as Root Key<br/>(global variable)
    participant State as State Backend
    participant DB as SQLite Database

    Note over Nexus: SPIKE Nexus starts

    Nexus->>Init: Initialize()

    Init->>RootKey: HasRootKey()?

    alt Root key exists (after restore)
        RootKey-->>Init: true
        Note over Init: Skip recovery
    else Root key does not exist
        RootKey-->>Init: false

        Init->>Recovery: InitializeBackingStoreFromKeepers()

        Recovery->>Keepers: Request shards from all Keepers
        Note right of Keepers: POST /v1/store/shard<br/>to each Keeper

        loop Until threshold shards collected
            Keepers-->>Recovery: ShamirShard

            Recovery->>Recovery: Collect shards
            Note right of Recovery: Validate no duplicates<br/>Check threshold
        end

        Recovery->>Shamir: ComputeRootKeyFromShards(shards)
        Note right of Shamir: Reconstruct secret from<br/>threshold number of shards

        Shamir->>Shamir: secretsharing.Recover(t, shares)
        Note right of Shamir: t = ShamirThreshold - 1<br/>Recover P256 scalar

        Shamir->>Shamir: Marshal scalar to [32]byte

        Shamir-->>Recovery: [32]byte root key

        Recovery->>RootKey: SetRootKey(rootKey)
        Note right of RootKey: Store in global variable<br/>Protected by mutex

        RootKey-->>Recovery: Success

        Recovery->>Recovery: mem.ClearRawBytes(shards)
        Note right of Recovery: Zero out all shards<br/>from memory

        Recovery-->>Init: Root key initialized
    end

    Init->>State: Initialize(rootKey)
    Note right of State: Create AES-GCM cipher<br/>from root key

    State->>DB: Load encrypted policies and secrets
    Note right of DB: Read from SQLite

    DB-->>State: Encrypted data

    State->>State: Decrypt policies
    Note right of State: Cache in memory

    State-->>Init: State initialized

    Init-->>Nexus: Ready to serve requests

    Note over Nexus: SPIKE Nexus operational
```
