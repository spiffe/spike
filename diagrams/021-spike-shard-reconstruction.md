![SPIKE](../assets/spike-banner-lg.png)

```mermaid
sequenceDiagram
    participant Collector as Shard Collector
    participant Shards as Collected Shards
    participant Converter as Format Converter
    participant Shamir as Shamir Secret<br/>Sharing (CIRCL)
    participant P256 as P256 Elliptic<br/>Curve
    participant RootKey as Root Key<br/>[32 bytes]

    Note over Collector: Collect shards until threshold met

    loop Collect shards
        Collector->>Shards: Append shard
        Shards->>Shards: Validate no duplicates
        Shards->>Shards: Check threshold
    end

    alt Threshold not met
        Shards-->>Collector: Continue collecting
    else Threshold met
        Shards-->>Collector: Ready to reconstruct

        Collector->>Converter: Convert ShamirShard to<br/>secretsharing.Share
        Note right of Converter: Extract ID and value<br/>Create CIRCL Share

        loop For each shard
            Converter->>Converter: Create Share{ID, Value}
        end

        Converter-->>Collector: []secretsharing.Share

        Collector->>Shamir: secretsharing.Recover(t, shares)
        Note right of Shamir: threshold t = ShamirThreshold - 1<br/>Lagrange interpolation

        Shamir->>Shamir: Validate sufficient shares
        Shamir->>Shamir: Reconstruct polynomial
        Shamir->>Shamir: Evaluate at x=0 to get secret

        Shamir-->>Collector: P256 scalar (secret)

        Collector->>P256: Marshal scalar to bytes
        P256-->>Collector: []byte

        Collector->>RootKey: Convert to [32]byte array
        RootKey-->>Collector: [32]byte root key

        Collector->>Collector: mem.ClearRawBytes(shards)
        Note right of Collector: Zero out all shards

        Collector-->>Collector: Root key reconstructed
    end
```
