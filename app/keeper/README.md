![SPIKE](../../assets/spike-banner-lg.png)

## SPIKE Keeper

**SPIKE Keeper** caches the **SPIKE Nexus**' root encryption key shards.
A shard is a cryptographically computed derivative of the root key. You
will need more than one shard to recover the root key. This way, if one shard
is compromised, it is not enough to recover the root key.

This is useful when you need to recover state without requiring manual
intervention if **SPIKE Nexus** crashes.

In cases where both **SPIKE Nexus** and **SPIKE Keeper** crash, an admin will
need to manually re-key the system. To reduce the possibility of this, multiple
**SPIKE Keeper** instances can be installed as a federated mesh for redundancy.

## How It Works

During system initialization, **SPIKE Bootstrap** generates the root key,
splits it into shards using Shamir's Secret Sharing, and distributes each
shard to a different **SPIKE Keeper** instance.

When **SPIKE Nexus** needs to recover (*e.g., after a restart*), it contacts
the **SPIKE Keeper** instances to retrieve enough shards to reconstruct the
root key.

## Security

All communication with **SPIKE Keeper** is secured using mTLS with SPIFFE
identities. Only **SPIKE Bootstrap** (*for shard contribution*) and **SPIKE
Nexus** (*for shard retrieval*) are authorized to communicate with **SPIKE
Keeper**. Requests from any other workload are rejected.
