![SPIKE](../assets/spike-banner-lg.png)

## Scenario 1: Keeper Unavailable During Distribution

```
Nexus tries to send shard → Keeper unreachable
↓
Log error, continue to next Keeper
↓
Wait for next interval (5 minutes)
↓
Retry all Keepers
```

**Impact:** Minimal. Other Keepers still receive shards.

## Scenario 2: Insufficient Keepers During Recovery

```
Nexus starts, needs 3 shards
↓
Only 2 Keepers are online
↓
Retrieve 2 shards, threshold not met
↓
Retry with exponential backoff
↓
Continue until 3rd Keeper comes online
↓
Retrieve 3rd shard, reconstruct root key
```

**Impact:** Nexus waits until sufficient Keepers online.

## Scenario 3: All Keepers Lost Shards (Restart)

```
All Keepers restart (shards in memory lost)
↓
Nexus periodic distribution sends new shards
↓
Keepers store shards in memory
↓
System recovers within 5 minutes
```

**Impact:** Temporary. Nexus continuously redistributes.
