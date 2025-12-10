![SPIKE](../assets/spike-banner-lg.png)

## Root Key Sharding with Shamir Secret Sharing

SPIKE uses Shamir Secret Sharing to split the root key into multiple shards.
A threshold number of shards (t+1) can reconstruct the original key. This
provides redundancy and enables break-the-glass recovery.

```mermaid
sequenceDiagram
    participant Caller as Bootstrap/Nexus
    participant State as State Manager
    participant RootKey as Root Key<br/>[32 bytes]
    participant Deterministic as Deterministic<br/>Reader
    participant P256 as P256 Elliptic<br/>Curve
    participant Shamir as Shamir Secret<br/>Sharing (CIRCL)
    participant Verify as Verification<br/>Logic

    Note over Caller: Generate shards from root key

    Caller->>State: computeShares(rootKey)

    State->>RootKey: Get root key
    RootKey-->>State: [32]byte

    State->>P256: Create scalar from root key
    Note right of P256: scalar := group.P256.NewScalar()<br/>scalar.SetBytes(rootKey)

    P256-->>State: P256 scalar (secret)

    State->>Deterministic: Create deterministic reader
    Note right of Deterministic: Seed reader with root key<br/>Critical for consistency

    Deterministic-->>State: io.Reader

    State->>Shamir: shamir.New(reader, threshold, secret)
    Note right of Shamir: threshold = ShamirThreshold - 1<br/>Example: t=1 (need 2 shards)<br/>total shares = ShamirShares<br/>Example: n=3 (3 total shards)

    Shamir->>Shamir: Split secret into n shares
    Note right of Shamir: Polynomial interpolation<br/>in finite field<br/>Degree t polynomial

    Shamir-->>State: []secretsharing.Share

    State->>State: Convert to ShamirShard format
    Note right of State: Extract ID and value<br/>from CIRCL Share

    loop For each share
        State->>State: Create ShamirShard
        Note right of State: ID: share index (0-based)<br/>Value: [32]byte
    end

    State-->>Verify: []ShamirShard

    Note over Verify: Verify shards can reconstruct secret

    Verify->>Verify: Take first t+1 shards
    Note right of Verify: Example: 2 of 3 shards

    Verify->>Shamir: secretsharing.Recover(t, shares)
    Note right of Shamir: Lagrange interpolation<br/>Reconstruct polynomial

    Shamir-->>Verify: Recovered scalar

    Verify->>Verify: Marshal scalar to bytes

    Verify->>Verify: Compare with original root key
    Note right of Verify: Constant-time comparison

    alt Mismatch detected
        Verify->>Verify: log.FatalErr()
        Note over Verify: Fail-fast security<br/>Cannot proceed
    else Match confirmed
        Verify-->>Caller: []ShamirShard (verified)

        Note over Caller: Shards ready for distribution
    end
```

## Deterministic Shard Generation

```mermaid
graph TD
    A[Root Key: [32 bytes]] --> B[Deterministic Reader]
    B --> C[Seeded Random Number Generator]
    C --> D[Generate Random Coefficients]
    D --> E[Construct Polynomial]
    E --> F[Evaluate at n Points]
    F --> G[Shard 1: f(1)]
    F --> H[Shard 2: f(2)]
    F --> I[Shard 3: f(3)]
    F --> J[Shard 4: f(4)]
    F --> K[Shard 5: f(5)]

    style A fill:#ff6b6b
    style B fill:#4ecdc4
    style G fill:#95e1d3
    style H fill:#95e1d3
    style I fill:#95e1d3
    style J fill:#95e1d3
    style K fill:#95e1d3
```

**Why Deterministic?**


**Benefits**:
1. **Consistency**: The same root key always produces the same shards
2. **Crash recovery**: Can regenerate identical shards after restart
3. **Keeper synchronization**: Keepers receive the same shard index consistently
4. **Verification**: Can verify distributed shards match expected values

**Critical for**:
* SPIKE Nexus restarts (regenerate and redistribute the same shards)
* Keeper failures (send the same shard to a replacement Keeper)
* Audit and verification
  **Example Shards:**

## Example Shards

```
Root Key: [0x12, 0x34, 0x56, ..., 0xAB]  (32 bytes)

Shard 0:
  ID: 0
  Value: [0xA1, 0xB2, 0xC3, ..., 0xD4]

Shard 1:
  ID: 1
  Value: [0xE5, 0xF6, 0x07, ..., 0x18]

Shard 2:
  ID: 2
  Value: [0x29, 0x3A, 0x4B, ..., 0x5C]

... (n total shards)
```
