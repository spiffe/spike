![SPIKE](../assets/spike-banner-lg.png)

```mermaid
sequenceDiagram
    participant Bootstrap as SPIKE Bootstrap
    participant Random as crypto/rand
    participant P256 as P256 Elliptic<br/>Curve
    participant RootKey as Root Key<br/>(in-memory)
    participant Shamir as Shamir Secret<br/>Sharing
    participant Keepers as SPIKE Keepers

    Note over Bootstrap: Initial system bootstrap

    Bootstrap->>Random: Read 32 cryptographically secure random bytes
    Note right of Random: crypto/rand.Read()<br/>OS entropy source

    Random-->>Bootstrap: [32]byte seed

    Bootstrap->>P256: Create scalar from seed
    Note right of P256: P256 elliptic curve<br/>Cloudflare CIRCL library<br/>group.P256.NewScalar()

    P256->>P256: scalar.SetBytes(seed)
    Note right of P256: Deterministic scalar<br/>from seed

    P256-->>Bootstrap: P256 scalar

    Bootstrap->>Bootstrap: Marshal scalar to bytes
    Note right of Bootstrap: scalar.MarshalBinary()<br/>Produces [32]byte

    Bootstrap->>RootKey: Store in memory
    Note right of RootKey: Global variable:<br/>var rootKey [32]byte

    Note over RootKey: Root key generated.<br/>Never touches disk.

    Bootstrap->>Shamir: Split into shards
    Note right of Shamir: See separate diagram<br/>for sharding details

    Shamir-->>Bootstrap: []ShamirShard

    Bootstrap->>Keepers: Distribute shards
    Note right of Keepers: Each Keeper receives<br/>one shard

    Note over Bootstrap,Keepers: Root key generation complete.<br/>System ready for operation.
```
