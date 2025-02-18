+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Recovery Procedures"
weight = 3
sort_by = "weight"
+++




# SPIKE Recovery Procedures

SPIKE ensures that your secrets are secure and resilient, aiming for seamless
operations even in the most challenging situations. This document outlines the 
steps required for recovering SPIKE in various scenarios, ensuring
you have the right guidance to handle any eventuality.

## SPIKE Nexus Crash Recovery

**SPIKE** is designed to automatically recover **SPIKE Nexus** from crashes.
Here is how this happens:

1. **SPIKE Nexus** crashes.
2. New **SPIKE Nexus** instance starts.
3. **SPIKE Nexus** ask for shards from **SPIKE Keeper**s.
4. Once **SPIKE Nexus** gathers adequate shards, it recreates its **root key**
   and resumes normal operations.

[spiffe]: https://spiffe.io/

## SPIKE Keeper Crash Recovery

**SPIKE Keeper** recovery is automatic, and does not require any manual
intervention.

**SPIKE Nexus** regularly sends the shard that a **SPIKE Keeper** has to store.
So, if a **SPIKE Keeper** instance crashes, it will eventually receive its
shard.


## Complete System Recovery

In the unlikely case that both **SPIKE Nexus** and all **SPIKE Keeper** 
instances crash all together, the system may transition to a state where
it cannot automatically recover.

In that case manual intervention will be necessary.

Here are the steps to take for this scenario:

1. Before complete system failure:
  * Change the **SPIFFE ID** of **SPIKE Pilot** to recovery mode by 
    executing `./hack/spire-server-entry-recover-register.sh`
  * Run `spike recover`
  * Save the files generated in `~/.spike/recover` folder to a safe,
    encrypted, and password-protected medium.
  * Securely erase the ~/.spike/recover` folder.
  * Change the **SPIFFE ID** of **SPIKE Pilot** back using
   `./hack/spire-server-entry-su-register.sh` or delete the registration
   entry entirely for extra security.
   * You can create the entry back using 
     `./hack/spire-server-entry-su-register.sh` when you need to use 
     **SPIKE Pilot**.
2. During complete system failure:
  * Change the **SPIFFE ID** of **SPIKE Pilot** to restore mode:
    `./hack/spire-server-entry-restore-register.sh`
  * Execute `spike restore` and enter the shards you created in the
    previous step one by one. Each `spike restore` call accepts a 
    single shard.
  * When you provide enough shards, the system will restore itself:
    **SPIKE Nexus** will restore its root key, and it will also hydrate
    its peer **SPIKE Keeper** instances to protect itself against future
    crashes.
  * Change the **SPIFFE ID** of **SPIKE Pilot** back using
    `./hack/spire-server-entry-su-register.sh` or delete the registration
    entry entirely for extra security.
    * You can create the entry back using
      `./hack/spire-server-entry-su-register.sh` when you need to use
      **SPIKE Pilot**.

1. Both **SPIKE Nexus**, **SPIKE Keeper** are unavailable, or the system is
   in on other irrecoverable state.
2. Admin executes `spike recover`.
3. Admin provides their **password**.
4. The encrypted **root key** is fetched from the database and injected to
   the memory of **SPIKE Nexus**.
5. **SPIKE Nexus** syncs the **root key** with **SPIKE Keeper**.
6. The system resumes normal operation.

### Total System Reset

This procedure is for resetting **SPIKE** to its factory defaults.

The situation:

* Both **SPIKE Nexus** and all **SPIKE Keeper** instances have crashed, there
  is no way to fetch the root key from **SPIKE Keeper**(s).
* The system administrator has not used `spike recover` to create recovery 
  shards, or they have lost access to the recovery shards.
* Everyone have learned their lessons, and now it's time to reset the system
  and conduct an extensive "what went wrong / what should have been done" 
  analysis.

How to proceed:

* Delete `~/.spike` folder, which will also delete all the persisted secrets
  in the SQLite backing store.
* Delete **SPIRE Server** registration entries.
* Redeploy **SPIKE** using your preferred method.
  * You can check out `./hack/start.sh` to see a sample startup/deployment
    script.
* This is a complete system reset; you'll lose all data and all former
  configuration, including secret access policies.

----

{{ toc_operations() }}

----

{{ toc_top() }}
