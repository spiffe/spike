![SPIKE](../../assets/spike-banner.png)

## SPIKE Nexus Automatic Recovery After Crash

```mermaid
sequenceDiagram
    Note over N,K: Whenever possible, retry<br>with exponential backoff

    participant N as SPIKE Nexus
    participant K as SPIKE Keeper
    alt not initialized
        Note over N: Generate root key<br>with strong entropy
        Note over N: Validate key format and strength

        N->>+K: Send root key 
        K-->>N: Acknowledge receipt

        Note over K: Verify key format before caching        
    else already initialized
        alt root key is empty
            N->>+K: Request root key 
            K-->>N: {root key}
            Note over N,K: Log if root key is still empty.
            Note over N,K: If root key is empty,<br>Manual admin intervention is required.
        end
    end

    loop Every 5mins (configurable)
        alt SPIKE Nexus not initialized
            Note over N,K: skip this iteration.
        end

        alt SPIKE Keep unreachable
            Note over N,K: skip this iteration.
        end

        alt when root key empty
            N->>+K: Fetch root key 
            K-->>N: {root key}

            Note right of N: Log, if root key is still empty.
            Note right of N: Skip the rest of the loop.
        else is root key in memory
            N->>+K: Send root key

            Note over K: Cache in Memory
        end
    end
```