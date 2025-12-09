![SPIKE](../assets/spike-banner-lg.png)

## SPIKE Nexus Recovery Flow (`spike recover`)

Generate recovery shards from the running SPIKE Nexus instance.

**Security Notes:**
* Only operators who can assign the SPIFFE ID 
  `spiffe://$trustRoot/spike/pilot/role/recover` to **SPIKE Pilot** can start
  the recovery procedure.
* Shards are deterministic (*the same root key produces the same shards*)
* Shards are zeroed from the memory after save
* Threshold scheme: need `ShamirThreshold` shards to reconstruct

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
    Note right of Shamir: threshold t = ShamirThreshold - 1<br/>total shares n = ShamirShares<br/>Example: t=1, n=3<br/>(need 2 of 3 shards)

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
