+++
title = "System Overview"
weight = 1
sort_by = "weight"
+++

[![SPIKE](/assets/spike-banner.png)](/)

# SPIKE System Overview

## SPIKE Components

**SPIKE** has the following system components:

* **SPIKE Nexus** (`./nexus`): The secrets store
* **SPIKE Pilot** (`./spike`): The CLI
* **SPIKE Keeper** (`./keeper`): The redundancy mechanism

The system provides high availability for secret storage with a manual recovery
mechanism in case of irrecoverable failure.

```txt
TODO: add an image of how nexus, spike, and keeper interact.

TODO: add an image about the bootstrapping flow

TODO: add an image indicating how nexus seeds keepers

TODO: add an image indicating how nexus recovers from crash using keepers

TODO: add an image about the shard creation and distribution.
```

Here is an overview of each **SPIKE** component:

### SPIKE Nexus

* **SPIKE Nexus** is the primary component responsible for secrets management.
* It creates and manages the root encryption key.
* It handles secret encryption and decryption.
* It syncs Shamir Shards with **SPIKE Keepers**s. These shards then can be 
  used to recover **SPIKE Nexus** upon a crash.
* It provides an admin interface for key management.

### SPIKE Keeper

* It is designed to be simple and reliable.
* Its only goal is to keep Shamir Shard in memory.
* By design, it does not have any knowledge about its peer **SPIKE Keepers**, 
  nor **SPIKE Nexus**. It doesn't require any configuration to be brought up.
  This makes it simple to operate, replace, scale, replicate.
* It enables automatic recovery if **SPIKE Nexus** crashes.
* Since it only contains a single shard, its compromise will not compromise the
  system. The more keepers you have, the more reliable and secure your **SPIKE**
  deployment will be. We recommend `5` **SPIKE Keeper** instances with a 
  shard-generation threshold of `3`, for production deployments.

```txt
TODO: link to SPIKE production hardening guide for details.
```

### SPIKE Pilot

* It is the CLI to the system.
* It converts CLI commands to RESTful mTLS API calls to **SPIKE Nexus**.
* It is the only management entry point to the system.
* Deleting/disabling/removing **SPIKE Pilot** reduces the attack surface
  of the system since admin operations will not be possible without
  **SPIKE Pilot**.
* Similarly, revoking the **SPIRE Server** registration of **SPIKE Pilot**'s
  **SVID** will effectively block administrative access to the system,
  improving the overall security posture.

```txt
TODO: link SPIKE production hardening guide and also explain how these can
be done in the SPIKE production hardening guide.
```