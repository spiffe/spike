+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "System Overview"
weight = 1
sort_by = "weight"
+++


# SPIKE System Overview

This document provides an overview of **SPIKE**, a [**SPIFFE**][spiffe]-native 
Secrets Management solution. It is designed to ensure secure storage, recovery, 
and management of sensitive data with a focus on simplicity, reliability, 
and scalability for production environments.

[spiffe]: https://spiffe.io/ "SPIFFE"

## SPIKE Components

**SPIKE** (*Secure Production Identity for Key Encryption*) is a secrets 
management system build on top of a [**SPIFFE**][spiffe] (*Secure Production
Identity Framework for Everyone*) identity control plane, consisting of three
components:

* **SPIKE Nexus** (`./nexus`): The secrets store
* **SPIKE Pilot** (`./spike`): The CLI
* **SPIKE Keeper** (`./keeper`): The redundancy mechanism

The system provides high availability for secret storage with a manual recovery
mechanism in case of irrecoverable failure.

## Identity Control Plane

The following diagram shows how [**SVID**s][svid] are assigned to **SPIKE** 
components and other actors in the system. **SVID**s, or *SPIFFE-Verifiable
Identity Documents*, are [x.509 Digital Certificates][x509], that contain
a [**SPIFFE ID**][spiffeid] in their [SAN][san] (*Subject Alternative Name*)

The following diagram illustrates how **SPIFFE** identities are distributed 
across different **SPIKE** system components using **SPIRE** as the identity 
control plane.

{{imglink(
href="/assets/docs/spike-svid-assignment.jpg"
src="/assets/docs/spike-svid-assignment.jpg"
alt="Establishing the Identity Control Plane."
)}}

In a **SPIKE** deployment, **SPIRE** acts as the central authority that issues 
**SVID**s to different workloads:

* Applications who need to manage secret lifecycles stored in **SPIKE Nexus**.
* **SPIKE** Infrastructure components:
  * **SPIKE** Nexus
  * **SPIKE** Pilot
  * Multiple **SPIKE** Keeper instances.

Each component receives its own SVID, which serves as a 
**cryptographically-verifiable** identity document. These SVIDs allow the 
components to:

* Prove their identity to other services
* Establish secure, authenticated [**mTLS**][mtls] connections
* Access resources they're authorized to use
* Communicate securely with other components in the system

The dashed boxes represent distinct security and deployment boundaries.
**SPIRE** provides identity management capabilities that span across these 
trust boundaries. This architecture allows administrative operations to be 
performed on a hardened, secured **SPIRE Server** instance (*shown in the top 
yellow box*), while restricting direct access to sensitive operations 
(*like creating SPIRE Server registration entries*) from users and applications 
located in other trust boundaries.

> **Zero Trust FTW!**
> 
> The approached described here is a common pattern in **zero-trust** 
> architectures, where every service needs to have a strong, verifiable 
> identity regardless of its network location.
> 
> This approach is more secure than traditional methods like shared secrets or 
> network-based security, as each workload gets its own unique, short-lived 
> identity that can be automatically rotated and revoked if needed.

[mtls]: https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/ "What is mTLS"
[svid]: https://spiffe.io/docs/latest/spire-about/spire-concepts/#a-day-in-the-life-of-an-svid "A Day in the Life of an SVID"
[spiffeid]: https://spiffe.io/docs/latest/spiffe-about/spiffe-concepts/ "SPIFFE Concepts"
[x509]: https://en.wikipedia.org/wiki/X.509 "X.509"
[san]: https://en.wikipedia.org/wiki/Public_key_certificate "Public Key Certificate"

## SPIKE Component Interaction

The following diagram depicts how various **SPIKE** components interact with
each other:

{{imglink(
  href="/assets/docs/spike-secret-management.jpg"
  src="/assets/docs/spike-secret-management.jpg"
  alt="Secret Management in SPIKE."
)}}

At the top level, there's an Application that consumes secrets through an 
mTLS (mutual TLS) connection to SPIKE Nexus. The application will likely use
the [**SPIKE Developer SDK**][spike-sdk-go] to consume secrets without having 
to implement the underlying SPIFFE mTLS wiring.

[spike-sdk-go]: https://github.com/spiffe/spike-sdk-go "SPIKE Go SDK"

The secrets are created/managed through:

An administrative user interacting with **SPIKE Pilot** through a command line 
interface (*the CLI is the `spike` binary itself*).

Then, **SPIKE Pilot** communicates with **SPIKE Nexus** over **mTLS** to create 
secrets.

**SPIKE Nexus** is the central management point for secrets. It's our
secrets store.

At the bottom of the diagram, multiple **SPIKE Keeper**s connect to 
**SPIKE Nexus** via mTLS. Each **SPIKE Keeper** holds a single 
[**Shamir Secret Share**][shamir] (*shard*) of the **root key** that 
**SPIKE Nexus** maintains in memory. 

This design ensures that compromising any individual **SPIKE Keeper** cannot 
breach the system, as a single shard is insufficient to reconstruct the 
root key.

The system's security can be tuned by configuring both the total number of 
**SPIKE Keepers** and the **threshold **of required shards needed to reconstruct 
the root key. 

During system bootstrapping, **SPIKE Nexus** distributes these shards to the 
**SPIKE Keeper**s. If **SPIKE Nexus** crashes or restarts, it automatically 
recovers by requesting shards from a threshold number of healthy 
**SPIKE Keeper**s to reconstruct the **root key**. 

This mechanism provides automatic resiliency and redundancy without requiring 
manual intervention or "*unsealing*" operations that are common in other secret 
management solutions.

The system's security and availability can be tuned by configuring both the 
total number of **SPIKE Keeper**s and the **threshold** of required shards 
needed to reconstruct the root key. This flexibility allows implementors to 
balance their security requirements against operational needs--from basic 
redundancy to highly paranoid configurations requiring many **SPIKE Keeper**s 
to be healthy.

**Both the individual shards and the assembled root key are exclusively held in 
memory and NEVER persisted to disk**, forming a core aspect of SPIKE's security
model.

[shamir]: https://en.wikipedia.org/wiki/Shamir%27s_secret_sharing "Shamir's Secret Sharing"

The system uses mTLS (mutual TLS) throughout for secure communication between 
components, which ensures:

* All communications are encrypted
* Both sides of each connection authenticate each other
* The system maintains a high level of security for secret management

## SPIKE Nexus Root Key Sharding

The following diagram shows how the **SPIKE Nexus** *root key* is split into 
shards and then delivered to **SPIKE Keeper**s:

{{imglink(
  href="/assets/docs/spike-shamir-sharding.jpg"
  src="/assets/docs/spike-shamir-sharding.jpg"
  alt="Secret Nexus root key sharding."
)}}

The **SPIKE Nexus** has a **root key** that's essential for encrypting the 
backing store. This **root key** is split into [Shamir shards][shamir] based 
on a configurable number and threshold. There should be as many keepers as
the created shards.

The key advantage of using *Shamir sharding* specifically (*versus other forms 
of key splitting*) is that it's **mathematically secure**: The shards are 
created using polynomial interpolation, meaning:

* Each shard contains no meaningful information about the original key by itself
* You need a threshold number of shards to reconstruct the key
* The system can be configured to require any M of N shards to reconstruct the 
  root key (*e.g., any 2 of 3, or 3 of 5, etc.*)

This provides both security and fault tolerance: The system can continue 
operating even if some **SPIKE Keeper**s become temporarily unavailable, as 
long as the threshold number of shards remain accessible.

## SPIKE Nexus Init Flow

The following diagram depicts **SPIKE Nexus** initial bootstrapping flow:

{{imglink(
  href="/assets/docs/spike-nexus-bootstrapping.jpg"
  src="/assets/docs/spike-nexus-bootstrapping.jpg"
  alt="Secret Nexus bootstrapping."
)}}

When **SPIKE Nexus** is configured to use an **in-memory** backing store, we
don't need **SPIKE Keeper** because the database is in **SPIKE Nexus**'s memory
and there is nothing to recover if **SPIKE Nexus** crashes. This is a convenient
setup to use for **development** purposes.

When **SPIKE Nexus** is configured to use a persistent backing store (*like
SQLite*), however, then it will follow two paths.

* If **SPIKE Nexus** has bootstrapped before, then it has just crashed. Which 
  means, it has lost its **root key**. So, it will try to recover its 
  **root key** by requesting **shards** from the **SPIKE Keeper**s.
* If, otherwise, it's the first time **SPIKE Nexus** is bootstrapping, then
  we compute a secure **root key**, initialize the backing store with that
  root key. Split the root key into [**Shamir**][shamir] shards, and send those
  shards to relevant **SPIKE Keeper** instances.

Regardless of the above flow, there is an ongoing operation (*shown in the
bottom part of the diagram*) that runs as a separate **goroutine**.

* At regular intervals, if **SPIKE Nexus** happens to have a **root key**, it
  computes [**Shamir**][shamir] shards out of it, and dispatches these shards
  to the **SPIKE Keeper**s.

This flow establishes a secure boot process that handles both initial setup and 
subsequent startups. The system ensures the **root key** is either properly 
recovered from existing shards or securely generated and distributed when 
starting fresh.

There is one edge case though: When there is a total system crash, and **SPIKE
Keeper**s don't have any shards in their memory, then you'll need a manual 
recovery.

This event is highly unlikely, as deploying a sufficient number of **SPIKE
Keepers** with proper geographic distribution significantly reduces the
probability of them all crashing simultaneously. Since **SPIKE Keepers** are
designed to operate independently and without requiring intercommunication,
failures caused by systemic issues are minimized. By ensuring redundancy across
diverse geographic locations, even large-scale outages or localized failures are
highly improbable to impact all **SPIKE Keepers** at once.

That being said, unexpected failures can occur, and the disaster recovery
procedure for these situations is described in the next section.

## SPIKE "break-the-glass" Disaster Recovery

> **Need a Runbook**?
>
> The [**SPIKE Recovery Procedures**](/@/operations/recovery.md) page contains
> step-by-step instructions to follow during, before, and a disaster occurs.
>
> You will need to prepare **beforehand** so that you can recover the root
> key when the system fails to automatically recover it from **SPIKE Keeper**s.

The following diagram outline **SPIKE**'s manual disaster recovery procedure.
You can open the picture on a new tab for an enlarged version of it.

{{imglink(
  href="/assets/docs/spike-doomsday-recovery.jpg"
  src="/assets/docs/spike-doomsday-recovery.jpg"
  alt="SPIKE Manual disaster recovery flow."
)}}

### Preventive Backup 

> **Run `spik recover` as Soon as You Can**
> 
> You must back up the **root key** shards using `spike recover` **BEFORE** a 
> disaster strikes.
>
> This is like having a spare key stored in a safe place before you lose your 
> main keys. Without this proactive backup step, there would be nothing to 
> recover from in a catastrophic failure.

This operation need to be done **BEFORE** any disaster; ideally, shortly after 
deploying **SPIKE**.

Here is how the flow goes:

* The Operator runs `spike recover` using **SPIKE Pilot**.
* **SPIKE Pilot** saves the recovery shards on the home directory of the system/
* The Operator encrypts and stores these shards in a secure medium, and securely
  erases the copies generated as an output to `spike recover`. 

When later recovery is needed, the Operator will provide these shards to 
**SPIKE** to restore the system back to its working state.

### Disaster Recovery

When disaster strikes (*shown in the red box*):

* **SPIKE Nexus** and SPIKE Keepers have simultaneously crashed and restarted.
* **SPIKE Nexus** has **lost** its **root key**
* **SPIKE Keeper**s don't have enough shards
* Thus, automatic recovery is impossible and the system requires manual 
  recovery.

In that case, the Operator uses `spike restore` to provide the previously 
backed-up shards one at a time

* **SPIKE Pilot** forwards the entered shard to **SPIKE Nexus**
* System acknowledges and tracks progress of shard restoration returning
  the amount of shards received, and the number of shards remaining to restore
  the root key.

### System Restoration

Once enough shards are provided, **SPIKE Nexus** reconstructs the **root key**.

A separate goroutine redistributes shards to **SPIKE Keeper**s and the System 
returns to normal operation.

## SPIKE Components

Here is an overview of each **SPIKE** component:

### SPIKE Nexus

* **SPIKE Nexus** is the primary component responsible for secrets management.
* It creates and manages the root encryption key.
* It handles secret encryption and decryption.
* It syncs the **root key**'s [**Shamir Shards**][shamir] with **SPIKE 
  Keepers**s. These shards then can be used to recover **SPIKE Nexus** 
  upon a crash.
* It provides an **RESTful mTLS API** for *secret lifecycle management*, 
  *policy management*, *admin operations*, and *disaster recovery*.

### SPIKE Keeper

* It is designed to be **simple** and **reliable**.
* It does one thing, and does it well.
* Its **only** goal is to keep a [**Shamir Shard**][shamir] in **memory**.
* By design, it does not have any knowledge about its peer **SPIKE Keepers**, 
  nor **SPIKE Nexus**. It doesn't require any configuration to be brought up.
  This makes it simple to operate, replace, scale, replicate.
* It enables automatic recovery if **SPIKE Nexus** crashes.

Since **SPIKE Keeper** only contains a single shard, its compromise will not 
compromise the system. 

The more keepers you have, the more reliable and secure your **SPIKE**
deployment will be. We recommend `5` **SPIKE Keeper** instances with a 
shard-generation threshold of `3`, for production deployments.

[Check out **SPIKE Production Hardening Guide**][production] for more
details.

[production]: /@/operations/production.md

### SPIKE Pilot

* It is the CLI to the system (*i.e., the `spike` binary that you see
  in the examples*).
* It converts CLI commands to **RESTful mTLS API calls** to **SPIKE Nexus**.

**SPIKE Pilot** is the only management entry point to the system. 
Thus, deleting/disabling/removing **SPIKE Pilot** reduces the attack surface
of the system since admin operations will not be possible without
**SPIKE Pilot**.

Similarly, revoking the **SPIRE Server** registration of **SPIKE Pilot**'s
**SVID** (*once SPIKE Pilot is no longer needed*) will effectively block 
administrative access to the system, improving the overall security posture.

### Builtin SPIFFE IDs

**SPIKE Nexus** recognizes the following builtin SPIFFE IDS:

* `spiffe://$trustRoot/spike/pilot/role/superuser`: Super Admin.
* `spiffe://$trustRoot/spike/pilot/role/recover`: Recovery Admin
* `spiffe://$trustRoot/spike/pilot/role/restore`: Restore Admin

You can check out the [**Administrative Access section of SPIKE security
model](@/architecture/security-model.md#administrative-access) for more
information about these roles.

----

{{ toc_architecture() }}

----

{{ toc_top() }}
