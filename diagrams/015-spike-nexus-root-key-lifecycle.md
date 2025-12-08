![SPIKE](../assets/spike-banner-lg.png)

```mermaid
stateDiagram-v2
    [*] --> NotSet: Nexus starts

    NotSet --> Recovering: Request shards from Keepers
    Recovering --> Set: Threshold shards received
    Recovering --> Recovering: Insufficient shards (retry)

    NotSet --> Restoring: Operator submits shards
    Restoring --> Set: Threshold shards received
    Restoring --> Restoring: Waiting for more shards

    Set --> InUse: Encrypt/decrypt operations
    InUse --> Set: Operation complete

    Set --> [*]: Nexus shutdown/restart<br/>(key erased from memory)

    note right of NotSet
        Root key is zero-initialized.
        Nexus waits for initialization.
    end note

    note right of Set
        Root key in memory.
        All operations functional.
    end note

    note right of InUse
        Cipher operations use root key.
        Thread-safe access via mutex.
    end note
```

**States:**

1. **NotSet**: Root key not in memory (zero-initialized)
    * Nexus cannot encrypt/decrypt
    * Waits for initialization

2. **Recovering**: Nexus requests shards from Keepers
    * Retry with exponential backoff
    * Continues until the threshold is met

3. **Restoring**: Operator manually submits shards
    * Break-the-glass recovery procedure
    * Stateful accumulation of shards

4. **Set**: Root key in memory
    * All cryptographic operations are functional
    * Normal operation mode

5. **InUse**: Active encryption/decryption operation
    * Thread-safe access via mutex
    * Returns to Set after operation

**Transitions:**
- **NotSet → Recovering**: Automatic (startup)
* **NotSet → Restoring**: Manual (operator action)
* **Recovering → Set**: Automatic (the threshold is met)
* **Restoring → Set**: Manual (the threshold is met)
* **Set → InUse**: Automatic (on crypto operation)
* **InUse → Set**: Automatic (operation complete)
* **Set → Exit**: Only on shutdown/restart (process termination)

