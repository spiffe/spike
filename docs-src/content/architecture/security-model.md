+++
# //    \\ SPIKE: Secure your secrets with SPIFFE.
# //  \\\\\ Copyright 2024-present SPIKE contributors.
# // \\\\\\\ SPDX-License-Identifier: Apache-2.

title = "SPIKE Security Model"
weight = 2
sort_by = "weight"
+++

# SPIKE Security Model

Here is a brief introduction to **SPIKE**'s security model.

## Machine as the Trust Boundary

**SPIKE** components are intended to be used as the foundation for 
cloud native secrets management in a zero trust environment. **SPIKE**
supports Linux and the BSD family (*including macOS*). Windows is not currently 
supported, though some early prototyping is a work in progress. 

**SPIKE** (*with the help of SPIFFE and SPIRE*) adheres to the zero trust 
networking security model in which it is assumed that network communication is 
hostile or presumably fully compromised. That said, it is also assumed that 
the hardware on which ***SPIKE** components run, as well as its operators, 
is trustworthy.

If the hardware is considered as an attack surface, or insider threats are
part of the threat model, then careful considerations should be made around 
key components. The physical placement of **SPIRE Server**, **SPIKE Nexus**, 
and **SPIKE Keeper**instances, and the security of their relevant configuration 
parameters will be important.

## Authentication and Communication

* All inter-component communication is secured through [**SPIFFE** mTLS][spiffe].
* Components identify each other using their **SVID**s.
* Network-level security is provided by **SPIFFE** mTLS.

[spiffe]: https://spiffe.io/

## Trust Boundaries

**The primary trust boundary is at the machine level**. Once the machine is
compromised, hardening **SPIKE** components will provide diminishing returns.
In that regard, both physical, and OS-level security is important.

For example, when the machine is compromised, an attacker with sufficient
privileges can observe and control the memory of **SPIKE Nexus**, or
**SPIKE Keeper**; they can inject their counterfeit workloads; they can modify
**SPIRE** and create their own registration entries

It's also worth noticing that, since **SPIKE Keeper** backs ups the **root key**
in memory, if **SPIKE Keeper** is compromised, the machine can be considered
compromised.

For containerized deployments, both **SPIKE Nexus** and **SPIKE Keeper**
shall be hardened.

## Threat Model Exceptions

The following are **not** considered part of **SPIKE**'s threat model:

* Protecting against the control of the storage backend: Any storage backend
  is considered untrustworthy by **SPIKE**, so any data saved in the storage
  backend is encrypted at rest, and only **SPIKE Nexus** can decrypt it.
  An attacker can perform arbitrary operations against the storage backend,
  It is not **SPIKE**'s responsibility to protect the storage backend itself;
  **SPIKE** only ensures that an attacker accessing the storage backend cannot
  reveal the data stored there.
* Protecting against memory analysis of running system components: If an 
  attacker can inspect the memory state of any component, then they already have
  direct access to the machine (*which is our primary trust boundary*). If this
  happens, then the confidentiality of the data may be compromised. Preventing 
  memory analysis is a common system security best practice, and it is out 
  of scope for **SPIKE** to enforce such measures.
    * System administrators should implement the following security measures to 
      prevent memory analysis:
      * Set `/proc/sys/kernel/yama/ptrace_scope` to `2` or `3`:
        * Value `2` restricts `ptrace` to `root-only` access
        * Value `3` disables `ptrace` completely, offering maximum security
      * Make this setting permanent by adding `kernel.yama.ptrace_scope = 2` to
        `/etc/sysctl.d/10-ptrace.conf`
      * Consider using **SELinux** or **AppArmor** profiles to further restrict 
        process debugging capabilities
      * If running in a container, ensure the container runtime is configured to
        disable ptrace capabilities (*e.g., 
        using `--security-opt=no-new-privileges` in Docker*)
      * Regular audit of processes with `CAP_SYS_PTRACE` capability, as this can
        bypass ptrace restrictions
* Protecting against malicious code execution on the underlying host system.
  This is again the system administrator's responsibility. **SPIKE** cannot 
  protect against malicious code execution as that ability likely requires 
  administrative privileges, which should be avoided for **SPIKE** components 
  in the first place to prevent privilege escalation.
* Protecting against the underlying system's flaws. The systems shall be 
  up to date with respect to dependencies, properly secured, monitored, and 
  hardened.
* Protecting against ill intent of **SPIKE** super admins: **SPIKE** assumes 
  trust for super administrators. Any malicious actions performed by super 
  admins, such as abusing their elevated privileges, are considered out of 
  scope for **SPIKE**'s threat model. It is the organization's responsibility 
  to enforce proper checks, balances, and monitoring mechanisms for super 
  admin activities.
* Protecting against **SPIKE** administrators supplying vulnerable or malicious
  configuration data. This includes both intentional or unintentional
  misconfiguration---an administrator is supposed to know what they are doing.
  Any data provided as configuration values to **SPIKE** should be
  validated. Misconfiguration of **SPIKE**, or **SPIFFE** can result in the 
  compromise of the confidentiality or the integrity of the data stored.

## The Backing Store is Untrusted

Since the storage backend resides outside the trusted boundary, **SPIKE** 
treats it as untrusted and encrypts data before sending it. This ensures that 
even if a malicious attacker gains access to the storage backend, the data 
remains secure, as it can only be decrypted by **SPIKE Nexus**. 

Additionally, the storage backend serves as a durable, persistent layer, 
ensuring data availability across application crashes and server restarts.

Especially when using an external data store other than the default local
SQLite backing store, although SPIKE assumes the store is untrusted, 
still considering the following will be prudent:

* If possible, have SPIKE's backing store as an isolated database not shared
  by any other service to reduce the attack surface.
* If that's not possible and the backing store is a shared database with other 
  services, be aware of who else has access to it and manages it?
* Be cognizant about how SPIKE Nexus will authenticate to this database. 
* Make sure the database connection is secure with TLS or mTLS.

## Network Isolation of SPIKE Keepers

SPIKE Keepers do not have any communication pathway between each other, and this
is a decision by design. This significantly limits the possibility of lateral
movements as even when an attacker gains a foothold on a SPIKE Keeper instance,
they cannot laterally move to other SPIKE Keeper instances.

SPIKE Nexus and SPIKE Keepers establish a hub-spoke topology where SPIKE
Keepers (the spokes) can only communicate with SPIKE Nexus (the hub).

## SPIKE Keeper Shard Distribution and Disaster Recovery

**SPIKE** uses **SPIKE Keeper**s, which are apps responsible for storing 
[Shamir shards][shamir] of the **root key**. Both the **root key** and the
**shards** are always in memory and **never** persisted to disk. 

**SPIKE Nexus** can establish a SPIFFE-based mTLS connection to request a shard 
from a **SPIKE Keeper**, enabling the system to auto-recover itself.

The security model allows for different levels of redundancy and control:

* A typical setup could involve three **SPIKE Keeper** instances. No single 
  share can reconstruct the root key alone, ensuring security. However, multiple 
  shares can be combined to restore the system when needed.
* **SPIKE Nexus** often automatically recovers itself from crashes using 
  **SPIKE Keeper**s. However, for the unlikely case of a total system crash, 
  each administrator can hold one of these shares and use `spike restore` to
  restore the system back to normal. Since, a single shard cannot recreate 
  the root key, we are mitigating risk by distributing trust.
* For those less concerned with strict separation, an alternative approach 
  could involve storing both shares on a single thumb drive or distributing 
  two shares across separate thumb drives in different safes. This trade-off 
  balances security with recovery convenience.

Ultimately, the design offers flexibility, allowing organizations to choose 
their preferred level of security while considering the operational impact of 
disaster recovery.

## Key Management

* The system assumes a long-lived, well-guarded, initial **root key**.
    * The root key will be periodically rotated, but still, it will be 
      **long-lived**.
* The **root key** is automatically generated by **SPIKE Nexus**, and it's
  **never** stored on disk in plain text (*i.e., it always lives in memory*)
* An administrator with adequate privileges can use `spike recover` to save
  [Shamir Shards][shamir] in an encrypted medium out-of-band for future
  break-the-glass disaster recovery.
* Root key rotation will also re-encrypt the secrets.

[shamir]: https://en.wikipedia.org/wiki/Shamir%27s_secret_sharing "Shamir's Secret Sharing"

## Workload Access

Workloads can securely access their secrets and perform lifecycle operations 
(*e.g., create, delete, and modify secrets*) based on access policies defined 
by an administrator (*using the `spike policy` command*). These policies 
specify what a workload is allowed to do with the secrets managed by 
**SPIKE Nexus**.

* **Default Deny:** By default, access to **SPIKE Nexus** is prohibited. Only
  super administrators have full access by default.
* **Policy Enforcement:** Workloads require a valid, explicitly defined policy
  to perform any lifecycle operation on paths that contain secrets.
* **Controlled Operations:** The access policies strictly govern operations such 
  as creating, deleting, or modifying secrets.
* **Access Scoping:** Policies can define the scope and level of access (*e.g.,
  read-only or full access*) on specific secret paths for each workload.

This ensures that workloads only access or modify the secrets they are
explicitly permitted to, in accordance with their predefined policies.

## Administrative Access

Although **SPIKE** uses policy-based access to secrets and administrative
operations, **SPIKE Nexus** recognizes certain builtin SPIFFE IDs and assigns
them predefined roles:

* Administrative access is granted using special SPIFFE IDs:
  * `spiffe://$trustRoot/spike/pilot/role/superuser`: Super Admin. Can do 
    everything but recovery or restore operations.
  * `spiffe://$trustRoot/spike/pilot/role/recover`: Recovery user. Can **only**
     recover the root key shards to the local file system.
  * `spiffe://$trustRoot/spike/pilot/role/restore`: Restore user. Can **only**
    restore the root key by providing one shard at a time.

This gives us the flexibility to have separate users own distinct operational
responsibilities. For example, a specific operator may only restore the system 
upon an unexpected crash, but they may not have the right to define access 
policies for secrets.

This separation also provides better auditability.

* Once the system is initialized, accidental re-initialization is prevented.
  * For emergencies the admin user can use an out-of-band script to 
    "*factory-reset*" **SPIKE**.

## Multi-Admin Support

Other than the three predefined roles (*superuser, recover, restore*), named
admin access to the system would only be possible using an external identity
manager such as an OIDC provider.

**SPIKE** focuses on secure and efficient secret storage. It delegates access 
and identity management to established standards like OIDC, keeping 
authentication concerns out of scope.

----

{{ toc_architecture() }}

----

{{ toc_top() }}
